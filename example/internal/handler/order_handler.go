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

type OrderHandler struct {
	orderService service.IOrderService
}

func NewOrderHandler(userService service.IOrderService) *OrderHandler {
	return &OrderHandler{
		orderService: userService,
	}
}

// CreateOrder 创建订单
// @Summary 创建订单
// @Description 创建订单
// @Tags mall
// @Accept json
// @Produce json
// @Param req body request.CreateOrderReq true "创建订单请求"
// @Success 200 {object} res.CreateOrderRes "创建订单响应"
// @Router /api/v1/orders [post]
func (u *OrderHandler) CreateOrder(c *gin.Context) {
	var req request.CreateOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse[any](errorutil.ParamsError.WithMessage(err.Error())))
		return
	}

	data, err := u.orderService.CreateOrder(c, req)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, response.NewErrorResponse[any](err))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse[res.CreateOrderRes](data))
}

// ListOrders 订单列表
// @Summary 订单列表
// @Description 订单列表
// @Tags mall
// @Accept json
// @Produce json
// @Param userId query integer false "按用户过滤" format(int64)
// @Param status query string false "订单状态,如 unpaid/paid/shipped"
// @Param page query integer false "页码" format(int32)
// @Param pageSize query integer false "每页条数" format(int32)
// @Success 200 {object} res.ListOrdersRes "订单列表响应"
// @Router /api/v1/user-orders [get]
func (u *OrderHandler) ListOrders(c *gin.Context) {
	var req request.ListOrdersReq
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse[any](errorutil.ParamsError.WithMessage(err.Error())))
		return
	}

	data, err := u.orderService.ListOrders(c, req)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, response.NewErrorResponse[any](err))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse[res.ListOrdersRes](data))
}

// GetOrderInfo 订单详情
// @Summary 订单详情
// @Description 订单详情
// @Tags mall
// @Accept json
// @Produce json
// @Param orderId path integer true "orderId" format(int64)
// @Success 200 {object} res.GetOrderInfoRes "订单详情响应"
// @Router /api/v1/orders/:orderId [get]
func (u *OrderHandler) GetOrderInfo(c *gin.Context) {
	orderIdStr := c.Param("orderId")
	orderId, err := strconv.ParseInt(orderIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse[any](errorutil.ParamsError.WithMessage(err.Error())))
		return
	}

	data, err := u.orderService.GetOrderInfo(c, orderId)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, response.NewErrorResponse[any](err))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse[res.GetOrderInfoRes](data))
}

// CancelOrder 取消订单
// @Summary 取消订单
// @Description 取消订单
// @Tags mall
// @Accept json
// @Produce json
// @Param orderId path integer true "orderId" format(int64)
// @Success 200 {object} bool "取消订单响应"
// @Router /api/v1/orders/:orderId/cancel [patch]
func (u *OrderHandler) CancelOrder(c *gin.Context) {
	orderIdStr := c.Param("orderId")
	orderId, err := strconv.ParseInt(orderIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse[any](errorutil.ParamsError.WithMessage(err.Error())))
		return
	}

	data, err := u.orderService.CancelOrder(c, orderId)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, response.NewErrorResponse[any](err))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse[bool](data))
}
