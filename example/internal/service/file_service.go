package service

import (
	"codegen-example/internal/model/gen/request"
	res "codegen-example/internal/model/gen/response"

	"github.com/gin-gonic/gin"
)

type IFileService interface {
	UploadFile(c *gin.Context, req request.UploadFileReq) (string, error)
	UploadFiles(c *gin.Context, req request.UploadFilesReq) (res.UploadFilesRes, error)
	GetFileInfo(c *gin.Context, id int64) (res.GetFileInfoRes, error)
}
