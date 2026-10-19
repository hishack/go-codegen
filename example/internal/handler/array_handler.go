package handler

import (
	"net/http"

	"codegen-example/internal/model/gen/request"
	res "codegen-example/internal/model/gen/response"
	"codegen-example/internal/service"
	"codegen-example/pkg/errorutil"
	"codegen-example/pkg/response"

	"github.com/gin-gonic/gin"
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
// @Success 200 {object} res.GetIntArrayRes "数组的各种写法:基础类型数组、对象数组、混合嵌套响应"
// @Router /api/v1/arrays/ints [get]
func (u *ArrayHandler) GetIntArray(c *gin.Context) {
	data, err := u.arrayService.GetIntArray(c)
	if err != nil {
		c.JSON(errorutil.HTTPStatusOf(err), response.NewErrorResponse[any](err))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse[res.GetIntArrayRes](data))
}

// GetStringArray GetStringArray
// @Summary GetStringArray
// @Description GetStringArray
// @Tags system
// @Accept json
// @Produce json
// @Success 200 {object} res.GetStringArrayRes "GetStringArray响应"
// @Router /api/v1/arrays/strings [get]
func (u *ArrayHandler) GetStringArray(c *gin.Context) {
	data, err := u.arrayService.GetStringArray(c)
	if err != nil {
		c.JSON(errorutil.HTTPStatusOf(err), response.NewErrorResponse[any](err))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse[res.GetStringArrayRes](data))
}

// GetInt64Array GetInt64Array
// @Summary GetInt64Array
// @Description GetInt64Array
// @Tags system
// @Accept json
// @Produce json
// @Success 200 {object} res.GetInt64ArrayRes "GetInt64Array响应"
// @Router /api/v1/arrays/int64s [get]
func (u *ArrayHandler) GetInt64Array(c *gin.Context) {
	data, err := u.arrayService.GetInt64Array(c)
	if err != nil {
		c.JSON(errorutil.HTTPStatusOf(err), response.NewErrorResponse[any](err))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse[res.GetInt64ArrayRes](data))
}

// GetObjectArray 对象数组:对应 res.list,会提升为 GetObjectArrayResListItem
// @Summary 对象数组:对应 res.list,会提升为 GetObjectArrayResListItem
// @Description 对象数组:对应 res.list,会提升为 GetObjectArrayResListItem
// @Tags system
// @Accept json
// @Produce json
// @Success 200 {object} res.GetObjectArrayRes "对象数组:对应 res.list,会提升为 GetObjectArrayResListItem响应"
// @Router /api/v1/arrays/objects [get]
func (u *ArrayHandler) GetObjectArray(c *gin.Context) {
	data, err := u.arrayService.GetObjectArray(c)
	if err != nil {
		c.JSON(errorutil.HTTPStatusOf(err), response.NewErrorResponse[any](err))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse[res.GetObjectArrayRes](data))
}

// GetOrderList 多字段对象数组
// @Summary 多字段对象数组
// @Description 多字段对象数组
// @Tags system
// @Accept json
// @Produce json
// @Success 200 {object} res.GetOrderListRes "多字段对象数组响应"
// @Router /api/v1/orders [get]
func (u *ArrayHandler) GetOrderList(c *gin.Context) {
	data, err := u.arrayService.GetOrderList(c)
	if err != nil {
		c.JSON(errorutil.HTTPStatusOf(err), response.NewErrorResponse[any](err))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse[res.GetOrderListRes](data))
}

// CreateBatchUser 请求里带数组
// @Summary 请求里带数组
// @Description 请求里带数组
// @Tags system
// @Accept json
// @Produce json
// @Param req body request.CreateBatchUserReq true "请求里带数组请求"
// @Success 200 {object} int64 "请求里带数组响应"
// @Router /api/v1/users/batch [post]
func (u *ArrayHandler) CreateBatchUser(c *gin.Context) {
	var req request.CreateBatchUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse[any](errorutil.ParamsError.WithMessage(err.Error())))
		return
	}

	data, err := u.arrayService.CreateBatchUser(c, req)
	if err != nil {
		c.JSON(errorutil.HTTPStatusOf(err), response.NewErrorResponse[any](err))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse[int64](data))
}
