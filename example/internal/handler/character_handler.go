package handler

import (
	"net/http"
	"strconv"

	res "codegen-example/internal/model/gen/response"
	"codegen-example/internal/service"
	"codegen-example/pkg/errorutil"
	"codegen-example/pkg/response"

	"github.com/gin-gonic/gin"
)

type CharacterHandler struct {
	characterService service.ICharacterService
}

func NewCharacterHandler(userService service.ICharacterService) *CharacterHandler {
	return &CharacterHandler{
		characterService: userService,
	}
}

// GetCharacterSkills 查询角色技能列表
// @Summary 查询角色技能列表
// @Description 查询角色技能列表
// @Tags system
// @Accept json
// @Produce json
// @Param id path integer true "id" format(int64)
// @Success 200 {object} res.GetCharacterSkillsRes "查询角色技能列表响应"
// @Router /api/v1/characters/:id/skills [get]
func (u *CharacterHandler) GetCharacterSkills(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse[any](errorutil.ParamsError.WithMessage(err.Error())))
		return
	}

	data, err := u.characterService.GetCharacterSkills(c, id)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, response.NewErrorResponse[any](err))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse[res.GetCharacterSkillsRes](data))
}
