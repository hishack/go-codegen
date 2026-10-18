package {{.PackageName}}

import (
	"github.com/gin-gonic/gin"
{{- if .HasRequest}}
	req "{{.Module}}/internal/model/{{.PackageName}}/request"
{{- end}}
{{- if .HasResponse}}
	res "{{.Module}}/internal/model/{{.PackageName}}/response"
{{- end }}
)

type {{$.ServiceName}} struct{}
{{range .Methods}}
// {{.Description}}
// Auth: {{$.Auth}} 
// Github: {{$.Github}}
// Time : {{$.Time}}
func (s *{{$.ServiceName}}) {{.MethodName}}(
	ctx *gin.Context,
{{- if .HasRequest}}
	r req.{{.RequestType}},
{{- end }}
) ({{if .HasResponse}}rs res.{{.ResponseType}}, {{end}}err error) {
	{{- if .HasResponse}}
	return rs, nil
	{{- else}}
	return nil
	{{- end}}
}
{{end}}
