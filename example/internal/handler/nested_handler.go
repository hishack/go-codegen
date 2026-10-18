package handler

import (
	"context"
	"strconv"

	"codegen-example/internal/dal/vo"
	"codegen-example/internal/service"
	"codegen-example/pkg/errorutil"
	"codegen-example/pkg/response"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

type NestedHandler struct {
	nestedService service.INestedService
}

func NewNestedHandler(userService service.INestedService) *NestedHandler {
	return &NestedHandler{
		nestedService: userService,
	}
}

// GetUserDetail 深度嵌套:res 会提升出多层命名结构体 UserRes -> UserResProfile -> UserResProfileAddress
// @Summary 深度嵌套:res 会提升出多层命名结构体 UserRes -> UserResProfile -> UserResProfileAddress
// @Description 深度嵌套:res 会提升出多层命名结构体 UserRes -> UserResProfile -> UserResProfileAddress
// @Tags system
// @Accept json
// @Produce json
// @Param id path integer true "id"
// @Success 200 {object} api.GetUserDetailResponse "深度嵌套:res 会提升出多层命名结构体 UserRes -> UserResProfile -> UserResProfileAddress响应"
// @Router /users/:id/detail [get]
func (u *NestedHandler) GetUserDetail(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(consts.StatusOK, response.NewErrorResponse[any](errorutil.ParamsError))
	}

	res, err := u.nestedService.GetUserDetail(ctx, id)
	if err != nil {
		c.JSON(consts.StatusOK, response.NewErrorResponse[any](err))
		return
	}
	c.JSON(consts.StatusOK, response.NewSuccessResponse[vo.GetUserDetailVo](res))
}

// GetOrderDetail 数组套对象套对象
// @Summary 数组套对象套对象
// @Description 数组套对象套对象
// @Tags system
// @Accept json
// @Produce json
// @Param orderId path integer true "orderId"
// @Success 200 {object} api.GetOrderDetailResponse "数组套对象套对象响应"
// @Router /orders/:orderId/detail [get]
func (u *NestedHandler) GetOrderDetail(ctx context.Context, c *app.RequestContext) {
	orderIdStr := c.Param("orderId")
	orderId, err := strconv.ParseInt(orderIdStr, 10, 64)
	if err != nil {
		c.JSON(consts.StatusOK, response.NewErrorResponse[any](errorutil.ParamsError))
	}

	res, err := u.nestedService.GetOrderDetail(ctx, orderId)
	if err != nil {
		c.JSON(consts.StatusOK, response.NewErrorResponse[any](err))
		return
	}
	c.JSON(consts.StatusOK, response.NewSuccessResponse[vo.GetOrderDetailVo](res))
}

// GetConfig map 类型
// @Summary map 类型
// @Description map 类型
// @Tags system
// @Accept json
// @Produce json
// @Success 200 {object} api.GetConfigResponse "map 类型响应"
// @Router /config [get]
func (u *NestedHandler) GetConfig(ctx context.Context, c *app.RequestContext) {
	res, err := u.nestedService.GetConfig(ctx)
	if err != nil {
		c.JSON(consts.StatusOK, response.NewErrorResponse[any](err))
		return
	}
	c.JSON(consts.StatusOK, response.NewSuccessResponse[vo.GetConfigVo](res))
}

// GetUserMap GetUserMap
// @Summary GetUserMap
// @Description GetUserMap
// @Tags system
// @Accept json
// @Produce json
// @Success 200 {object} api.GetUserMapResponse "GetUserMap响应"
// @Router /users/map [get]
func (u *NestedHandler) GetUserMap(ctx context.Context, c *app.RequestContext) {
	res, err := u.nestedService.GetUserMap(ctx)
	if err != nil {
		c.JSON(consts.StatusOK, response.NewErrorResponse[any](err))
		return
	}
	c.JSON(consts.StatusOK, response.NewSuccessResponse[vo.GetUserMapVo](res))
}

// GetConfigA 同名字段但结构不同 -> 靠 ApiName 前缀区分,不会撞车
// @Summary 同名字段但结构不同 -> 靠 ApiName 前缀区分,不会撞车
// @Description 同名字段但结构不同 -> 靠 ApiName 前缀区分,不会撞车
// @Tags system
// @Accept json
// @Produce json
// @Success 200 {object} api.GetConfigAResponse "同名字段但结构不同 -> 靠 ApiName 前缀区分,不会撞车响应"
// @Router /config/a [get]
func (u *NestedHandler) GetConfigA(ctx context.Context, c *app.RequestContext) {
	res, err := u.nestedService.GetConfigA(ctx)
	if err != nil {
		c.JSON(consts.StatusOK, response.NewErrorResponse[any](err))
		return
	}
	c.JSON(consts.StatusOK, response.NewSuccessResponse[vo.GetConfigAVo](res))
}

// GetConfigB GetConfigB
// @Summary GetConfigB
// @Description GetConfigB
// @Tags system
// @Accept json
// @Produce json
// @Success 200 {object} api.GetConfigBResponse "GetConfigB响应"
// @Router /config/b [get]
func (u *NestedHandler) GetConfigB(ctx context.Context, c *app.RequestContext) {
	res, err := u.nestedService.GetConfigB(ctx)
	if err != nil {
		c.JSON(consts.StatusOK, response.NewErrorResponse[any](err))
		return
	}
	c.JSON(consts.StatusOK, response.NewSuccessResponse[vo.GetConfigBVo](res))
}
