package logic

import (
	"strings"

	"codegen-example/internal/model/gen/request"
	res "codegen-example/internal/model/gen/response"
	"codegen-example/internal/store"
	"codegen-example/internal/validate"

	"github.com/gin-gonic/gin"
)

// ProfileService 用户资料业务逻辑(伪实现)
type ProfileService struct {
	profiles *store.ProfileStore
}

// NewProfileService 创建资料服务
func NewProfileService() *ProfileService {
	return &ProfileService{profiles: store.NewProfileStore()}
}

// currentProfileId 伪实现:固定取第一个用户作为"当前用户"。
// 真实项目应从会话/令牌中取当前登录用户 id。
func (s *ProfileService) currentProfileId() int64 {
	all := s.profiles.Filter(nil)
	if len(all) == 0 {
		return 0
	}
	return all[0].Id
}

// GetUserProfile 用户详情
func (s *ProfileService) GetUserProfile(c *gin.Context) (res.GetUserProfileRes, error) {
	id := s.currentProfileId()
	p, ok := s.profiles.Get(id)
	if !ok {
		return res.GetUserProfileRes{}, validate.NotFoundErr("用户资料", id)
	}

	tags := p.Tags
	if tags == nil {
		tags = []string{}
	}
	return res.GetUserProfileRes{
		Id: p.Id, Nickname: p.Nickname, Avatar: p.Avatar, Level: p.Level,
		CreatedAt: p.CreatedAt, Tags: tags,
		Profile: &res.GetUserProfileResProfile{Gender: p.Gender, Signature: p.Signature},
	}, nil
}

// UpdateUserProfile 更新用户资料。两个字段都为空时视为无意义请求
func (s *ProfileService) UpdateUserProfile(c *gin.Context, req request.UpdateUserProfileReq) (bool, error) {
	id := s.currentProfileId()
	if _, ok := s.profiles.Get(id); !ok {
		return false, validate.NotFoundErr("用户资料", id)
	}
	if strings.TrimSpace(req.Nickname) == "" && strings.TrimSpace(req.Signature) == "" {
		return false, validate.Wrap(validate.Invalid, "nickname 与 signature 不能同时为空")
	}
	if len([]rune(req.Nickname)) > 32 {
		return false, validate.Wrap(validate.Invalid, "nickname 最多 32 个字符")
	}
	if len([]rune(req.Signature)) > 140 {
		return false, validate.Wrap(validate.Invalid, "signature 最多 140 个字符")
	}

	_, ok := s.profiles.Update(id, func(p store.Profile) store.Profile {
		if strings.TrimSpace(req.Nickname) != "" {
			p.Nickname = req.Nickname
		}
		if strings.TrimSpace(req.Signature) != "" {
			p.Signature = req.Signature
		}
		return p
	})
	return ok, nil
}
