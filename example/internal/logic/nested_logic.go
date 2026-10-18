package logic

import (
	"context"

	"codegen-example/internal/dal/vo"
)

type NestedService struct{}

func NewNestedService() *NestedService {
	return &NestedService{}
}

// GetUserDetail 深度嵌套:res 会提升出多层命名结构体 UserRes -> UserResProfile -> UserResProfileAddress
func (s *NestedService) GetUserDetail(ctx context.Context, id int64) (vo.GetUserDetailVo, error) {
	// TODO: 实现业务逻辑:深度嵌套:res 会提升出多层命名结构体 UserRes -> UserResProfile -> UserResProfileAddress
	return vo.GetUserDetailVo{}, nil
}

// GetOrderDetail 数组套对象套对象
func (s *NestedService) GetOrderDetail(ctx context.Context, orderId int64) (vo.GetOrderDetailVo, error) {
	// TODO: 实现业务逻辑:数组套对象套对象
	return vo.GetOrderDetailVo{}, nil
}

// GetConfig map 类型
func (s *NestedService) GetConfig(ctx context.Context) (vo.GetConfigVo, error) {
	// TODO: 实现业务逻辑:map 类型
	return vo.GetConfigVo{}, nil
}

// GetUserMap GetUserMap
func (s *NestedService) GetUserMap(ctx context.Context) (vo.GetUserMapVo, error) {
	// TODO: 实现业务逻辑:GetUserMap
	return vo.GetUserMapVo{}, nil
}

// GetConfigA 同名字段但结构不同 -> 靠 ApiName 前缀区分,不会撞车
func (s *NestedService) GetConfigA(ctx context.Context) (vo.GetConfigAVo, error) {
	// TODO: 实现业务逻辑:同名字段但结构不同 -> 靠 ApiName 前缀区分,不会撞车
	return vo.GetConfigAVo{}, nil
}

// GetConfigB GetConfigB
func (s *NestedService) GetConfigB(ctx context.Context) (vo.GetConfigBVo, error) {
	// TODO: 实现业务逻辑:GetConfigB
	return vo.GetConfigBVo{}, nil
}
