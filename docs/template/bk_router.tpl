//router 解析
package {{.PackageName}}

import (
	"github.com/gin-gonic/gin"
	"{{.Module}}/internal/model/common"
	"{{.Module}}/internal/utils"
)

type {{.ModelName}}Router struct{}

func (s *{{.ModelName}}Router) Init{{.ModelName}}Router(Router *gin.RouterGroup, apiSet map[string]struct{}) (R gin.IRoutes) {
	{{.FileName}}Router := Router.Group("")
	utils.RegisterApi({{.FileName}}Router, apiSet, "",
{{- range .Apis}}
		utils.NewRegisterApiParam(common.Http{{title .Method}}, "{{.Path}}", "{{.Description}}", {{.HandlerVar}}.{{.MethodName}}),
{{- end}}
	)
	return {{.FileName}}Router
}
