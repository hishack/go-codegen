package logic

import (
	"codegen-example/internal/model/gen/request"
	res "codegen-example/internal/model/gen/response"

	"github.com/gin-gonic/gin"
)

type TestService struct{}

func NewTestService() *TestService {
	return &TestService{}
}

// TestPing ping 测试
func (s *TestService) TestPing(c *gin.Context) (res.TestPingRes, error) {
	return res.TestPingRes{}, nil
}

// TestHello hello 测试
func (s *TestService) TestHello(c *gin.Context) (res.TestHelloRes, error) {
	return res.TestHelloRes{}, nil
}

// TestLogger 日志组件测试
func (s *TestService) TestLogger(c *gin.Context) (res.TestLoggerRes, error) {
	return res.TestLoggerRes{}, nil
}

// TestRedis redis 读写测试
func (s *TestService) TestRedis(c *gin.Context, req request.TestRedisReq) (res.TestRedisRes, error) {
	return res.TestRedisRes{}, nil
}

// TestSuccess 成功响应校验
func (s *TestService) TestSuccess(c *gin.Context) (res.TestSuccessRes, error) {
	return res.TestSuccessRes{}, nil
}

// TestError 错误响应校验
func (s *TestService) TestError(c *gin.Context) (res.TestErrorRes, error) {
	return res.TestErrorRes{}, nil
}

// TestLong 长耗时请求模拟
func (s *TestService) TestLong(c *gin.Context, req request.TestLongReq) (res.TestLongRes, error) {
	return res.TestLongRes{}, nil
}
