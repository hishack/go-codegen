package logic

import (
	res "codegen-example/internal/model/gen/response"
	"codegen-example/internal/store"
	"codegen-example/internal/validate"

	"github.com/gin-gonic/gin"
)

// CharacterService 角色业务逻辑(伪实现:技能表存内存)
type CharacterService struct {
	skills *store.SkillStore
}

// NewCharacterService 创建角色服务
func NewCharacterService() *CharacterService {
	return &CharacterService{skills: store.NewSkillStore()}
}

// GetCharacterSkills 查询角色技能列表
func (s *CharacterService) GetCharacterSkills(c *gin.Context, id int64) (res.GetCharacterSkillsRes, error) {
	if err := validate.RequiredID("id", id); err != nil {
		return res.GetCharacterSkillsRes{}, err
	}
	list, ok := s.skills.List(id)
	if !ok {
		return res.GetCharacterSkillsRes{}, validate.NotFoundErr("角色", id)
	}

	out := res.GetCharacterSkillsRes{List: make([]*res.GetCharacterSkillsResListItem, 0, len(list))}
	for _, sk := range list {
		level := sk.Level
		if level == nil {
			level = []int{} // 避免返回 null
		}
		out.List = append(out.List, &res.GetCharacterSkillsResListItem{
			Id: sk.Id, Name: sk.Name, Level: level,
		})
	}
	return out, nil
}
