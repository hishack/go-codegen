package service

import (
	"codegen-example/internal/model/gen/request"
	res "codegen-example/internal/model/gen/response"

	"github.com/gin-gonic/gin"
)

type IOrderService interface {
	CreateOrder(c *gin.Context, req request.CreateOrderReq) (res.CreateOrderRes, error)
	ListOrders(c *gin.Context, req request.ListOrdersReq) (res.ListOrdersRes, error)
	GetOrderInfo(c *gin.Context, orderId int64) (res.GetOrderInfoRes, error)
	CancelOrder(c *gin.Context, orderId int64) (bool, error)
}
