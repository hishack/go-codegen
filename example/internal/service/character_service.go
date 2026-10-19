package service

import (
	res "codegen-example/internal/model/gen/response"

	"github.com/gin-gonic/gin"
)

type ICharacterService interface {
	GetCharacterSkills(c *gin.Context, id int64) (res.GetCharacterSkillsRes, error)
}
