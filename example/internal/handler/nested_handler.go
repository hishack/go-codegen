package handler

import (
	"net/http"
	"strconv"

	res "codegen-example/internal/model/gen/response"
	"codegen-example/internal/service"
	"codegen-example/pkg/errorutil"
	"codegen-example/pkg/response"

	"github.com/gin-gonic/gin"
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
// @Param userId path integer true "userId"
// @Success 200 {object} res.GetUserDetailRes "深度嵌套:res 会提升出多层命名结构体 UserRes -> UserResProfile -> UserResProfileAddress响应"
// @Router /users/:userId/detail [get]
func (u *NestedHandler) GetUserDetail(c *gin.Context) {
	userIdStr := c.Param("userId")
	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, response.NewErrorResponse[any](errorutil.ParamsError))
	}

	data, err := u.nestedService.GetUserDetail(c, userId)
	if err != nil {
		c.JSON(http.StatusOK, response.NewErrorResponse[any](err))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse[res.GetUserDetailRes](data))
}
// GetOrderDetail 数组套对象套对象
// @Summary 数组套对象套对象
// @Description 数组套对象套对象
// @Tags system
// @Accept json
// @Produce json
// @Param orderId path integer true "orderId"
// @Success 200 {object} res.GetOrderDetailRes "数组套对象套对象响应"
// @Router /orders/:orderId/detail [get]
func (u *NestedHandler) GetOrderDetail(c *gin.Context) {
	orderIdStr := c.Param("orderId")
	orderId, err := strconv.ParseInt(orderIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, response.NewErrorResponse[any](errorutil.ParamsError))
	}

	data, err := u.nestedService.GetOrderDetail(c, orderId)
	if err != nil {
		c.JSON(http.StatusOK, response.NewErrorResponse[any](err))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse[res.GetOrderDetailRes](data))
}

// GetConfig map 类型
// @Summary map 类型
// @Description map 类型
// @Tags system
// @Accept json
// @Produce json
// @Success 200 {object} res.GetConfigRes "map 类型响应"
// @Router /config [get]
func (u *NestedHandler) GetConfig(c *gin.Context) {
	data, err := u.nestedService.GetConfig(c)
	if err != nil {
		c.JSON(http.StatusOK, response.NewErrorResponse[any](err))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse[res.GetConfigRes](data))
}

// GetUserMap GetUserMap
// @Summary GetUserMap
// @Description GetUserMap
// @Tags system
// @Accept json
// @Produce json
// @Success 200 {object} res.GetUserMapRes "GetUserMap响应"
// @Router /users/map [get]
func (u *NestedHandler) GetUserMap(c *gin.Context) {
	data, err := u.nestedService.GetUserMap(c)
	if err != nil {
		c.JSON(http.StatusOK, response.NewErrorResponse[any](err))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse[res.GetUserMapRes](data))
}

// GetConfigA 同名字段但结构不同 -> 靠 ApiName 前缀区分,不会撞车
// @Summary 同名字段但结构不同 -> 靠 ApiName 前缀区分,不会撞车
// @Description 同名字段但结构不同 -> 靠 ApiName 前缀区分,不会撞车
// @Tags system
// @Accept json
// @Produce json
// @Success 200 {object} res.GetConfigARes "同名字段但结构不同 -> 靠 ApiName 前缀区分,不会撞车响应"
// @Router /config/a [get]
func (u *NestedHandler) GetConfigA(c *gin.Context) {
	data, err := u.nestedService.GetConfigA(c)
	if err != nil {
		c.JSON(http.StatusOK, response.NewErrorResponse[any](err))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse[res.GetConfigARes](data))
}

// GetConfigB GetConfigB
// @Summary GetConfigB
// @Description GetConfigB
// @Tags system
// @Accept json
// @Produce json
// @Success 200 {object} res.GetConfigBRes "GetConfigB响应"
// @Router /config/b [get]
func (u *NestedHandler) GetConfigB(c *gin.Context) {
	data, err := u.nestedService.GetConfigB(c)
	if err != nil {
		c.JSON(http.StatusOK, response.NewErrorResponse[any](err))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse[res.GetConfigBRes](data))
}
