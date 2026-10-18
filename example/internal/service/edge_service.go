package service

import (
	"codegen-example/internal/model/gen/request"
	res "codegen-example/internal/model/gen/response"

	"github.com/gin-gonic/gin"
)

type IEdgeService interface {
	HealthCheck(c *gin.Context) error
	Logout(c *gin.Context, req request.LogoutReq) error
	GetTime(c *gin.Context) (res.GetTimeRes, error)
	Ping(c *gin.Context) error
	SearchUser(c *gin.Context, req request.SearchUserReq) (res.SearchUserRes, error)
	GetUserOrder(c *gin.Context, userId int64, orderId int64) (res.GetUserOrderRes, error)
	GetArticle(c *gin.Context, id string) (res.GetArticleRes, error)
	GetOrgUser(c *gin.Context, orgId int64, userId int64) (res.GetOrgUserRes, error)
}
