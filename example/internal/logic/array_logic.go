package logic

import (
	"codegen-example/internal/model/gen/request"
	res "codegen-example/internal/model/gen/response"
	"fmt"

	"github.com/gin-gonic/gin"
)

type ArrayService struct{}

func NewArrayService() *ArrayService {
	return &ArrayService{}
}

// GetIntArray 数组的各种写法:基础类型数组、对象数组、混合嵌套
func (s *ArrayService) GetIntArray(c *gin.Context) (res.GetIntArrayRes, error) {
	ctx := c.Request.Context()
	fmt.Print(ctx)
	return res.GetIntArrayRes{}, nil
}

// GetStringArray GetStringArray
func (s *ArrayService) GetStringArray(c *gin.Context) (res.GetStringArrayRes, error) {
	return res.GetStringArrayRes{}, nil
}

// GetInt64Array GetInt64Array
func (s *ArrayService) GetInt64Array(c *gin.Context) (res.GetInt64ArrayRes, error) {
	return res.GetInt64ArrayRes{}, nil
}

// GetObjectArray 对象数组:对应 res.list,会提升为 GetObjectArrayResListItem
func (s *ArrayService) GetObjectArray(c *gin.Context) (res.GetObjectArrayRes, error) {
	return res.GetObjectArrayRes{}, nil
}

// GetOrderList 多字段对象数组
func (s *ArrayService) GetOrderList(c *gin.Context) (res.GetOrderListRes, error) {
	return res.GetOrderListRes{}, nil
}

// CreateBatchUser 请求里带数组
func (s *ArrayService) CreateBatchUser(c *gin.Context, req request.CreateBatchUserReq) (int64, error) {
	return 0, nil
}
