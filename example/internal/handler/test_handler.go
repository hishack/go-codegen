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

type TestHandler struct {
	testService service.ITestService
}

func NewTestHandler(userService service.ITestService) *TestHandler {
	return &TestHandler{
		testService: userService,
	}
}

// TestPing ping 测试
// @Summary ping 测试
// @Description ping 测试
// @Tags test
// @Accept json
// @Produce json
// @Success 200 {object} res.TestPingRes "ping 测试响应"
// @Router /api/v1/test/ping [get]
func (u *TestHandler) TestPing(c *gin.Context) {
	data, err := u.testService.TestPing(c)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, response.NewErrorResponse[any](err))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse[res.TestPingRes](data))
}

// TestHello hello 测试
// @Summary hello 测试
// @Description hello 测试
// @Tags test
// @Accept json
// @Produce json
// @Success 200 {object} res.TestHelloRes "hello 测试响应"
// @Router /api/v1/test/hello [get]
func (u *TestHandler) TestHello(c *gin.Context) {
	data, err := u.testService.TestHello(c)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, response.NewErrorResponse[any](err))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse[res.TestHelloRes](data))
}

// TestLogger 日志组件测试
// @Summary 日志组件测试
// @Description 日志组件测试
// @Tags test
// @Accept json
// @Produce json
// @Success 200 {object} res.TestLoggerRes "日志组件测试响应"
// @Router /api/v1/test/logger [get]
func (u *TestHandler) TestLogger(c *gin.Context) {
	data, err := u.testService.TestLogger(c)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, response.NewErrorResponse[any](err))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse[res.TestLoggerRes](data))
}

// TestRedis redis 读写测试
// @Summary redis 读写测试
// @Description redis 读写测试
// @Tags test
// @Accept json
// @Produce json
// @Param req body request.TestRedisReq true "redis 读写测试请求"
// @Success 200 {object} res.TestRedisRes "redis 读写测试响应"
// @Router /api/v1/test/redis [post]
func (u *TestHandler) TestRedis(c *gin.Context) {
	var req request.TestRedisReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse[any](errorutil.ParamsError.WithMessage(err.Error())))
		return
	}

	data, err := u.testService.TestRedis(c, req)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, response.NewErrorResponse[any](err))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse[res.TestRedisRes](data))
}

// TestSuccess 成功响应校验
// @Summary 成功响应校验
// @Description 成功响应校验
// @Tags test
// @Accept json
// @Produce json
// @Success 200 {object} res.TestSuccessRes "成功响应校验响应"
// @Router /api/v1/test/success [get]
func (u *TestHandler) TestSuccess(c *gin.Context) {
	data, err := u.testService.TestSuccess(c)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, response.NewErrorResponse[any](err))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse[res.TestSuccessRes](data))
}

// TestError 错误响应校验
// @Summary 错误响应校验
// @Description 错误响应校验
// @Tags test
// @Accept json
// @Produce json
// @Success 200 {object} res.TestErrorRes "错误响应校验响应"
// @Router /api/v1/test/error [get]
func (u *TestHandler) TestError(c *gin.Context) {
	data, err := u.testService.TestError(c)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, response.NewErrorResponse[any](err))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse[res.TestErrorRes](data))
}

// TestLong 长耗时请求模拟
// @Summary 长耗时请求模拟
// @Description 长耗时请求模拟
// @Tags test
// @Accept json
// @Produce json
// @Param req body request.TestLongReq true "长耗时请求模拟请求"
// @Success 200 {object} res.TestLongRes "长耗时请求模拟响应"
// @Router /api/v1/test/long [post]
func (u *TestHandler) TestLong(c *gin.Context) {
	var req request.TestLongReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse[any](errorutil.ParamsError.WithMessage(err.Error())))
		return
	}

	data, err := u.testService.TestLong(c, req)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, response.NewErrorResponse[any](err))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse[res.TestLongRes](data))
}
