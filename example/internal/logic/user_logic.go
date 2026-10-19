// user_logic.go 是生成器产出的骨架(接口 + TODO),
// 本文件在骨架之上补齐了一版可运行的伪实现:数据落在内存 store,
// 不接数据库,用于联调 handler/router 与验证生成结果。
//
// 注意:本文件是"手写代码",而生成器对 logic 层采取 PolicyIfSigDiff 策略——
// 只要方法签名没变,实现体就会原样保留;只有签名变了才会提示人工同步。
// 因此可以放心重新生成。
package logic

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"codegen-example/internal/model/gen/request"
	res "codegen-example/internal/model/gen/response"
	"codegen-example/internal/store"
	"codegen-example/pkg/errorutil"

	"github.com/gin-gonic/gin"
)

// 业务错误:统一走 errorutil,便于 handler 层转成统一响应
var (
	errUserNotFound = errorutil.BusinessError{Code: 40400, Message: "用户不存在"}
	errUserExists   = errorutil.BusinessError{Code: 40900, Message: "用户已存在"}
	errInvalidName  = errorutil.BusinessError{Code: 40001, Message: "用户名不能为空"}
	errInvalidAge   = errorutil.BusinessError{Code: 40002, Message: "年龄必须在 0-150 之间"}
	errFileTooLarge = errorutil.BusinessError{Code: 41300, Message: "文件超过大小限制"}
)

// maxAvatarSize 头像大小上限 5MB(伪实现,真实项目应改为配置或网关限制)
const maxAvatarSize = 5 << 20 // 5MB

// wrap 在保留原始错误描述的前提下追加上下文,例如
// "用户不存在: id=9999"。比直接 WithMessage 更适合排查问题——
// WithMessage 会整体覆盖描述,会把"用户不存在"这类关键信息冲掉。
func wrap(err errorutil.BusinessError, ctx string) errorutil.BusinessError {
	if ctx == "" {
		return err
	}
	return errorutil.BusinessError{Code: err.Code, Message: err.Message + ": " + ctx}
}

// UserService 用户业务逻辑
type UserService struct {
	users *store.UserStore
}

// NewUserService 创建用户服务,内部持有内存数据源
func NewUserService() *UserService {
	return &UserService{users: store.NewUserStore()}
}

// CreateUser 新增用户
func (s *UserService) CreateUser(c *gin.Context, req request.CreateUserReq) (res.CreateUserRes, error) {
	// 1. 参数校验:先做便宜的校验,避免污染数据源
	if err := validateName(req.Name); err != nil {
		return res.CreateUserRes{}, err
	}
	if err := validateAge(req.Age); err != nil {
		return res.CreateUserRes{}, err
	}

	// 2. 分配 id 并写入数据源
	id := s.users.NextId()
	s.users.Create(&store.User{
		Id: id, Name: req.Name, Age: req.Age,
		Height: req.Height, Weight: req.Weight, Enabled: req.Enabled,
		Score: req.Score, Level: req.Level, Count: req.Count, Ratio: req.Ratio,
		Birthday: req.Birthday, Duration: req.Duration, Nickname: req.Nickname,
		CreatedAt: time.Now(),
	})

	// 3. 返回创建结果
	return res.CreateUserRes{Id: id, CreatedAt: time.Now()}, nil
}

// GetUser 按 id 查询单个用户
func (s *UserService) GetUser(c *gin.Context, userId int64) (res.GetUserRes, error) {
	if userId <= 0 {
		return res.GetUserRes{}, errorutil.ParamsError.WithMessage("userId 非法")
	}
	u, ok := s.users.Get(userId)
	if !ok {
		return res.GetUserRes{}, wrap(errUserNotFound, fmt.Sprintf("id=%d", userId))
	}
	return res.GetUserRes{Id: u.Id, Name: u.Name, Age: u.Age}, nil
}

// ListUsers 分页列出用户。伪实现里忽略分页参数,全量返回
func (s *UserService) ListUsers(c *gin.Context) (res.ListUsersRes, error) {
	all := s.users.List()
	out := res.ListUsersRes{Total: int64(len(all))}
	out.List = make([]*res.ListUsersResListItem, 0, len(all))
	for _, u := range all {
		out.List = append(out.List, &res.ListUsersResListItem{
			Id: u.Id, Name: u.Name, Age: u.Age,
		})
	}
	return out, nil
}

