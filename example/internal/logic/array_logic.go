package logic

import (
	"context"

	"codegen-example/internal/api"
	"codegen-example/internal/dal/vo"
)

type ArrayService struct{}

func NewArrayService() *ArrayService {
	return &ArrayService{}
}

// GetIntArray 数组的各种写法:基础类型数组、对象数组、混合嵌套
func (s *ArrayService) GetIntArray(ctx context.Context) (vo.GetIntArrayVo, error) {
	// TODO: 实现业务逻辑:数组的各种写法:基础类型数组、对象数组、混合嵌套
	return vo.GetIntArrayVo{}, nil
}

// GetStringArray GetStringArray
func (s *ArrayService) GetStringArray(ctx context.Context) (vo.GetStringArrayVo, error) {
	// TODO: 实现业务逻辑:GetStringArray
	return vo.GetStringArrayVo{}, nil
}

// GetInt64Array GetInt64Array
func (s *ArrayService) GetInt64Array(ctx context.Context) (vo.GetInt64ArrayVo, error) {
	// TODO: 实现业务逻辑:GetInt64Array
	return vo.GetInt64ArrayVo{}, nil
}

// GetObjectArray 对象数组:对应 res.list,会提升为 GetObjectArrayResListItem
func (s *ArrayService) GetObjectArray(ctx context.Context) (vo.GetObjectArrayVo, error) {
	// TODO: 实现业务逻辑:对象数组:对应 res.list,会提升为 GetObjectArrayResListItem
	return vo.GetObjectArrayVo{}, nil
}

// GetOrderList 多字段对象数组
func (s *ArrayService) GetOrderList(ctx context.Context) (vo.GetOrderListVo, error) {
	// TODO: 实现业务逻辑:多字段对象数组
	return vo.GetOrderListVo{}, nil
}

// CreateBatchUser 请求里带数组
func (s *ArrayService) CreateBatchUser(ctx context.Context, req *api.CreateBatchUserRequest) (int64, error) {
	// TODO: 实现业务逻辑:请求里带数组
	return 0, nil
}
