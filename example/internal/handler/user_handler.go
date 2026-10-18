package handler

import (
	"context"
	"strconv"

	"codegen-example/internal/api"
	"codegen-example/internal/dal/vo"
	"codegen-example/internal/service"
	"codegen-example/pkg/errorutil"
	"codegen-example/pkg/response"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
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
// @Param req body api.CreateUserRequest true "基础增删改查 + 全量基础类型覆盖请求"
// @Success 200 {object} api.CreateUserResponse "基础增删改查 + 全量基础类型覆盖响应"
// @Router /users [post]
func (u *UserHandler) CreateUser(ctx context.Context, c *app.RequestContext) {
	req := &api.CreateUserRequest{}
	err := c.BindAndValidate(req)
	if err != nil {
		c.JSON(consts.StatusOK, response.NewErrorResponse[any](err))
		return
	}

	res, err := u.userService.CreateUser(ctx, req)
	if err != nil {
		c.JSON(consts.StatusOK, response.NewErrorResponse[any](err))
		return
	}
	c.JSON(consts.StatusOK, response.NewSuccessResponse[vo.CreateUserVo](res))
}

// GetUser GetUser
// @Summary GetUser
// @Description GetUser
// @Tags system
// @Accept json
// @Produce json
// @Param userId path integer true "userId"
// @Success 200 {object} api.GetUserResponse "GetUser响应"
// @Router /users/:userId [get]
func (u *UserHandler) GetUser(ctx context.Context, c *app.RequestContext) {
	userIdStr := c.Param("userId")
	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		c.JSON(consts.StatusOK, response.NewErrorResponse[any](errorutil.ParamsError))
	}

	res, err := u.userService.GetUser(ctx, userId)
	if err != nil {
		c.JSON(consts.StatusOK, response.NewErrorResponse[any](err))
		return
	}
	c.JSON(consts.StatusOK, response.NewSuccessResponse[vo.GetUserVo](res))
}

// ListUsers ListUsers
// @Summary ListUsers
// @Description ListUsers
// @Tags system
// @Accept json
// @Produce json
// @Success 200 {object} api.ListUsersResponse "ListUsers响应"
// @Router /users [get]
func (u *UserHandler) ListUsers(ctx context.Context, c *app.RequestContext) {
	res, err := u.userService.ListUsers(ctx)
	if err != nil {
		c.JSON(consts.StatusOK, response.NewErrorResponse[any](err))
		return
	}
	c.JSON(consts.StatusOK, response.NewSuccessResponse[vo.ListUsersVo](res))
}

// UpdateUser UpdateUser
// @Summary UpdateUser
// @Description UpdateUser
// @Tags system
// @Accept json
// @Produce json
// @Param userId path integer true "userId"
// @Param req body api.UpdateUserRequest true "UpdateUser请求"
// @Success 200 {object} api.UpdateUserResponse "UpdateUser响应"
// @Router /users/:userId [put]
func (u *UserHandler) UpdateUser(ctx context.Context, c *app.RequestContext) {
	req := &api.UpdateUserRequest{}
	err := c.BindAndValidate(req)
	if err != nil {
		c.JSON(consts.StatusOK, response.NewErrorResponse[any](err))
		return
	}

	userIdStr := c.Param("userId")
	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		c.JSON(consts.StatusOK, response.NewErrorResponse[any](errorutil.ParamsError))
	}

	res, err := u.userService.UpdateUser(ctx, userId, req)
	if err != nil {
		c.JSON(consts.StatusOK, response.NewErrorResponse[any](err))
		return
	}
	c.JSON(consts.StatusOK, response.NewSuccessResponse[bool](res))
}

// PatchUser PatchUser
// @Summary PatchUser
// @Description PatchUser
// @Tags system
// @Accept json
// @Produce json
// @Param userId path integer true "userId"
// @Param req body api.PatchUserRequest true "PatchUser请求"
// @Success 200 {object} api.PatchUserResponse "PatchUser响应"
// @Router /users/:userId [patch]
func (u *UserHandler) PatchUser(ctx context.Context, c *app.RequestContext) {
	req := &api.PatchUserRequest{}
	err := c.BindAndValidate(req)
	if err != nil {
		c.JSON(consts.StatusOK, response.NewErrorResponse[any](err))
		return
	}

	userIdStr := c.Param("userId")
	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		c.JSON(consts.StatusOK, response.NewErrorResponse[any](errorutil.ParamsError))
	}

	res, err := u.userService.PatchUser(ctx, userId, req)
	if err != nil {
		c.JSON(consts.StatusOK, response.NewErrorResponse[any](err))
		return
	}
	c.JSON(consts.StatusOK, response.NewSuccessResponse[bool](res))
}

// DeleteUser DeleteUser
// @Summary DeleteUser
// @Description DeleteUser
// @Tags system
// @Accept json
// @Produce json
// @Param userId path integer true "userId"
// @Success 200 {object} api.DeleteUserResponse "DeleteUser响应"
// @Router /users/:userId [delete]
func (u *UserHandler) DeleteUser(ctx context.Context, c *app.RequestContext) {
	userIdStr := c.Param("userId")
	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		c.JSON(consts.StatusOK, response.NewErrorResponse[any](errorutil.ParamsError))
	}

	res, err := u.userService.DeleteUser(ctx, userId)
	if err != nil {
		c.JSON(consts.StatusOK, response.NewErrorResponse[any](err))
		return
	}
	c.JSON(consts.StatusOK, response.NewSuccessResponse[bool](res))
}

// GetUserProfileVo 根据ID获取用户详情
// @Summary 根据ID获取用户详情
// @Description 根据ID获取用户详情
// @Tags system
// @Accept json
// @Produce json
// @Param id path integer true "id"
// @Success 200 {object} api.GetUserProfileVoResponse "根据ID获取用户详情响应"
// @Router /users/profile/:id [get]
func (u *UserHandler) GetUserProfileVo(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(consts.StatusOK, response.NewErrorResponse[any](errorutil.ParamsError))
	}

	res, err := u.userService.GetUserProfileVo(ctx, id)
	if err != nil {
		c.JSON(consts.StatusOK, response.NewErrorResponse[any](err))
		return
	}
	c.JSON(consts.StatusOK, response.NewSuccessResponse[vo.GetUserProfileVo](res))
}
