package service

import (
	"codegen-example/internal/model/gen/request"

	"github.com/gin-gonic/gin"
)

type IThresholdService interface {
	CreateThreshold(c *gin.Context, req request.CreateThresholdReq) (int64, error)
	UpdateThreshold(c *gin.Context, id int64, req request.UpdateThresholdReq) (bool, error)
}
