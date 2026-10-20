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
	// TODO: 实现业务逻辑:ping 测试
	return res.TestPingRes{}, nil
}

// TestHello hello 测试
func (s *TestService) TestHello(c *gin.Context) (res.TestHelloRes, error) {
	// TODO: 实现业务逻辑:hello 测试
	return res.TestHelloRes{}, nil
}

// TestLogger 日志组件测试
func (s *TestService) TestLogger(c *gin.Context) (res.TestLoggerRes, error) {
	// TODO: 实现业务逻辑:日志组件测试
	return res.TestLoggerRes{}, nil
}

// TestRedis redis 读写测试
func (s *TestService) TestRedis(c *gin.Context, req request.TestRedisReq) (res.TestRedisRes, error) {
	// TODO: 实现业务逻辑:redis 读写测试
	return res.TestRedisRes{}, nil
}

// TestSuccess 成功响应校验
func (s *TestService) TestSuccess(c *gin.Context) (res.TestSuccessRes, error) {
	// TODO: 实现业务逻辑:成功响应校验
	return res.TestSuccessRes{}, nil
}

// TestError 错误响应校验
func (s *TestService) TestError(c *gin.Context) (res.TestErrorRes, error) {
	// TODO: 实现业务逻辑:错误响应校验
	return res.TestErrorRes{}, nil
}

// TestLong 长耗时请求模拟
func (s *TestService) TestLong(c *gin.Context, req request.TestLongReq) (res.TestLongRes, error) {
	// TODO: 实现业务逻辑:长耗时请求模拟
	return res.TestLongRes{}, nil
}
