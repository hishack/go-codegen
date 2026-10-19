package handler

import (
	"net/http"
	"strconv"

	"codegen-example/internal/model/gen/request"
	res "codegen-example/internal/model/gen/response"
	"codegen-example/internal/service"
	"codegen-example/pkg/errorutil"
	"codegen-example/pkg/response"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService service.IUserService
}

func NewUserHandler(userService service.IUserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// CreateUser 基础增删改查 + 全量基础类型覆盖
// @Summary 基础增删改查 + 全量基础类型覆盖
// @Description 基础增删改查 + 全量基础类型覆盖
// @Tags system
// @Accept json
// @Produce json
// @Param req body request.CreateUserReq true "基础增删改查 + 全量基础类型覆盖请求"
// @Success 200 {object} res.CreateUserRes "基础增删改查 + 全量基础类型覆盖响应"
// @Router /api/v1/users [post]
func (u *UserHandler) CreateUser(c *gin.Context) {
	var req request.CreateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse[any](errorutil.ParamsError.WithMessage(err.Error())))
		return
	}

	data, err := u.userService.CreateUser(c, req)
	if err != nil {
		c.JSON(errorutil.HTTPStatusOf(err), response.NewErrorResponse[any](err))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse[res.CreateUserRes](data))
}

// GetUser GetUser
// @Summary GetUser
// @Description GetUser
// @Tags system
// @Accept json
// @Produce json
// @Param userId path integer true "userId" format(int64)
// @Success 200 {object} res.GetUserRes "GetUser响应"
// @Router /api/v1/users/:userId [get]
func (u *UserHandler) GetUser(c *gin.Context) {
	userIdStr := c.Param("userId")
	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse[any](errorutil.ParamsError.WithMessage(err.Error())))
		return
	}

	data, err := u.userService.GetUser(c, userId)
	if err != nil {
		c.JSON(errorutil.HTTPStatusOf(err), response.NewErrorResponse[any](err))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse[res.GetUserRes](data))
}

// ListUsers ListUsers
// @Summary ListUsers
// @Description ListUsers
// @Tags system
// @Accept json
// @Produce json
// @Success 200 {object} res.ListUsersRes "ListUsers响应"
// @Router /api/v1/users [get]
func (u *UserHandler) ListUsers(c *gin.Context) {
	data, err := u.userService.ListUsers(c)
	if err != nil {
		c.JSON(errorutil.HTTPStatusOf(err), response.NewErrorResponse[any](err))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse[res.ListUsersRes](data))
}

// UpdateUser UpdateUser
// @Summary UpdateUser
// @Description UpdateUser
// @Tags system
// @Accept json
// @Produce json
// @Param userId path integer true "userId" format(int64)
// @Param req body request.UpdateUserReq true "UpdateUser请求"
// @Success 200 {object} bool "UpdateUser响应"
// @Router /api/v1/users/:userId [put]
func (u *UserHandler) UpdateUser(c *gin.Context) {
	var req request.UpdateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse[any](errorutil.ParamsError.WithMessage(err.Error())))
		return
	}

	userIdStr := c.Param("userId")
	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse[any](errorutil.ParamsError.WithMessage(err.Error())))
		return
	}

	data, err := u.userService.UpdateUser(c, userId, req)
	if err != nil {
		c.JSON(errorutil.HTTPStatusOf(err), response.NewErrorResponse[any](err))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse[bool](data))
}

// PatchUser PatchUser
// @Summary PatchUser
// @Description PatchUser
// @Tags system
// @Accept json
// @Produce json
// @Param userId path integer true "userId" format(int64)
// @Param req body request.PatchUserReq true "PatchUser请求"
// @Success 200 {object} bool "PatchUser响应"
// @Router /api/v1/users/:userId [patch]
func (u *UserHandler) PatchUser(c *gin.Context) {
	var req request.PatchUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse[any](errorutil.ParamsError.WithMessage(err.Error())))
		return
	}

	userIdStr := c.Param("userId")
	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse[any](errorutil.ParamsError.WithMessage(err.Error())))
		return
	}

	data, err := u.userService.PatchUser(c, userId, req)
	if err != nil {
		c.JSON(errorutil.HTTPStatusOf(err), response.NewErrorResponse[any](err))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse[bool](data))
}

// DeleteUser DeleteUser
// @Summary DeleteUser
// @Description DeleteUser
// @Tags system
// @Accept json
// @Produce json
// @Param userId path integer true "userId" format(int64)
// @Success 200 {object} bool "DeleteUser响应"
// @Router /api/v1/users/:userId [delete]
func (u *UserHandler) DeleteUser(c *gin.Context) {
	userIdStr := c.Param("userId")
	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse[any](errorutil.ParamsError.WithMessage(err.Error())))
		return
	}

	data, err := u.userService.DeleteUser(c, userId)
	if err != nil {
		c.JSON(errorutil.HTTPStatusOf(err), response.NewErrorResponse[any](err))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse[bool](data))
}

// GetUserProfileVo 根据ID获取用户详情
// @Summary 根据ID获取用户详情
// @Description 根据ID获取用户详情
// @Tags system
// @Accept json
// @Produce json
// @Param id path integer true "id" format(int64)
// @Success 200 {object} res.GetUserProfileVoRes "根据ID获取用户详情响应"
// @Router /api/v1/users/profile/:id [get]
func (u *UserHandler) GetUserProfileVo(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse[any](errorutil.ParamsError.WithMessage(err.Error())))
		return
	}

	data, err := u.userService.GetUserProfileVo(c, id)
	if err != nil {
		c.JSON(errorutil.HTTPStatusOf(err), response.NewErrorResponse[any](err))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse[res.GetUserProfileVoRes](data))
}

// UploadAvatar 上传头像(multipart 文件上传)
// @Summary 上传头像(multipart 文件上传)
// @Description 上传头像(multipart 文件上传)
// @Tags system
// @Accept json
// @Produce json
// @Param avatar formData file true "头像文件"
// @Success 200 {object} string "上传头像(multipart 文件上传)响应"
// @Router /api/v1/users/avatar [post]
func (u *UserHandler) UploadAvatar(c *gin.Context) {
	var req request.UploadAvatarReq
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse[any](errorutil.ParamsError.WithMessage(err.Error())))
		return
	}

	data, err := u.userService.UploadAvatar(c, req)
	if err != nil {
		c.JSON(errorutil.HTTPStatusOf(err), response.NewErrorResponse[any](err))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse[string](data))
}
