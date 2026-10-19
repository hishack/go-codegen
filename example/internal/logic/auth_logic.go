package logic

import (
	"codegen-example/internal/model/gen/request"
	res "codegen-example/internal/model/gen/response"

	"github.com/gin-gonic/gin"
)

type AuthService struct{}

func NewAuthService() *AuthService {
	return &AuthService{}
}

// Login 本文件演示:一个 yaml 靠接口级 filename 拆成多个领域。 三个接口分属 user / profile / auth 三组文件,互不干扰。 登录
func (s *AuthService) Login(c *gin.Context, req request.LoginReq) (res.LoginRes, error) {
	// TODO: 实现业务逻辑:本文件演示:一个 yaml 靠接口级 filename 拆成多个领域。 三个接口分属 user / profile / auth 三组文件,互不干扰。 登录
	return res.LoginRes{}, nil
}

// RefreshToken 刷新令牌
func (s *AuthService) RefreshToken(c *gin.Context, req request.RefreshTokenReq) (string, error) {
	// TODO: 实现业务逻辑:刷新令牌
	return "", nil
}

// LogoutAccount 退出登录
func (s *AuthService) LogoutAccount(c *gin.Context, req request.LogoutAccountReq) error {
	// TODO: 实现业务逻辑:退出登录
	return nil
}
