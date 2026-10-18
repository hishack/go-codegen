package service

import (
	"context"

	"codegen-example/internal/dal/vo"
)

type INestedService interface {
	GetUserDetail(ctx context.Context, id int64) (vo.GetUserDetailVo, error)
	GetOrderDetail(ctx context.Context, orderId int64) (vo.GetOrderDetailVo, error)
	GetConfig(ctx context.Context) (vo.GetConfigVo, error)
	GetUserMap(ctx context.Context) (vo.GetUserMapVo, error)
	GetConfigA(ctx context.Context) (vo.GetConfigAVo, error)
	GetConfigB(ctx context.Context) (vo.GetConfigBVo, error)
}
