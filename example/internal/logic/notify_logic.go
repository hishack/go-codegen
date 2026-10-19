package logic

import (
	"fmt"
	"regexp"
	"sort"
	"time"

	"codegen-example/internal/model/gen/request"
	res "codegen-example/internal/model/gen/response"
	"codegen-example/internal/store"
	"codegen-example/internal/validate"

	"github.com/gin-gonic/gin"
)

// NotifyService 通知业务逻辑(伪实现:验证码存内存)
type NotifyService struct {
	notifies *store.NotifyStore
	codes    *store.SmsStore
}

// NewNotifyService 创建通知服务
func NewNotifyService() *NotifyService {
	return &NotifyService{
		notifies: store.NewNotifyStore(),
		codes:    store.NewSmsStore(),
	}
}

// phoneRe 简化的手机号校验:11 位、以 1 开头
var phoneRe = regexp.MustCompile(`^1[0-9]{10}$`)

// allowedScenes 允许的验证码场景
var allowedScenes = []string{"login", "register", "reset"}

// smsCodeTTL 验证码有效期
const smsCodeTTL = 5 * time.Minute

// SendSmsCode 发送短信验证码。同一手机号 1 分钟内只允许发一次
func (s *NotifyService) SendSmsCode(c *gin.Context, req request.SendSmsCodeReq) (int, error) {
	if err := validate.Required("phone", req.Phone); err != nil {
		return 0, err
	}
	if !phoneRe.MatchString(req.Phone) {
		return 0, validate.Wrap(validate.Invalid, "phone 格式不正确(需 11 位,以 1 开头)")
	}
	if err := validate.OneOf("scene", req.Scene, allowedScenes...); err != nil {
		return 0, err
	}
	if old, ok := s.codes.Get(req.Phone, req.Scene); ok {
		if time.Since(old.CreatedAt) < time.Minute {
			return 0, validate.Wrap(validate.Invalid, "发送过于频繁,请 1 分钟后再试")
		}
	}

	code := fmt.Sprintf("%06d", time.Now().UnixNano()%1000000)
	s.codes.Save(store.SmsCode{
		Phone: req.Phone, Scene: req.Scene, Code: code,
		ExpireAt: time.Now().Add(smsCodeTTL), CreatedAt: time.Now(),
	})
	return int(smsCodeTTL.Seconds()), nil
}

// ListNotifies 通知列表。userId 为 0 时返回全部;read=true 时只返回已读
func (s *NotifyService) ListNotifies(c *gin.Context, req request.ListNotifiesReq) (res.ListNotifiesRes, error) {
	if req.UserId < 0 {
		return res.ListNotifiesRes{}, validate.Wrap(validate.OutOfRange, "userId 不能为负数")
	}
	filtered := s.notifies.Filter(func(n store.Notify) bool {
		if req.UserId > 0 && n.UserId != req.UserId {
			return false
		}
		if req.Read {
			return n.Read
		}
		return true
	})

	sort.SliceStable(filtered, func(i, j int) bool {
		return filtered[i].CreateAt.After(filtered[j].CreateAt)
	})

	out := res.ListNotifiesRes{Total: int64(len(filtered)), List: make([]*res.ListNotifiesResListItem, 0, len(filtered))}
	for _, n := range filtered {
		out.List = append(out.List, &res.ListNotifiesResListItem{
			Id: n.Id, Title: n.Title, Content: n.Content,
			Read: n.Read, CreateAt: n.CreateAt,
		})
	}
	return out, nil
}

// ReadNotify 标记通知已读。重复标记视为成功(幂等)
func (s *NotifyService) ReadNotify(c *gin.Context, id int64) (bool, error) {
	if err := validate.RequiredID("id", id); err != nil {
		return false, err
	}
	if _, ok := s.notifies.Get(id); !ok {
		return false, validate.NotFoundErr("通知", id)
	}
	_, ok := s.notifies.Update(id, func(n store.Notify) store.Notify {
		n.Read = true
		return n
	})
	return ok, nil
}

// GetNotifyConfig 查询通知渠道配置
func (s *NotifyService) GetNotifyConfig(c *gin.Context) (res.GetNotifyConfigRes, error) {
	// map 不能为 nil,否则序列化成 null
	extras := map[string]string{
		"sms":    "13800000000",
		"wechat": "off",
		"email":  "off",
	}
	return res.GetNotifyConfigRes{
		Channels: []string{"sms", "wechat", "email"},
		Extras:   extras,
	}, nil
}
