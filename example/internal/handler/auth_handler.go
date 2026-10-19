package handler

import (
	"net/http"

	"codegen-example/internal/model/gen/request"
	res "codegen-example/internal/model/gen/response"
	"codegen-example/internal/service"
	"codegen-example/pkg/response"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService service.IAuthService
}

func NewAuthHandler(userService service.IAuthService) *AuthHandler {
	return &AuthHandler{
		authService: userService,
	}
}

// Login 本文件演示:一个 yaml 靠接口级 filename 拆成多个领域。 三个接口分属 user / profile / auth 三组文件,互不干扰。 登录
// @Summary 本文件演示:一个 yaml 靠接口级 filename 拆成多个领域。 三个接口分属 user / profile / auth 三组文件,互不干扰。 登录
// @Description 本文件演示:一个 yaml 靠接口级 filename 拆成多个领域。 三个接口分属 user / profile / auth 三组文件,互不干扰。 登录
// @Tags system
// @Accept json
// @Produce json
// @Param req body request.LoginReq true "本文件演示:一个 yaml 靠接口级 filename 拆成多个领域。 三个接口分属 user / profile / auth 三组文件,互不干扰。 登录请求"
// @Success 200 {object} res.LoginRes "本文件演示:一个 yaml 靠接口级 filename 拆成多个领域。 三个接口分属 user / profile / auth 三组文件,互不干扰。 登录响应"
// @Router /auth/login [post]
func (u *AuthHandler) Login(c *gin.Context) {
	var req request.LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, response.NewErrorResponse[any](err))
		return
	}

	data, err := u.authService.Login(c, req)
	if err != nil {
		c.JSON(http.StatusOK, response.NewErrorResponse[any](err))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse[res.LoginRes](data))
}

// RefreshToken 刷新令牌
// @Summary 刷新令牌
// @Description 刷新令牌
// @Tags system
// @Accept json
// @Produce json
// @Param req body request.RefreshTokenReq true "刷新令牌请求"
// @Success 200 {object} string "刷新令牌响应"
// @Router /auth/refresh [post]
func (u *AuthHandler) RefreshToken(c *gin.Context) {
	var req request.RefreshTokenReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, response.NewErrorResponse[any](err))
		return
	}

	data, err := u.authService.RefreshToken(c, req)
	if err != nil {
		c.JSON(http.StatusOK, response.NewErrorResponse[any](err))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse[string](data))
}

// LogoutAccount 退出登录
// @Summary 退出登录
// @Description 退出登录
// @Tags system
// @Accept json
// @Produce json
// @Param req body request.LogoutAccountReq true "退出登录请求"
// @Success 200 {object} bool "退出登录响应"
// @Router /auth/logout [post]
func (u *AuthHandler) LogoutAccount(c *gin.Context) {
	var req request.LogoutAccountReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, response.NewErrorResponse[any](err))
		return
	}

	err := u.authService.LogoutAccount(c, req)
	if err != nil {
		c.JSON(http.StatusOK, response.NewErrorResponse[any](err))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse[any](true))
}
