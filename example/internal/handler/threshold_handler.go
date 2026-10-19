package handler

import (
	"net/http"
	"strconv"

	"codegen-example/internal/model/gen/request"
	"codegen-example/internal/service"
	"codegen-example/pkg/errorutil"
	"codegen-example/pkg/response"

	"github.com/gin-gonic/gin"
)

type ThresholdHandler struct {
	thresholdService service.IThresholdService
}

func NewThresholdHandler(userService service.IThresholdService) *ThresholdHandler {
	return &ThresholdHandler{
		thresholdService: userService,
	}
}

// CreateThreshold 新增阈值:创建一个成长阈值,例如"等级达到 10 需要 1000 经验"
// @Summary 新增阈值:创建一个成长阈值,例如"等级达到 10 需要 1000 经验"
// @Description 新增阈值:创建一个成长阈值,例如"等级达到 10 需要 1000 经验"
// @Tags system
// @Accept json
// @Produce json
// @Param req body request.CreateThresholdReq true "新增阈值:创建一个成长阈值,例如"等级达到 10 需要 1000 经验"请求"
// @Success 200 {object} int64 "新增阈值:创建一个成长阈值,例如"等级达到 10 需要 1000 经验"响应"
// @Router /thresholds [post]
func (u *ThresholdHandler) CreateThreshold(c *gin.Context) {
	var req request.CreateThresholdReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, response.NewErrorResponse[any](err))
		return
	}

	data, err := u.thresholdService.CreateThreshold(c, req)
	if err != nil {
		c.JSON(http.StatusOK, response.NewErrorResponse[any](err))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse[int64](data))
}

// UpdateThreshold 更新阈值:按主键局部更新阈值配置
// @Summary 更新阈值:按主键局部更新阈值配置
// @Description 更新阈值:按主键局部更新阈值配置
// @Tags system
// @Accept json
// @Produce json
// @Param id path integer true "id"
// @Param req body request.UpdateThresholdReq true "更新阈值:按主键局部更新阈值配置请求"
// @Success 200 {object} bool "更新阈值:按主键局部更新阈值配置响应"
// @Router /thresholds/:id [put]
func (u *ThresholdHandler) UpdateThreshold(c *gin.Context) {
	var req request.UpdateThresholdReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, response.NewErrorResponse[any](err))
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, response.NewErrorResponse[any](errorutil.ParamsError))
	}

	data, err := u.thresholdService.UpdateThreshold(c, id, req)
	if err != nil {
		c.JSON(http.StatusOK, response.NewErrorResponse[any](err))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse[bool](data))
}
