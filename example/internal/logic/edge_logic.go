package logic

import (
	"codegen-example/internal/model/gen/request"
	res "codegen-example/internal/model/gen/response"

	"github.com/gin-gonic/gin"
)

type EdgeService struct{}

func NewEdgeService() *EdgeService {
	return &EdgeService{}
}

// HealthCheck 没有 req 也没有 res
func (s *EdgeService) HealthCheck(c *gin.Context) error {
	// TODO: 实现业务逻辑:没有 req 也没有 res
	return nil
}

// Logout 只有 req,没有 res
func (s *EdgeService) Logout(c *gin.Context, req request.LogoutReq) error {
	// TODO: 实现业务逻辑:只有 req,没有 res
	return nil
}

// GetTime 只有 res,没有 req
func (s *EdgeService) GetTime(c *gin.Context) (res.GetTimeRes, error) {
	// TODO: 实现业务逻辑:只有 res,没有 req
	return res.GetTimeRes{}, nil
}

// Ping 空 req 对象
func (s *EdgeService) Ping(c *gin.Context) error {
	// TODO: 实现业务逻辑:空 req 对象
	return nil
}

// SearchUser GET 带 req(query 参数场景)
func (s *EdgeService) SearchUser(c *gin.Context, req request.SearchUserReq) (res.SearchUserRes, error) {
	// TODO: 实现业务逻辑:GET 带 req(query 参数场景)
	return res.SearchUserRes{}, nil
}

// GetUserOrder 多路径参数
func (s *EdgeService) GetUserOrder(c *gin.Context, userId int64, orderId int64) (res.GetUserOrderRes, error) {
	// TODO: 实现业务逻辑:多路径参数
	return res.GetUserOrderRes{}, nil
}

// GetArticle 无类型的路径参数,默认按 string
func (s *EdgeService) GetArticle(c *gin.Context, id string) (res.GetArticleRes, error) {
	// TODO: 实现业务逻辑:无类型的路径参数,默认按 string
	return res.GetArticleRes{}, nil
}

// GetOrgUser 多级路径参数
func (s *EdgeService) GetOrgUser(c *gin.Context, orgId int64, userId int64) (res.GetOrgUserRes, error) {
	// TODO: 实现业务逻辑:多级路径参数
	return res.GetOrgUserRes{}, nil
}
