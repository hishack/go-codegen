package service

import (
	"codegen-example/internal/model/gen/request"
	res "codegen-example/internal/model/gen/response"

	"github.com/gin-gonic/gin"
)

type IArrayService interface {
	GetIntArray(c *gin.Context) (res.GetIntArrayRes, error)
	GetStringArray(c *gin.Context) (res.GetStringArrayRes, error)
	GetInt64Array(c *gin.Context) (res.GetInt64ArrayRes, error)
	GetObjectArray(c *gin.Context) (res.GetObjectArrayRes, error)
	GetOrderList(c *gin.Context) (res.GetOrderListRes, error)
	CreateBatchUser(c *gin.Context, req request.CreateBatchUserReq) (int64, error)
}
