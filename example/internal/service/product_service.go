package service

import (
	"codegen-example/internal/model/gen/request"
	res "codegen-example/internal/model/gen/response"

	"github.com/gin-gonic/gin"
)

type IProductService interface {
	ListProducts(c *gin.Context, req request.ListProductsReq) (res.ListProductsRes, error)
	GetProduct(c *gin.Context, id int64) (res.GetProductRes, error)
	CreateProduct(c *gin.Context, req request.CreateProductReq) (int64, error)
	UpdateProduct(c *gin.Context, id int64, req request.UpdateProductReq) (bool, error)
	DeleteProduct(c *gin.Context, id int64) (bool, error)
}
