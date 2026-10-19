package service

import (
	"codegen-example/internal/model/gen/request"
	res "codegen-example/internal/model/gen/response"

	"github.com/gin-gonic/gin"
)

type INotifyService interface {
	SendSmsCode(c *gin.Context, req request.SendSmsCodeReq) (int, error)
	ListNotifies(c *gin.Context, req request.ListNotifiesReq) (res.ListNotifiesRes, error)
	ReadNotify(c *gin.Context, id int64) (bool, error)
	GetNotifyConfig(c *gin.Context) (res.GetNotifyConfigRes, error)
}
