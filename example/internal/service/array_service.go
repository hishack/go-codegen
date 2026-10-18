package service

import (
	"context"

	"codegen-example/internal/api"
	"codegen-example/internal/dal/vo"
)

type IArrayService interface {
	GetIntArray(ctx context.Context) (vo.GetIntArrayVo, error)
	GetStringArray(ctx context.Context) (vo.GetStringArrayVo, error)
	GetInt64Array(ctx context.Context) (vo.GetInt64ArrayVo, error)
	GetObjectArray(ctx context.Context) (vo.GetObjectArrayVo, error)
	GetOrderList(ctx context.Context) (vo.GetOrderListVo, error)
	CreateBatchUser(ctx context.Context, req *api.CreateBatchUserRequest) (int64, error)
}
