package logic

import (
	"codegen-example/internal/model/gen/request"
	res "codegen-example/internal/model/gen/response"

	"github.com/gin-gonic/gin"
)

type FileService struct{}

func NewFileService() *FileService {
	return &FileService{}
}

// UploadFile 单文件上传
func (s *FileService) UploadFile(c *gin.Context, req request.UploadFileReq) (string, error) {
	// TODO: 实现业务逻辑:单文件上传
	return "", nil
}

// UploadFiles 多文件上传
func (s *FileService) UploadFiles(c *gin.Context, req request.UploadFilesReq) (res.UploadFilesRes, error) {
	// TODO: 实现业务逻辑:多文件上传
	return res.UploadFilesRes{}, nil
}

// GetFileInfo 文件信息
func (s *FileService) GetFileInfo(c *gin.Context, id int64) (res.GetFileInfoRes, error) {
	// TODO: 实现业务逻辑:文件信息
	return res.GetFileInfoRes{}, nil
}
