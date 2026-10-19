package logic

import (
	"fmt"
	"strings"
	"time"

	"codegen-example/internal/model/gen/request"
	res "codegen-example/internal/model/gen/response"
	"codegen-example/internal/store"
	"codegen-example/internal/validate"

	"github.com/gin-gonic/gin"
)

// AuthService 认证业务逻辑(伪实现:令牌存内存)
type AuthService struct {
	tokens *store.TokenStore
	users  *store.UserStore
}

// NewAuthService 创建认证服务
func NewAuthService() *AuthService {
	return &AuthService{
		tokens: store.NewTokenStore(),
		users:  store.NewUserStore(),
	}
}

// tokenTTL 令牌有效期
const tokenTTL = 2 * time.Hour

// Login 登录。伪实现:用户名存在即视为凭证校验通过,签发令牌
func (s *AuthService) Login(c *gin.Context, req request.LoginReq) (res.LoginRes, error) {
	if err := validate.Required("username", req.Username); err != nil {
		return res.LoginRes{}, err
	}
	if err := validate.Required("password", req.Password); err != nil {
		return res.LoginRes{}, err
	}
	if len(req.Password) < 6 {
		return res.LoginRes{}, validate.Wrap(validate.Invalid, "password 至少 6 位")
	}

	found := false
	for _, u := range s.users.Filter(nil) {
		if u.Name == req.Username {
			found = true
			break
		}
	}
	if !found {
		return res.LoginRes{}, validate.Wrap(validate.Unauthorized, "用户名或密码错误")
	}

	token := fmt.Sprintf("tk-%s-%d", strings.TrimSpace(req.Username), time.Now().UnixNano())
	s.tokens.Save(store.Token{
		Value: token, Username: req.Username, ExpireAt: time.Now().Add(tokenTTL),
	})
	return res.LoginRes{Token: token, ExpireIn: int(tokenTTL.Seconds())}, nil
}

// RefreshToken 刷新令牌。伪实现:旧令牌未过期则轮换出新令牌
func (s *AuthService) RefreshToken(c *gin.Context, req request.RefreshTokenReq) (string, error) {
	if err := validate.Required("token", req.Token); err != nil {
		return "", err
	}
	old, ok := s.tokens.Get(req.Token)
	if !ok {
		return "", validate.Wrap(validate.Unauthorized, "令牌不存在或已失效")
	}
	if time.Now().After(old.ExpireAt) {
		s.tokens.Del(req.Token)
		return "", validate.Wrap(validate.Unauthorized, "令牌已过期")
	}

	s.tokens.Del(old.Value)
	newToken := fmt.Sprintf("tk-%s-%d", old.Username, time.Now().UnixNano())
	s.tokens.Save(store.Token{
		Value: newToken, Username: old.Username, ExpireAt: time.Now().Add(tokenTTL),
	})
	return newToken, nil
}

// LogoutAccount 退出登录
func (s *AuthService) LogoutAccount(c *gin.Context, req request.LogoutAccountReq) error {
	if err := validate.Required("token", req.Token); err != nil {
		return err
	}
	if !s.tokens.Del(req.Token) {
		return validate.Wrap(validate.Unauthorized, "令牌不存在或已失效")
	}
	return nil
}
