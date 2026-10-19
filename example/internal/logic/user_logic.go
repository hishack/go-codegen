package logic

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"codegen-example/internal/model/gen/request"
	res "codegen-example/internal/model/gen/response"
	"codegen-example/internal/store"
	"codegen-example/internal/validate"

	"github.com/gin-gonic/gin"
)

// UserService 用户业务逻辑(伪实现:数据落内存)
type UserService struct {
	users *store.UserStore
}

// NewUserService 创建用户服务
func NewUserService() *UserService {
	return &UserService{users: store.NewUserStore()}
}

// maxAvatarSize 头像大小上限 5MB
const maxAvatarSize = 5 << 20

// CreateUser 新增用户
func (s *UserService) CreateUser(c *gin.Context, req request.CreateUserReq) (res.CreateUserRes, error) {
	if err := validate.Required("name", req.Name); err != nil {
		return res.CreateUserRes{}, err
	}
	if err := validate.RangeInt("age", req.Age, 0, 150); err != nil {
		return res.CreateUserRes{}, err
	}

	id := s.users.NextId()
	now := time.Now()
	s.users.Put(id, store.User{
		Id: id, Name: req.Name, Age: req.Age,
		Height: req.Height, Weight: req.Weight, Enabled: req.Enabled,
		Score: req.Score, Level: req.Level, Count: req.Count, Ratio: req.Ratio,
		Birthday: req.Birthday, Duration: req.Duration, Nickname: req.Nickname,
		CreatedAt: now,
	})
	return res.CreateUserRes{Id: id, CreatedAt: now}, nil
}

// GetUser 按 id 查询单个用户
func (s *UserService) GetUser(c *gin.Context, userId int64) (res.GetUserRes, error) {
	if err := validate.RequiredID("userId", userId); err != nil {
		return res.GetUserRes{}, err
	}
	u, ok := s.users.Get(userId)
	if !ok {
		return res.GetUserRes{}, validate.NotFoundErr("用户", userId)
	}
	return res.GetUserRes{Id: u.Id, Name: u.Name, Age: u.Age}, nil
}

// ListUsers 分页列出用户。伪实现里忽略分页参数,全量返回
func (s *UserService) ListUsers(c *gin.Context) (res.ListUsersRes, error) {
	all := s.users.Filter(nil)
	out := res.ListUsersRes{Total: int64(len(all)), List: make([]*res.ListUsersResListItem, 0, len(all))}
	for _, u := range all {
		out.List = append(out.List, &res.ListUsersResListItem{Id: u.Id, Name: u.Name, Age: u.Age})
	}
	return out, nil
}

// UpdateUser 整体更新用户(PUT 语义)
func (s *UserService) UpdateUser(c *gin.Context, userId int64, req request.UpdateUserReq) (bool, error) {
	if err := validate.RequiredID("userId", userId); err != nil {
		return false, err
	}
	if err := validate.Required("name", req.Name); err != nil {
		return false, err
	}
	if err := validate.RangeInt("age", req.Age, 0, 150); err != nil {
		return false, err
	}
	if _, ok := s.users.Get(userId); !ok {
		return false, validate.NotFoundErr("用户", userId)
	}

	_, ok := s.users.Update(userId, func(u store.User) store.User {
		u.Name = req.Name
		u.Age = req.Age
		return u
	})
	return ok, nil
}

// PatchUser 局部更新昵称(PATCH 语义:空值视为"不修改")
func (s *UserService) PatchUser(c *gin.Context, userId int64, req request.PatchUserReq) (bool, error) {
	if _, ok := s.users.Get(userId); !ok {
		return false, validate.NotFoundErr("用户", userId)
	}
	if strings.TrimSpace(req.Nickname) == "" {
		// 昵称为空则不改动,返回 true 表示"无需变更即成功"
		return true, nil
	}
	_, ok := s.users.Update(userId, func(u store.User) store.User {
		u.Nickname = req.Nickname
		return u
	})
	return ok, nil
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(c *gin.Context, userId int64) (bool, error) {
	if err := validate.RequiredID("userId", userId); err != nil {
		return false, err
	}
	if !s.users.Del(userId) {
		return false, validate.NotFoundErr("用户", userId)
	}
	return true, nil
}

// GetUserProfileVo 按 id 查询用户详情(含昵称与标签)
func (s *UserService) GetUserProfileVo(c *gin.Context, id int64) (res.GetUserProfileVoRes, error) {
	if err := validate.RequiredID("id", id); err != nil {
		return res.GetUserProfileVoRes{}, err
	}
	u, ok := s.users.Get(id)
	if !ok {
		return res.GetUserProfileVoRes{}, validate.NotFoundErr("用户", id)
	}
	tags := s.users.Tags(id)
	if tags == nil {
		tags = []string{} // 避免返回 null
	}
	return res.GetUserProfileVoRes{
		Id:      u.Id,
		Profile: &res.GetUserProfileVoResProfile{Nickname: u.Nickname, Tags: tags},
	}, nil
}

// UploadAvatar 上传头像与附件。伪实现只校验并返回虚拟地址,不落盘
func (s *UserService) UploadAvatar(c *gin.Context, req request.UploadAvatarReq) (string, error) {
	if err := validate.RequiredID("userId", req.UserId); err != nil {
		return "", err
	}
	if !s.users.Has(req.UserId) {
		return "", validate.NotFoundErr("用户", req.UserId)
	}
	if req.Avatar == nil {
		return "", validate.Wrap(validate.EmptyField, "缺少头像文件")
	}
	if req.Avatar.Size > maxAvatarSize {
		return "", validate.Wrap(validate.TooLarge,
			fmt.Sprintf("%d 字节,上限 %d", req.Avatar.Size, maxAvatarSize))
	}
	// 扩展名白名单
	if ext := strings.ToLower(filepath.Ext(req.Avatar.Filename)); ext != ".jpg" &&
		ext != ".jpeg" && ext != ".png" && ext != ".webp" {
		return "", validate.Wrap(validate.Invalid, "头像仅支持 jpg/png/webp")
	}
	for i, att := range req.Attachments {
		if att == nil {
			continue
		}
		if att.Size > maxAvatarSize {
			return "", validate.Wrap(validate.TooLarge,
				fmt.Sprintf("附件[%d] 过大", i))
		}
	}

	url := fmt.Sprintf("/static/avatar/%d%s", req.UserId,
		strings.ToLower(filepath.Ext(req.Avatar.Filename)))
	s.users.SetAvatar(req.UserId, url)
	return url, nil
}
