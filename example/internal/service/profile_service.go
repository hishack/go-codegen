package service

import (
	"codegen-example/internal/model/gen/request"
	res "codegen-example/internal/model/gen/response"

	"github.com/gin-gonic/gin"
)

type IProfileService interface {
	GetUserProfile(c *gin.Context) (res.GetUserProfileRes, error)
	UpdateUserProfile(c *gin.Context, req request.UpdateUserProfileReq) (bool, error)
}
