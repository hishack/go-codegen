package logic

import (
	res "codegen-example/internal/model/gen/response"

	"github.com/gin-gonic/gin"
)

type NestedService struct{}

func NewNestedService() *NestedService {
	return &NestedService{}
}

// GetUserDetail 深度嵌套:res 会提升出多层命名结构体 UserRes -> UserResProfile -> UserResProfileAddress
func (s *NestedService) GetUserDetail(c *gin.Context, id int64) (res.GetUserDetailRes, error) {
	return res.GetUserDetailRes{}, nil
}

// GetOrderDetail 数组套对象套对象
func (s *NestedService) GetOrderDetail(c *gin.Context, orderId int64) (res.GetOrderDetailRes, error) {
	return res.GetOrderDetailRes{}, nil
}

// GetConfig map 类型
func (s *NestedService) GetConfig(c *gin.Context) (res.GetConfigRes, error) {
	return res.GetConfigRes{}, nil
}

// GetUserMap GetUserMap
func (s *NestedService) GetUserMap(c *gin.Context) (res.GetUserMapRes, error) {
	return res.GetUserMapRes{}, nil
}

// GetConfigA 同名字段但结构不同 -> 靠 ApiName 前缀区分,不会撞车
func (s *NestedService) GetConfigA(c *gin.Context) (res.GetConfigARes, error) {
	return res.GetConfigARes{}, nil
}

// GetConfigB GetConfigB
func (s *NestedService) GetConfigB(c *gin.Context) (res.GetConfigBRes, error) {
	return res.GetConfigBRes{}, nil
}
