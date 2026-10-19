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

type NotifyHandler struct {
	notifyService service.INotifyService
}

func NewNotifyHandler(userService service.INotifyService) *NotifyHandler {
	return &NotifyHandler{
		notifyService: userService,
	}
}

// SendSmsCode 发送短信验证码
// @Summary 发送短信验证码
// @Description 发送短信验证码
// @Tags system
// @Accept json
// @Produce json
// @Param req body request.SendSmsCodeReq true "发送短信验证码请求"
// @Success 200 {object} int "发送短信验证码响应"
// @Router /api/v1/notify/sms [post]
func (u *NotifyHandler) SendSmsCode(c *gin.Context) {
	var req request.SendSmsCodeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse[any](errorutil.ParamsError.WithMessage(err.Error())))
		return
	}

	data, err := u.notifyService.SendSmsCode(c, req)
	if err != nil {
		c.JSON(errorutil.HTTPStatusOf(err), response.NewErrorResponse[any](err))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse[int](data))
}

// ListNotifies 查询通知列表
// @Summary 查询通知列表
// @Description 查询通知列表
// @Tags system
// @Accept json
// @Produce json
// @Param userId query integer false "按用户过滤" format(int64)
// @Param read query boolean false "是否已读"
// @Success 200 {object} res.ListNotifiesRes "查询通知列表响应"
// @Router /api/v1/notifies [get]
func (u *NotifyHandler) ListNotifies(c *gin.Context) {
	var req request.ListNotifiesReq
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse[any](errorutil.ParamsError.WithMessage(err.Error())))
		return
	}

	data, err := u.notifyService.ListNotifies(c, req)
	if err != nil {
		c.JSON(errorutil.HTTPStatusOf(err), response.NewErrorResponse[any](err))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse[res.ListNotifiesRes](data))
}

// ReadNotify 标记通知已读
// @Summary 标记通知已读
// @Description 标记通知已读
// @Tags system
// @Accept json
// @Produce json
// @Param id path integer true "id" format(int64)
// @Success 200 {object} bool "标记通知已读响应"
// @Router /api/v1/notifies/:id/read [patch]
func (u *NotifyHandler) ReadNotify(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse[any](errorutil.ParamsError.WithMessage(err.Error())))
		return
	}

	data, err := u.notifyService.ReadNotify(c, id)
	if err != nil {
		c.JSON(errorutil.HTTPStatusOf(err), response.NewErrorResponse[any](err))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse[bool](data))
}

// GetNotifyConfig 扩展属性示例:map 类型
// @Summary 扩展属性示例:map 类型
// @Description 扩展属性示例:map 类型
// @Tags system
// @Accept json
// @Produce json
// @Success 200 {object} res.GetNotifyConfigRes "扩展属性示例:map 类型响应"
// @Router /api/v1/notifies/config [get]
func (u *NotifyHandler) GetNotifyConfig(c *gin.Context) {
	data, err := u.notifyService.GetNotifyConfig(c)
	if err != nil {
		c.JSON(errorutil.HTTPStatusOf(err), response.NewErrorResponse[any](err))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse[res.GetNotifyConfigRes](data))
}