// UpdateUser 整体更新用户(PUT 语义:name/age 全量覆盖)
func (s *UserService) UpdateUser(c *gin.Context, userId int64, req request.UpdateUserReq) (bool, error) {
	if err := validateName(req.Name); err != nil {
		return false, err
	}
	if err := validateAge(req.Age); err != nil {
		return false, err
	}

	u, ok := s.users.Get(userId)
	if !ok {
		return false, wrap(errUserNotFound, fmt.Sprintf("id=%d", userId))
	}

	// 只覆盖本次传入的字段,其余保持不变(PUT 也可以做全量覆盖,
	// 这里按 UpdateUserReq 实际拥有的字段处理)
	u.Name = req.Name
	u.Age = req.Age
	return s.users.Update(u), nil
}

// PatchUser 局部更新用户昵称(PATCH 语义:空值视为"不修改")
func (s *UserService) PatchUser(c *gin.Context, userId int64, req request.PatchUserReq) (bool, error) {
	if strings.TrimSpace(req.Nickname) == "" {
		// 昵称为空则不改动,返回 true 表示"无需变更即成功"
		if _, ok := s.users.Get(userId); !ok {
			return false, wrap(errUserNotFound, fmt.Sprintf("id=%d", userId))
		}
		return true, nil
	}

	ok, found := s.users.Patch(userId, func(u *store.User) {
		u.Nickname = req.Nickname
	})
	if !found {
		return false, wrap(errUserNotFound, fmt.Sprintf("id=%d", userId))
	}
	return ok, nil
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(c *gin.Context, userId int64) (bool, error) {
	if !s.users.Delete(userId) {
		return false, wrap(errUserNotFound, fmt.Sprintf("id=%d", userId))
	}
	return true, nil
}

// GetUserProfileVo 按 id 查询用户详情(含昵称与标签)
func (s *UserService) GetUserProfileVo(c *gin.Context, id int64) (res.GetUserProfileVoRes, error) {
	if id <= 0 {
		return res.GetUserProfileVoRes{}, errorutil.ParamsError.WithMessage("id 非法")
	}
	u, ok := s.users.Get(id)
	if !ok {
		return res.GetUserProfileVoRes{}, wrap(errUserNotFound, fmt.Sprintf("id=%d", id))
	}

	tags := s.users.Tags(id)
	if tags == nil {
		tags = []string{} // 避免返回 null
	}
	return res.GetUserProfileVoRes{
		Id: u.Id,
		Profile: &res.GetUserProfileVoResProfile{
			Nickname: u.Nickname,
			Tags:     tags,
		},
	}, nil
}

// UploadAvatar 上传头像与附件。伪实现只校验并返回虚拟地址,不落盘
func (s *UserService) UploadAvatar(c *gin.Context, req request.UploadAvatarReq) (string, error) {
	if req.UserId <= 0 {
		return "", errorutil.ParamsError.WithMessage("userId 非法")
	}
	if !s.users.Exists(req.UserId) {
		return "", wrap(errUserNotFound, fmt.Sprintf("id=%d", req.UserId))
	}
	if req.Avatar == nil {
		return "", errorutil.ParamsError.WithMessage("缺少头像文件")
	}

	// 头像大小校验
	if req.Avatar.Size > maxAvatarSize {
		return "", wrap(errFileTooLarge,
			fmt.Sprintf("%d 字节,上限 %d", req.Avatar.Size, maxAvatarSize))
	}
	// 扩展名白名单:真实项目这里应校验 MIME 与内容,防止伪造扩展名
	if ext := strings.ToLower(filepath.Ext(req.Avatar.Filename)); ext != ".jpg" &&
		ext != ".jpeg" && ext != ".png" && ext != ".webp" {
		return "", errorutil.ParamsError.WithMessage("头像仅支持 jpg/png/webp")
	}

	// 伪实现:读一下文件头,确保真的能读到内容,随后丢弃
	if _, err := req.Avatar.Open(); err != nil {
		return "", errorutil.ConvertError(err)
	}

	urls := make([]string, 0, len(req.Attachments)+1)
	url := fmt.Sprintf("/static/avatar/%d%s", req.UserId, filepath.Ext(req.Avatar.Filename))
	urls = append(urls, url)
	for i, att := range req.Attachments {
		if att == nil {
			continue
		}
		if att.Size > maxAvatarSize {
			return "", wrap(errFileTooLarge, fmt.Sprintf("附件[%d] 过大", i))
		}
		urls = append(urls, fmt.Sprintf("/static/attachment/%d/%d%s", req.UserId, i, filepath.Ext(att.Filename)))
	}

	s.users.SetAvatar(req.UserId, url)
	return url, nil
}

// ============ 参数校验 ============

func validateName(name string) error {
	if strings.TrimSpace(name) == "" {
		return errInvalidName
	}
	return nil
}

func validateAge(age int) error {
	if age < 0 || age > 150 {
		return errInvalidAge
	}
	return nil
}
