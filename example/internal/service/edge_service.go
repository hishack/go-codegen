package service

import (
	"context"

	"codegen-example/internal/api"
	"codegen-example/internal/dal/vo"
)

type IEdgeService interface {
	HealthCheck(ctx context.Context) error
	Logout(ctx context.Context, req *api.LogoutRequest) error
	GetTime(ctx context.Context) (vo.GetTimeVo, error)
	Ping(ctx context.Context) error
	SearchUser(ctx context.Context, req *api.SearchUserRequest) (vo.SearchUserVo, error)
	GetUserOrder(ctx context.Context, userId int64, orderId int64) (vo.GetUserOrderVo, error)
	GetArticle(ctx context.Context, id string) (vo.GetArticleVo, error)
	GetOrgUser(ctx context.Context, orgId int64, userId int64) (vo.GetOrgUserVo, error)
}
