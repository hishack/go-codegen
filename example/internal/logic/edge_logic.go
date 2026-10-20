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
	return nil
}

// Logout 只有 req,没有 res
func (s *EdgeService) Logout(c *gin.Context, req request.LogoutReq) error {
	return nil
}

// GetTime 只有 res,没有 req
func (s *EdgeService) GetTime(c *gin.Context) (res.GetTimeRes, error) {
	return res.GetTimeRes{}, nil
}

// Ping 空 req 对象
func (s *EdgeService) Ping(c *gin.Context) error {
	return nil
}

// SearchUser GET 带 req(query 参数场景)
func (s *EdgeService) SearchUser(c *gin.Context, req request.SearchUserReq) (res.SearchUserRes, error) {
	return res.SearchUserRes{}, nil
}

// GetUserOrder 多路径参数
func (s *EdgeService) GetUserOrder(c *gin.Context, userId int64, orderId int64) (res.GetUserOrderRes, error) {
	return res.GetUserOrderRes{}, nil
}

// GetArticle 无类型的路径参数,默认按 string
func (s *EdgeService) GetArticle(c *gin.Context, id string) (res.GetArticleRes, error) {
	return res.GetArticleRes{}, nil
}

// GetOrgUser 多级路径参数
func (s *EdgeService) GetOrgUser(c *gin.Context, orgId int64, userId int64) (res.GetOrgUserRes, error) {
	return res.GetOrgUserRes{}, nil
}
