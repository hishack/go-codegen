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

type FileHandler struct {
	fileService service.IFileService
}

func NewFileHandler(userService service.IFileService) *FileHandler {
	return &FileHandler{
		fileService: userService,
	}
}

// UploadFile 单文件上传
// @Summary 单文件上传
// @Description 单文件上传
// @Tags system
// @Accept json
// @Produce json
// @Param file formData file true "上传的文件"
// @Success 200 {object} string "单文件上传响应"
// @Router /api/v1/files [post]
func (u *FileHandler) UploadFile(c *gin.Context) {
	var req request.UploadFileReq
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse[any](errorutil.ParamsError.WithMessage(err.Error())))
		return
	}

	data, err := u.fileService.UploadFile(c, req)
	if err != nil {
		c.JSON(errorutil.HTTPStatusOf(err), response.NewErrorResponse[any](err))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse[string](data))
}

// UploadFiles 多文件上传
// @Summary 多文件上传
// @Description 多文件上传
// @Tags system
// @Accept json
// @Produce json
// @Param req body request.UploadFilesReq true "多文件上传请求"
// @Success 200 {object} res.UploadFilesRes "多文件上传响应"
// @Router /api/v1/files/batch [post]
func (u *FileHandler) UploadFiles(c *gin.Context) {
	var req request.UploadFilesReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse[any](errorutil.ParamsError.WithMessage(err.Error())))
		return
	}

	data, err := u.fileService.UploadFiles(c, req)
	if err != nil {
		c.JSON(errorutil.HTTPStatusOf(err), response.NewErrorResponse[any](err))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse[res.UploadFilesRes](data))
}

// GetFileInfo 文件信息
// @Summary 文件信息
// @Description 文件信息
// @Tags system
// @Accept json
// @Produce json
// @Param id path integer true "id" format(int64)
// @Success 200 {object} res.GetFileInfoRes "文件信息响应"
// @Router /api/v1/files/:id [get]
func (u *FileHandler) GetFileInfo(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse[any](errorutil.ParamsError.WithMessage(err.Error())))
		return
	}

	data, err := u.fileService.GetFileInfo(c, id)
	if err != nil {
		c.JSON(errorutil.HTTPStatusOf(err), response.NewErrorResponse[any](err))
		return
	}
	c.JSON(http.StatusOK, response.NewSuccessResponse[res.GetFileInfoRes](data))
}
