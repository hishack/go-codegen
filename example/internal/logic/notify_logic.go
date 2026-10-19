package logic

import (
	"codegen-example/internal/model/gen/request"
	res "codegen-example/internal/model/gen/response"

	"github.com/gin-gonic/gin"
)

type NotifyService struct{}

func NewNotifyService() *NotifyService {
	return &NotifyService{}
}

// SendSmsCode 发送短信验证码
func (s *NotifyService) SendSmsCode(c *gin.Context, req request.SendSmsCodeReq) (int, error) {
	// TODO: 实现业务逻辑:发送短信验证码
	return 0, nil
}

// ListNotifies 查询通知列表
func (s *NotifyService) ListNotifies(c *gin.Context, req request.ListNotifiesReq) (res.ListNotifiesRes, error) {
	// TODO: 实现业务逻辑:查询通知列表
	return res.ListNotifiesRes{}, nil
}

// ReadNotify 标记通知已读
func (s *NotifyService) ReadNotify(c *gin.Context, id int64) (bool, error) {
	// TODO: 实现业务逻辑:标记通知已读
	return false, nil
}

// GetNotifyConfig 扩展属性示例:map 类型
func (s *NotifyService) GetNotifyConfig(c *gin.Context) (res.GetNotifyConfigRes, error) {
	// TODO: 实现业务逻辑:扩展属性示例:map 类型
	return res.GetNotifyConfigRes{}, nil
}
