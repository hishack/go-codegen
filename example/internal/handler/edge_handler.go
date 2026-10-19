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

type EdgeHandler struct {
	edgeService service.IEdgeService
}

func NewEdgeHandler(userService service.IEdgeService) *EdgeHandler {
	return &EdgeHandler{
		edgeService: userService,
	}
}

// HealthCheck 没有 req 也没有 res
// @Summary 没有 req 也没有 res
// @Description 没有 req 也没有 res
// @Tags system
// @Accept json
// @Produce json
// @Success 200 {object} bool "没有 req 也没有 res响应"
// @Router /api/v1/health [get]
func (u *EdgeHandler) HealthCheck(c *gin.Context) {
	err := u.edgeService.HealthCheck(c)
	if err != nil {
		c.JSON(errorutil.HTTPStatusOf(err), response.NewErrorResponse[any](err))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse[any](true))
}

// Logout 只有 req,没有 res
// @Summary 只有 req,没有 res
// @Description 只有 req,没有 res
// @Tags system
// @Accept json
// @Produce json
// @Param req body request.LogoutReq true "只有 req,没有 res请求"
// @Success 200 {object} bool "只有 req,没有 res响应"
// @Router /api/v1/logout [post]
func (u *EdgeHandler) Logout(c *gin.Context) {
	var req request.LogoutReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse[any](errorutil.ParamsError.WithMessage(err.Error())))
		return
	}

	err := u.edgeService.Logout(c, req)
	if err != nil {
		c.JSON(errorutil.HTTPStatusOf(err), response.NewErrorResponse[any](err))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse[any](true))
}

// GetTime 只有 res,没有 req
// @Summary 只有 res,没有 req
// @Description 只有 res,没有 req
// @Tags system
// @Accept json
// @Produce json
// @Success 200 {object} res.GetTimeRes "只有 res,没有 req响应"
// @Router /api/v1/time [get]
func (u *EdgeHandler) GetTime(c *gin.Context) {
	data, err := u.edgeService.GetTime(c)
	if err != nil {
		c.JSON(errorutil.HTTPStatusOf(err), response.NewErrorResponse[any](err))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse[res.GetTimeRes](data))
}

// Ping 空 req 对象
// @Summary 空 req 对象
// @Description 空 req 对象
// @Tags system
// @Accept json
// @Produce json
// @Success 200 {object} bool "空 req 对象响应"
// @Router /api/v1/ping [get]
func (u *EdgeHandler) Ping(c *gin.Context) {
	err := u.edgeService.Ping(c)
	if err != nil {
		c.JSON(errorutil.HTTPStatusOf(err), response.NewErrorResponse[any](err))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse[any](true))
}

// SearchUser GET 带 req(query 参数场景)
// @Summary GET 带 req(query 参数场景)
// @Description GET 带 req(query 参数场景)
// @Tags system
// @Accept json
// @Produce json
// @Param keyword query string false "keyword"
// @Param page query integer false "page" format(int32)
// @Param pageSize query integer false "pageSize" format(int32)
// @Success 200 {object} res.SearchUserRes "GET 带 req(query 参数场景)响应"
// @Router /api/v1/users/search [get]
func (u *EdgeHandler) SearchUser(c *gin.Context) {
	var req request.SearchUserReq
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse[any](errorutil.ParamsError.WithMessage(err.Error())))
		return
	}

	data, err := u.edgeService.SearchUser(c, req)
	if err != nil {
		c.JSON(errorutil.HTTPStatusOf(err), response.NewErrorResponse[any](err))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse[res.SearchUserRes](data))
}

// GetUserOrder 多路径参数
// @Summary 多路径参数
// @Description 多路径参数
// @Tags system
// @Accept json
// @Produce json
// @Param userId path integer true "userId" format(int64)
// @Param orderId path integer true "orderId" format(int64)
// @Success 200 {object} res.GetUserOrderRes "多路径参数响应"
// @Router /api/v1/users/:userId/orders/:orderId [get]
func (u *EdgeHandler) GetUserOrder(c *gin.Context) {
	userIdStr := c.Param("userId")
	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse[any](errorutil.ParamsError.WithMessage(err.Error())))
		return
	}
	orderIdStr := c.Param("orderId")
	orderId, err := strconv.ParseInt(orderIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse[any](errorutil.ParamsError.WithMessage(err.Error())))
		return
	}

	data, err := u.edgeService.GetUserOrder(c, userId, orderId)
	if err != nil {
		c.JSON(errorutil.HTTPStatusOf(err), response.NewErrorResponse[any](err))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse[res.GetUserOrderRes](data))
}

// GetArticle 无类型的路径参数,默认按 string
// @Summary 无类型的路径参数,默认按 string
// @Description 无类型的路径参数,默认按 string
// @Tags system
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} res.GetArticleRes "无类型的路径参数,默认按 string响应"
// @Router /api/v1/articles/:id [get]
func (u *EdgeHandler) GetArticle(c *gin.Context) {
	id := c.Param("id")

	data, err := u.edgeService.GetArticle(c, id)
	if err != nil {
		c.JSON(errorutil.HTTPStatusOf(err), response.NewErrorResponse[any](err))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse[res.GetArticleRes](data))
}

// GetOrgUser 多级路径参数
// @Summary 多级路径参数
// @Description 多级路径参数
// @Tags system
// @Accept json
// @Produce json
// @Param orgId path integer true "orgId" format(int64)
// @Param userId path integer true "userId" format(int64)
// @Success 200 {object} res.GetOrgUserRes "多级路径参数响应"
// @Router /api/v1/orgs/:orgId/users/:userId [get]
func (u *EdgeHandler) GetOrgUser(c *gin.Context) {
	orgIdStr := c.Param("orgId")
	orgId, err := strconv.ParseInt(orgIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse[any](errorutil.ParamsError.WithMessage(err.Error())))
		return
	}
	userIdStr := c.Param("userId")
	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse[any](errorutil.ParamsError.WithMessage(err.Error())))
		return
	}

	data, err := u.edgeService.GetOrgUser(c, orgId, userId)
	if err != nil {
		c.JSON(errorutil.HTTPStatusOf(err), response.NewErrorResponse[any](err))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse[res.GetOrgUserRes](data))
}
