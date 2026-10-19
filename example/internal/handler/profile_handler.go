package handler

import (
	"net/http"

	"codegen-example/internal/model/gen/request"
	res "codegen-example/internal/model/gen/response"
	"codegen-example/internal/service"
	"codegen-example/pkg/response"

	"github.com/gin-gonic/gin"
)

type ProfileHandler struct {
	profileService service.IProfileService
}

func NewProfileHandler(userService service.IProfileService) *ProfileHandler {
	return &ProfileHandler{
		profileService: userService,
	}
}

// GetUserProfile 用户详情
// @Summary 用户详情
// @Description 用户详情
// @Tags system
// @Accept json
// @Produce json
// @Success 200 {object} res.GetUserProfileRes "用户详情响应"
// @Router /users/profile [get]
func (u *ProfileHandler) GetUserProfile(c *gin.Context) {
	data, err := u.profileService.GetUserProfile(c)
	if err != nil {
		c.JSON(http.StatusOK, response.NewErrorResponse[any](err))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse[res.GetUserProfileRes](data))
}

// UpdateUserProfile 更新用户资料
// @Summary 更新用户资料
// @Description 更新用户资料
// @Tags system
// @Accept json
// @Produce json
// @Param req body request.UpdateUserProfileReq true "更新用户资料请求"
// @Success 200 {object} bool "更新用户资料响应"
// @Router /users/profile [put]
func (u *ProfileHandler) UpdateUserProfile(c *gin.Context) {
	var req request.UpdateUserProfileReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, response.NewErrorResponse[any](err))
		return
	}

	data, err := u.profileService.UpdateUserProfile(c, req)
	if err != nil {
		c.JSON(http.StatusOK, response.NewErrorResponse[any](err))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse[bool](data))
}
