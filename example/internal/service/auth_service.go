package service

import (
	"codegen-example/internal/model/gen/request"
	res "codegen-example/internal/model/gen/response"

	"github.com/gin-gonic/gin"
)

type IAuthService interface {
	Login(c *gin.Context, req request.LoginReq) (res.LoginRes, error)
	RefreshToken(c *gin.Context, req request.RefreshTokenReq) (string, error)
	LogoutAccount(c *gin.Context, req request.LogoutAccountReq) error
}
