package handler

import (
	"net/http"
	"strconv"

	"codegen-example/internal/model/gen/request"
	res "codegen-example/internal/model/gen/response"
	"codegen-example/internal/service"
	"codegen-example/pkg/errorutil"
	"codegen-example/pkg/response"

	"github.com/gin-gonic/gin"
)

type ArticleHandler struct {
	articleService service.IArticleService
}

func NewArticleHandler(userService service.IArticleService) *ArticleHandler {
	return &ArticleHandler{
		articleService: userService,
	}
}

// ListArticles 文章列表
// @Summary 文章列表
// @Description 文章列表
// @Tags cms
// @Accept json
// @Produce json
// @Param authorId query integer false "按作者过滤" format(int64)
// @Param tag query string false "按标签过滤"
// @Success 200 {object} res.ListArticlesRes "文章列表响应"
// @Router /api/v1/articles [get]
func (u *ArticleHandler) ListArticles(c *gin.Context) {
	var req request.ListArticlesReq
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse[any](errorutil.ParamsError.WithMessage(err.Error())))
		return
	}

	data, err := u.articleService.ListArticles(c, req)
	if err != nil {
		c.JSON(errorutil.HTTPStatusOf(err), response.NewErrorResponse[any](err))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse[res.ListArticlesRes](data))
}

// GetArticleById 文章详情。路径用 /cms-articles 以避开主测试集已占用的 /articles
// @Summary 文章详情。路径用 /cms-articles 以避开主测试集已占用的 /articles
// @Description 文章详情。路径用 /cms-articles 以避开主测试集已占用的 /articles
// @Tags cms
// @Accept json
// @Produce json
// @Param id path integer true "id" format(int64)
// @Success 200 {object} res.GetArticleByIdRes "文章详情。路径用 /cms-articles 以避开主测试集已占用的 /articles响应"
// @Router /api/v1/cms-articles/:id [get]
func (u *ArticleHandler) GetArticleById(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse[any](errorutil.ParamsError.WithMessage(err.Error())))
		return
	}

	data, err := u.articleService.GetArticleById(c, id)
	if err != nil {
		c.JSON(errorutil.HTTPStatusOf(err), response.NewErrorResponse[any](err))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse[res.GetArticleByIdRes](data))
}

// PublishArticle 发布文章
// @Summary 发布文章
// @Description 发布文章
// @Tags cms
// @Accept json
// @Produce json
// @Param req body request.PublishArticleReq true "发布文章请求"
// @Success 200 {object} int64 "发布文章响应"
// @Router /api/v1/articles [post]
func (u *ArticleHandler) PublishArticle(c *gin.Context) {
	var req request.PublishArticleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse[any](errorutil.ParamsError.WithMessage(err.Error())))
		return
	}

	data, err := u.articleService.PublishArticle(c, req)
	if err != nil {
		c.JSON(errorutil.HTTPStatusOf(err), response.NewErrorResponse[any](err))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse[int64](data))
}

// GetArticleStats 统计文章数据
// @Summary 统计文章数据
// @Description 统计文章数据
// @Tags cms
// @Accept json
// @Produce json
// @Success 200 {object} res.GetArticleStatsRes "统计文章数据响应"
// @Router /api/v1/articles/stats [get]
func (u *ArticleHandler) GetArticleStats(c *gin.Context) {
	data, err := u.articleService.GetArticleStats(c)
	if err != nil {
		c.JSON(errorutil.HTTPStatusOf(err), response.NewErrorResponse[any](err))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse[res.GetArticleStatsRes](data))
}

// GetArticleStats1 统计文章数据1
// @Summary 统计文章数据1
// @Description 统计文章数据1
// @Tags cms
// @Accept json
// @Produce json
// @Success 200 {object} res.GetArticleStats1Res "统计文章数据1响应"
// @Router /api/v1/articles/stats1 [get]
func (u *ArticleHandler) GetArticleStats1(c *gin.Context) {
	data, err := u.articleService.GetArticleStats1(c)
	if err != nil {
		c.JSON(errorutil.HTTPStatusOf(err), response.NewErrorResponse[any](err))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse[res.GetArticleStats1Res](data))
}

// GetArticleStats2 统计文章数据2
// @Summary 统计文章数据2
// @Description 统计文章数据2
// @Tags cms
// @Accept json
// @Produce json
// @Success 200 {object} res.GetArticleStats2Res "统计文章数据2响应"
// @Router /api/v1/articles/stats2 [get]
func (u *ArticleHandler) GetArticleStats2(c *gin.Context) {
	data, err := u.articleService.GetArticleStats2(c)
	if err != nil {
		c.JSON(errorutil.HTTPStatusOf(err), response.NewErrorResponse[any](err))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse[res.GetArticleStats2Res](data))
}
