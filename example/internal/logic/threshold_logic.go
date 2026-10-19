package logic

import (
	"codegen-example/internal/model/gen/request"

	"github.com/gin-gonic/gin"
)

type ThresholdService struct{}

func NewThresholdService() *ThresholdService {
	return &ThresholdService{}
}

// CreateThreshold 新增阈值:创建一个成长阈值,例如"等级达到 10 需要 1000 经验"
func (s *ThresholdService) CreateThreshold(c *gin.Context, req request.CreateThresholdReq) (int64, error) {
	// TODO: 实现业务逻辑:新增阈值:创建一个成长阈值,例如"等级达到 10 需要 1000 经验"
	return 0, nil
}

// UpdateThreshold 更新阈值:按主键局部更新阈值配置
func (s *ThresholdService) UpdateThreshold(c *gin.Context, id int64, req request.UpdateThresholdReq) (bool, error) {
	// TODO: 实现业务逻辑:更新阈值:按主键局部更新阈值配置
	return false, nil
}
