package handler

import (
	"context"
	"strconv"

	"codegen-example/internal/api"
	"codegen-example/internal/dal/vo"
	"codegen-example/internal/service"
	"codegen-example/pkg/errorutil"
	"codegen-example/pkg/response"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
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
// @Router /health [get]
func (u *EdgeHandler) HealthCheck(ctx context.Context, c *app.RequestContext) {
	err := u.edgeService.HealthCheck(ctx)
	if err != nil {
		c.JSON(consts.StatusOK, response.NewErrorResponse[any](err))
		return
	}
	c.JSON(consts.StatusOK, response.NewSuccessResponse[any](true))
}

// Logout 只有 req,没有 res
// @Summary 只有 req,没有 res
// @Description 只有 req,没有 res
// @Tags system
// @Accept json
// @Produce json
// @Param req body api.LogoutRequest true "只有 req,没有 res请求"
// @Success 200 {object} bool "只有 req,没有 res响应"
// @Router /logout [post]
func (u *EdgeHandler) Logout(ctx context.Context, c *app.RequestContext) {
	req := &api.LogoutRequest{}
	err := c.BindAndValidate(req)
	if err != nil {
		c.JSON(consts.StatusOK, response.NewErrorResponse[any](err))
		return
	}

	err = u.edgeService.Logout(ctx, req)
	if err != nil {
		c.JSON(consts.StatusOK, response.NewErrorResponse[any](err))
		return
	}
	c.JSON(consts.StatusOK, response.NewSuccessResponse[any](true))
}

// GetTime 只有 res,没有 req
// @Summary 只有 res,没有 req
// @Description 只有 res,没有 req
// @Tags system
// @Accept json
// @Produce json
// @Success 200 {object} api.GetTimeResponse "只有 res,没有 req响应"
// @Router /time [get]
func (u *EdgeHandler) GetTime(ctx context.Context, c *app.RequestContext) {
	res, err := u.edgeService.GetTime(ctx)
	if err != nil {
		c.JSON(consts.StatusOK, response.NewErrorResponse[any](err))
		return
	}
	c.JSON(consts.StatusOK, response.NewSuccessResponse[vo.GetTimeVo](res))
}

// Ping 空 req 对象
// @Summary 空 req 对象
// @Description 空 req 对象
// @Tags system
// @Accept json
// @Produce json
// @Success 200 {object} bool "空 req 对象响应"
// @Router /ping [get]
func (u *EdgeHandler) Ping(ctx context.Context, c *app.RequestContext) {
	err := u.edgeService.Ping(ctx)
	if err != nil {
		c.JSON(consts.StatusOK, response.NewErrorResponse[any](err))
		return
	}
	c.JSON(consts.StatusOK, response.NewSuccessResponse[any](true))
}

// SearchUser GET 带 req(query 参数场景)
// @Summary GET 带 req(query 参数场景)
// @Description GET 带 req(query 参数场景)
// @Tags system
// @Accept json
// @Produce json
// @Param keyword query string false "keyword"
// @Param page query integer false "page"
// @Param pageSize query integer false "pageSize"
// @Success 200 {object} api.SearchUserResponse "GET 带 req(query 参数场景)响应"
// @Router /users/search [get]
func (u *EdgeHandler) SearchUser(ctx context.Context, c *app.RequestContext) {
	req := &api.SearchUserRequest{}
	err := c.BindAndValidate(req)
	if err != nil {
		c.JSON(consts.StatusOK, response.NewErrorResponse[any](err))
		return
	}

	res, err := u.edgeService.SearchUser(ctx, req)
	if err != nil {
		c.JSON(consts.StatusOK, response.NewErrorResponse[any](err))
		return
	}
	c.JSON(consts.StatusOK, response.NewSuccessResponse[vo.SearchUserVo](res))
}

// GetUserOrder 多路径参数
// @Summary 多路径参数
// @Description 多路径参数
// @Tags system
// @Accept json
// @Produce json
// @Param userId path integer true "userId"
// @Param orderId path integer true "orderId"
// @Success 200 {object} api.GetUserOrderResponse "多路径参数响应"
// @Router /users/:userId/orders/:orderId [get]
func (u *EdgeHandler) GetUserOrder(ctx context.Context, c *app.RequestContext) {
	userIdStr := c.Param("userId")
	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		c.JSON(consts.StatusOK, response.NewErrorResponse[any](errorutil.ParamsError))
	}
	orderIdStr := c.Param("orderId")
	orderId, err := strconv.ParseInt(orderIdStr, 10, 64)
	if err != nil {
		c.JSON(consts.StatusOK, response.NewErrorResponse[any](errorutil.ParamsError))
	}

	res, err := u.edgeService.GetUserOrder(ctx, userId, orderId)
	if err != nil {
		c.JSON(consts.StatusOK, response.NewErrorResponse[any](err))
		return
	}
	c.JSON(consts.StatusOK, response.NewSuccessResponse[vo.GetUserOrderVo](res))
}

// GetArticle 无类型的路径参数,默认按 string
// @Summary 无类型的路径参数,默认按 string
// @Description 无类型的路径参数,默认按 string
// @Tags system
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} api.GetArticleResponse "无类型的路径参数,默认按 string响应"
// @Router /articles/:id [get]
func (u *EdgeHandler) GetArticle(ctx context.Context, c *app.RequestContext) {
	id := c.Param("id")

	res, err := u.edgeService.GetArticle(ctx, id)
	if err != nil {
		c.JSON(consts.StatusOK, response.NewErrorResponse[any](err))
		return
	}
	c.JSON(consts.StatusOK, response.NewSuccessResponse[vo.GetArticleVo](res))
}

// GetOrgUser 多级路径参数
// @Summary 多级路径参数
// @Description 多级路径参数
// @Tags system
// @Accept json
// @Produce json
// @Param orgId path integer true "orgId"
// @Param userId path integer true "userId"
// @Success 200 {object} api.GetOrgUserResponse "多级路径参数响应"
// @Router /orgs/:orgId/users/:userId [get]
func (u *EdgeHandler) GetOrgUser(ctx context.Context, c *app.RequestContext) {
	orgIdStr := c.Param("orgId")
	orgId, err := strconv.ParseInt(orgIdStr, 10, 64)
	if err != nil {
		c.JSON(consts.StatusOK, response.NewErrorResponse[any](errorutil.ParamsError))
	}
	userIdStr := c.Param("userId")
	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		c.JSON(consts.StatusOK, response.NewErrorResponse[any](errorutil.ParamsError))
	}

	res, err := u.edgeService.GetOrgUser(ctx, orgId, userId)
	if err != nil {
		c.JSON(consts.StatusOK, response.NewErrorResponse[any](err))
		return
	}
	c.JSON(consts.StatusOK, response.NewSuccessResponse[vo.GetOrgUserVo](res))
}
