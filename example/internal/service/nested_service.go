package service

import (
	res "codegen-example/internal/model/gen/response"

	"github.com/gin-gonic/gin"
)

type INestedService interface {
	GetUserDetail(c *gin.Context, id int64) (res.GetUserDetailRes, error)
	GetOrderDetail(c *gin.Context, orderId int64) (res.GetOrderDetailRes, error)
	GetConfig(c *gin.Context) (res.GetConfigRes, error)
	GetUserMap(c *gin.Context) (res.GetUserMapRes, error)
	GetConfigA(c *gin.Context) (res.GetConfigARes, error)
	GetConfigB(c *gin.Context) (res.GetConfigBRes, error)
}
