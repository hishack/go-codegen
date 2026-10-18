package handler

import (
	"context"

	"codegen-example/internal/api"
	"codegen-example/internal/dal/vo"
	"codegen-example/internal/service"
	"codegen-example/pkg/response"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

type ArrayHandler struct {
	arrayService service.IArrayService
}

func NewArrayHandler(userService service.IArrayService) *ArrayHandler {
	return &ArrayHandler{
		arrayService: userService,
	}
}

// GetIntArray 数组的各种写法:基础类型数组、对象数组、混合嵌套
// @Summary 数组的各种写法:基础类型数组、对象数组、混合嵌套
// @Description 数组的各种写法:基础类型数组、对象数组、混合嵌套
// @Tags system
// @Accept json
// @Produce json
// @Success 200 {object} api.GetIntArrayResponse "数组的各种写法:基础类型数组、对象数组、混合嵌套响应"
// @Router /arrays/ints [get]
func (u *ArrayHandler) GetIntArray(ctx context.Context, c *app.RequestContext) {
	res, err := u.arrayService.GetIntArray(ctx)
	if err != nil {
		c.JSON(consts.StatusOK, response.NewErrorResponse[any](err))
		return
	}
	c.JSON(consts.StatusOK, response.NewSuccessResponse[vo.GetIntArrayVo](res))
}

// GetStringArray GetStringArray
// @Summary GetStringArray
// @Description GetStringArray
// @Tags system
// @Accept json
// @Produce json
// @Success 200 {object} api.GetStringArrayResponse "GetStringArray响应"
// @Router /arrays/strings [get]
func (u *ArrayHandler) GetStringArray(ctx context.Context, c *app.RequestContext) {
	res, err := u.arrayService.GetStringArray(ctx)
	if err != nil {
		c.JSON(consts.StatusOK, response.NewErrorResponse[any](err))
		return
	}
	c.JSON(consts.StatusOK, response.NewSuccessResponse[vo.GetStringArrayVo](res))
}

// GetInt64Array GetInt64Array
// @Summary GetInt64Array
// @Description GetInt64Array
// @Tags system
// @Accept json
// @Produce json
// @Success 200 {object} api.GetInt64ArrayResponse "GetInt64Array响应"
// @Router /arrays/int64s [get]
func (u *ArrayHandler) GetInt64Array(ctx context.Context, c *app.RequestContext) {
	res, err := u.arrayService.GetInt64Array(ctx)
	if err != nil {
		c.JSON(consts.StatusOK, response.NewErrorResponse[any](err))
		return
	}
	c.JSON(consts.StatusOK, response.NewSuccessResponse[vo.GetInt64ArrayVo](res))
}

// GetObjectArray 对象数组:对应 res.list,会提升为 GetObjectArrayResListItem
// @Summary 对象数组:对应 res.list,会提升为 GetObjectArrayResListItem
// @Description 对象数组:对应 res.list,会提升为 GetObjectArrayResListItem
// @Tags system
// @Accept json
// @Produce json
// @Success 200 {object} api.GetObjectArrayResponse "对象数组:对应 res.list,会提升为 GetObjectArrayResListItem响应"
// @Router /arrays/objects [get]
func (u *ArrayHandler) GetObjectArray(ctx context.Context, c *app.RequestContext) {
	res, err := u.arrayService.GetObjectArray(ctx)
	if err != nil {
		c.JSON(consts.StatusOK, response.NewErrorResponse[any](err))
		return
	}
	c.JSON(consts.StatusOK, response.NewSuccessResponse[vo.GetObjectArrayVo](res))
}

// GetOrderList 多字段对象数组
// @Summary 多字段对象数组
// @Description 多字段对象数组
// @Tags system
// @Accept json
// @Produce json
// @Success 200 {object} api.GetOrderListResponse "多字段对象数组响应"
// @Router /orders [get]
func (u *ArrayHandler) GetOrderList(ctx context.Context, c *app.RequestContext) {
	res, err := u.arrayService.GetOrderList(ctx)
	if err != nil {
		c.JSON(consts.StatusOK, response.NewErrorResponse[any](err))
		return
	}
	c.JSON(consts.StatusOK, response.NewSuccessResponse[vo.GetOrderListVo](res))
}

// CreateBatchUser 请求里带数组
// @Summary 请求里带数组
// @Description 请求里带数组
// @Tags system
// @Accept json
// @Produce json
// @Param req body api.CreateBatchUserRequest true "请求里带数组请求"
// @Success 200 {object} api.CreateBatchUserResponse "请求里带数组响应"
// @Router /users/batch [post]
func (u *ArrayHandler) CreateBatchUser(ctx context.Context, c *app.RequestContext) {
	req := &api.CreateBatchUserRequest{}
	err := c.BindAndValidate(req)
	if err != nil {
		c.JSON(consts.StatusOK, response.NewErrorResponse[any](err))
		return
	}

	res, err := u.arrayService.CreateBatchUser(ctx, req)
	if err != nil {
		c.JSON(consts.StatusOK, response.NewErrorResponse[any](err))
		return
	}
	c.JSON(consts.StatusOK, response.NewSuccessResponse[int64](res))
}
