package logic

import (
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"codegen-example/internal/model/gen/request"
	res "codegen-example/internal/model/gen/response"
	"codegen-example/internal/store"
	"codegen-example/internal/validate"

	"github.com/gin-gonic/gin"
)

// FileService 文件业务逻辑(伪实现:只登记元信息,不落盘)
type FileService struct {
	files *store.FileStore
}

// NewFileService 创建文件服务
func NewFileService() *FileService {
	return &FileService{files: store.NewFileStore()}
}

// 上传限制
const (
	maxUploadSize  = 10 << 20 // 单文件最大 10MB
	maxUploadCount = 9        // 单次最多 9 个文件
)

// allowedExts 允许的扩展名
var allowedExts = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".webp": true, ".gif": true,
	".pdf": true, ".txt": true, ".md": true, ".csv": true, ".xlsx": true,
}

// UploadFile 上传单个文件
func (s *FileService) UploadFile(c *gin.Context, req request.UploadFileReq) (string, error) {
	if req.File == nil {
		return "", validate.Wrap(validate.EmptyField, "缺少上传文件")
	}
	if err := checkFile(req.File, maxUploadSize); err != nil {
		return "", err
	}

	id := s.files.NextId()
	url := fmt.Sprintf("/static/upload/%d%s", id, strings.ToLower(filepath.Ext(req.File.Filename)))
	s.files.Put(id, store.FileInfo{
		Id: id, Name: filepath.Base(req.File.Filename),
		Size: req.File.Size, Url: url, UploadAt: time.Now(),
	})
	return url, nil
}

// UploadFiles 批量上传。任一失败则整批失败
func (s *FileService) UploadFiles(c *gin.Context, req request.UploadFilesReq) (res.UploadFilesRes, error) {
	if err := validate.NonEmpty("files", len(req.Files)); err != nil {
		return res.UploadFilesRes{}, err
	}
	if len(req.Files) > maxUploadCount {
		return res.UploadFilesRes{}, validate.Wrap(validate.Invalid,
			fmt.Sprintf("files 最多 %d 个", maxUploadCount))
	}

	urls := make([]string, 0, len(req.Files))
	for i, item := range req.Files {
		if item == nil || item.File == nil {
			return res.UploadFilesRes{}, validate.Wrap(validate.EmptyField,
				fmt.Sprintf("files[%d] 缺少上传文件", i))
		}
		if strings.TrimSpace(item.Key) == "" {
			return res.UploadFilesRes{}, validate.Wrap(validate.EmptyField,
				fmt.Sprintf("files[%d].key 不能为空", i))
		}
		if err := checkFile(item.File, maxUploadSize); err != nil {
			return res.UploadFilesRes{}, validate.Wrap(validate.Invalid,
				fmt.Sprintf("files[%d](%s): %s", i, item.Key, err.Error()))
		}

		id := s.files.NextId()
		url := fmt.Sprintf("/static/upload/%d%s", id, strings.ToLower(filepath.Ext(item.File.Filename)))
		s.files.Put(id, store.FileInfo{
			Id: id, Name: filepath.Base(item.File.Filename),
			Size: item.File.Size, Url: url, UploadAt: time.Now(),
		})
		urls = append(urls, url)
	}
	return res.UploadFilesRes{Urls: urls}, nil
}

// GetFileInfo 查询文件元信息
func (s *FileService) GetFileInfo(c *gin.Context, id int64) (res.GetFileInfoRes, error) {
	if err := validate.RequiredID("id", id); err != nil {
		return res.GetFileInfoRes{}, err
	}
	f, ok := s.files.Get(id)
	if !ok {
		return res.GetFileInfoRes{}, validate.NotFoundErr("文件", id)
	}
	return res.GetFileInfoRes{
		Id: f.Id, Name: f.Name, Size: f.Size, Url: f.Url, UploadAt: f.UploadAt,
	}, nil
}

// checkFile 校验文件大小与扩展名白名单
func checkFile(f *multipart.FileHeader, max int64) error {
	if f.Size == 0 {
		return validate.Wrap(validate.EmptyField, "文件为空")
	}
	if f.Size > max {
		return validate.Wrap(validate.TooLarge,
			fmt.Sprintf("%s 大小 %d,上限 %d", f.Filename, f.Size, max))
	}
	ext := strings.ToLower(filepath.Ext(f.Filename))
	if !allowedExts[ext] {
		return validate.Wrap(validate.Invalid,
			fmt.Sprintf("%s 不支持的扩展名 %q", f.Filename, ext))
	}
	return nil
}
