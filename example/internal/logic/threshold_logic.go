package logic

import (
	"strings"

	"codegen-example/internal/model/gen/request"
	"codegen-example/internal/store"
	"codegen-example/internal/validate"

	"github.com/gin-gonic/gin"
)

// ThresholdService 阈值业务逻辑(伪实现)
type ThresholdService struct {
	thresholds *store.ThresholdStore
}

// NewThresholdService 创建阈值服务
func NewThresholdService() *ThresholdService {
	return &ThresholdService{thresholds: store.NewThresholdStore()}
}

// allowedScores 允许的阈值档位
var allowedScores = []string{"easy", "normal", "hard"}

// CreateThreshold 新增阈值
func (s *ThresholdService) CreateThreshold(c *gin.Context, req request.CreateThresholdReq) (int64, error) {
	if err := validate.Required("name", req.Name); err != nil {
		return 0, err
	}
	if err := validate.Required("score", req.Score); err != nil {
		return 0, err
	}
	if err := validate.OneOf("score", req.Score, allowedScores...); err != nil {
		return 0, err
	}
	if len([]rune(req.Name)) > 64 {
		return 0, validate.Wrap(validate.Invalid, "name 最多 64 个字符")
	}
	for _, t := range s.thresholds.Filter(nil) {
		if strings.EqualFold(t.Score, req.Score) {
			return 0, validate.Wrap(validate.Existed, "该档位已存在阈值: "+req.Score)
		}
	}

	id := s.thresholds.NextId()
	s.thresholds.Put(id, store.Threshold{Id: id, Name: req.Name, Score: req.Score})
	return id, nil
}

// UpdateThreshold 更新阈值。只允许改名称
func (s *ThresholdService) UpdateThreshold(c *gin.Context, id int64, req request.UpdateThresholdReq) (bool, error) {
	if err := validate.RequiredID("id", id); err != nil {
		return false, err
	}
	if err := validate.Required("name", req.Name); err != nil {
		return false, err
	}
	if _, ok := s.thresholds.Get(id); !ok {
		return false, validate.NotFoundErr("阈值", id)
	}
	if len([]rune(req.Name)) > 64 {
		return false, validate.Wrap(validate.Invalid, "name 最多 64 个字符")
	}

	_, ok := s.thresholds.Update(id, func(t store.Threshold) store.Threshold {
		t.Name = req.Name
		return t
	})
	return ok, nil
}
