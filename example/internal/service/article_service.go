package service

import (
	"codegen-example/internal/model/gen/request"
	res "codegen-example/internal/model/gen/response"

	"github.com/gin-gonic/gin"
)

type IArticleService interface {
	ListArticles(c *gin.Context, req request.ListArticlesReq) (res.ListArticlesRes, error)
	GetArticleById(c *gin.Context, id int64) (res.GetArticleByIdRes, error)
	PublishArticle(c *gin.Context, req request.PublishArticleReq) (int64, error)
	GetArticleStats(c *gin.Context) (res.GetArticleStatsRes, error)
}
