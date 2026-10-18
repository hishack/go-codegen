package response
{{if .HasTimeField}}

import "time"
{{end}}
{{range $i, $s := .Structs}}{{if $i}}

{{end}}{{if $s.Comment}}// {{$s.Comment}}
{{end}}type {{$s.StructName}} struct {
{{range $s.Fields}}	{{.Name}} {{.Type}} `json:"{{.JsonTag}}"`{{if .Comment}} // {{.Comment}}{{end}}
{{end}}}
{{- end}}
