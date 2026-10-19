package logic

import (
	"codegen-example/internal/model/gen/request"
	res "codegen-example/internal/model/gen/response"
	"fmt"

	"github.com/gin-gonic/gin"
)

type ArticleService struct{}

func NewArticleService() *ArticleService {
	return &ArticleService{}
}

// ListArticles 文章列表
func (s *ArticleService) ListArticles(c *gin.Context, req request.ListArticlesReq) (res.ListArticlesRes, error) {
	// TODO: 实现业务逻辑:文章列表
	ctx := c.Request.Context()
	fmt.Println(ctx)
	return res.ListArticlesRes{}, nil
}

// GetArticleById 文章详情
func (s *ArticleService) GetArticleById(c *gin.Context, id int64) (res.GetArticleByIdRes, error) {
	// TODO: 实现业务逻辑:文章详情
	return res.GetArticleByIdRes{}, nil
}

// PublishArticle 发布文章
func (s *ArticleService) PublishArticle(c *gin.Context, req request.PublishArticleReq) (int64, error) {
	// TODO: 实现业务逻辑:发布文章
	return 0, nil
}

// GetArticleStats 统计文章数据
func (s *ArticleService) GetArticleStats(c *gin.Context) (res.GetArticleStatsRes, error) {
	// TODO: 实现业务逻辑:统计文章数据
	return res.GetArticleStatsRes{}, nil
}
