package service

import (
	"codegen-example/internal/model/gen/request"
	res "codegen-example/internal/model/gen/response"

	"github.com/gin-gonic/gin"
)

type IUserService interface {
	CreateUser(c *gin.Context, req request.CreateUserReq) (res.CreateUserRes, error)
	GetUser(c *gin.Context, userId int64) (res.GetUserRes, error)
	ListUsers(c *gin.Context) (res.ListUsersRes, error)
	UpdateUser(c *gin.Context, userId int64, req request.UpdateUserReq) (bool, error)
	PatchUser(c *gin.Context, userId int64, req request.PatchUserReq) (bool, error)
	DeleteUser(c *gin.Context, userId int64) (bool, error)
	GetUserProfileVo(c *gin.Context, id int64) (res.GetUserProfileVoRes, error)
	UploadAvatar(c *gin.Context, req request.UploadAvatarReq) (string, error)
}
