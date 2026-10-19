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

type ProductHandler struct {
	productService service.IProductService
}

func NewProductHandler(userService service.IProductService) *ProductHandler {
	return &ProductHandler{
		productService: userService,
	}
}

// ListProducts 商品列表查询
// @Summary 商品列表查询
// @Description 商品列表查询
// @Tags mall
// @Accept json
// @Produce json
// @Param keyword query string false "搜索关键字"
// @Param categoryId query integer false "分类 id" format(int64)
// @Param page query integer false "页码,从 1 开始" format(int32)
// @Param pageSize query integer false "每页条数" format(int32)
// @Success 200 {object} res.ListProductsRes "商品列表查询响应"
// @Router /api/v1/products [get]
func (u *ProductHandler) ListProducts(c *gin.Context) {
	var req request.ListProductsReq
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse[any](errorutil.ParamsError.WithMessage(err.Error())))
		return
	}

	data, err := u.productService.ListProducts(c, req)
	if err != nil {
		c.JSON(errorutil.HTTPStatusOf(err), response.NewErrorResponse[any](err))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse[res.ListProductsRes](data))
}

// GetProduct 商品详情
// @Summary 商品详情
// @Description 商品详情
// @Tags mall
// @Accept json
// @Produce json
// @Param id path integer true "id" format(int64)
// @Success 200 {object} res.GetProductRes "商品详情响应"
// @Router /api/v1/products/:id [get]
func (u *ProductHandler) GetProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse[any](errorutil.ParamsError.WithMessage(err.Error())))
		return
	}

	data, err := u.productService.GetProduct(c, id)
	if err != nil {
		c.JSON(errorutil.HTTPStatusOf(err), response.NewErrorResponse[any](err))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse[res.GetProductRes](data))
}

// CreateProduct 创建商品
// @Summary 创建商品
// @Description 创建商品
// @Tags mall
// @Accept json
// @Produce json
// @Param req body request.CreateProductReq true "创建商品请求"
// @Success 200 {object} int64 "创建商品响应"
// @Router /api/v1/products [post]
func (u *ProductHandler) CreateProduct(c *gin.Context) {
	var req request.CreateProductReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse[any](errorutil.ParamsError.WithMessage(err.Error())))
		return
	}

	data, err := u.productService.CreateProduct(c, req)
	if err != nil {
		c.JSON(errorutil.HTTPStatusOf(err), response.NewErrorResponse[any](err))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse[int64](data))
}

// UpdateProduct 更新商品
// @Summary 更新商品
// @Description 更新商品
// @Tags mall
// @Accept json
// @Produce json
// @Param id path integer true "id" format(int64)
// @Param req body request.UpdateProductReq true "更新商品请求"
// @Success 200 {object} bool "更新商品响应"
// @Router /api/v1/products/:id [put]
func (u *ProductHandler) UpdateProduct(c *gin.Context) {
	var req request.UpdateProductReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse[any](errorutil.ParamsError.WithMessage(err.Error())))
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse[any](errorutil.ParamsError.WithMessage(err.Error())))
		return
	}

	data, err := u.productService.UpdateProduct(c, id, req)
	if err != nil {
		c.JSON(errorutil.HTTPStatusOf(err), response.NewErrorResponse[any](err))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse[bool](data))
}

// DeleteProduct 删除商品
// @Summary 删除商品
// @Description 删除商品
// @Tags mall
// @Accept json
// @Produce json
// @Param id path integer true "id" format(int64)
// @Success 200 {object} bool "删除商品响应"
// @Router /api/v1/products/:id [delete]
func (u *ProductHandler) DeleteProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse[any](errorutil.ParamsError.WithMessage(err.Error())))
		return
	}

	data, err := u.productService.DeleteProduct(c, id)
	if err != nil {
		c.JSON(errorutil.HTTPStatusOf(err), response.NewErrorResponse[any](err))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse[bool](data))
}
