package request
{{if .HasTimeField}}

import "time"
{{end}}
{{range $i, $s := .Structs}}{{if $i}}

{{end}}{{if $s.Description}}// {{$s.Description}}
{{end}}type {{$s.StructName}} struct {
{{range $s.Fields}}	{{.Name}} {{.Type}} `json:"{{.JsonTag}}" form:"{{.JsonTag}}"{{if .IsPathParams}} json:"-"{{end}}{{if .ValidateTag}} validate:"{{.ValidateTag}}"{{end}}`{{if .Comment}} // {{.Comment}}{{end}}
{{end}}}
{{- end}}
