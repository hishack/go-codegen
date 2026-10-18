package logic

import (
	"context"

	"codegen-example/internal/api"
	"codegen-example/internal/dal/vo"
)

type EdgeService struct{}

func NewEdgeService() *EdgeService {
	return &EdgeService{}
}

// HealthCheck 没有 req 也没有 res
func (s *EdgeService) HealthCheck(ctx context.Context) error {
	// TODO: 实现业务逻辑:没有 req 也没有 res
	return nil
}

// Logout 只有 req,没有 res
func (s *EdgeService) Logout(ctx context.Context, req *api.LogoutRequest) error {
	// TODO: 实现业务逻辑:只有 req,没有 res
	return nil
}

// GetTime 只有 res,没有 req
func (s *EdgeService) GetTime(ctx context.Context) (vo.GetTimeVo, error) {
	// TODO: 实现业务逻辑:只有 res,没有 req
	return vo.GetTimeVo{}, nil
}

// Ping 空 req 对象
func (s *EdgeService) Ping(ctx context.Context) error {
	// TODO: 实现业务逻辑:空 req 对象
	return nil
}

// SearchUser GET 带 req(query 参数场景)
func (s *EdgeService) SearchUser(ctx context.Context, req *api.SearchUserRequest) (vo.SearchUserVo, error) {
	// TODO: 实现业务逻辑:GET 带 req(query 参数场景)
	return vo.SearchUserVo{}, nil
}

// GetUserOrder 多路径参数
func (s *EdgeService) GetUserOrder(ctx context.Context, userId int64, orderId int64) (vo.GetUserOrderVo, error) {
	// TODO: 实现业务逻辑:多路径参数
	return vo.GetUserOrderVo{}, nil
}

// GetArticle 无类型的路径参数,默认按 string
func (s *EdgeService) GetArticle(ctx context.Context, id string) (vo.GetArticleVo, error) {
	// TODO: 实现业务逻辑:无类型的路径参数,默认按 string
	return vo.GetArticleVo{}, nil
}

// GetOrgUser 多级路径参数
func (s *EdgeService) GetOrgUser(ctx context.Context, orgId int64, userId int64) (vo.GetOrgUserVo, error) {
	// TODO: 实现业务逻辑:多级路径参数
	return vo.GetOrgUserVo{}, nil
}
