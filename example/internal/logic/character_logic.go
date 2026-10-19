package logic

import (
	res "codegen-example/internal/model/gen/response"

	"github.com/gin-gonic/gin"
)

type CharacterService struct{}

func NewCharacterService() *CharacterService {
	return &CharacterService{}
}

// GetCharacterSkills 查询角色技能列表
func (s *CharacterService) GetCharacterSkills(c *gin.Context, id int64) (res.GetCharacterSkillsRes, error) {
	// TODO: 实现业务逻辑:查询角色技能列表
	return res.GetCharacterSkillsRes{}, nil
}
