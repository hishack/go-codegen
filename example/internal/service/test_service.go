package service

import (
	"codegen-example/internal/model/gen/request"
	res "codegen-example/internal/model/gen/response"

	"github.com/gin-gonic/gin"
)

type ITestService interface {
	TestPing(c *gin.Context) (res.TestPingRes, error)
	TestHello(c *gin.Context) (res.TestHelloRes, error)
	TestLogger(c *gin.Context) (res.TestLoggerRes, error)
	TestRedis(c *gin.Context, req request.TestRedisReq) (res.TestRedisRes, error)
	TestSuccess(c *gin.Context) (res.TestSuccessRes, error)
	TestError(c *gin.Context) (res.TestErrorRes, error)
	TestLong(c *gin.Context, req request.TestLongReq) (res.TestLongRes, error)
}
