package {{.PackageName}}

import (

	"github.com/gin-gonic/gin"
{{- if .NeedStrconv}}
	"strconv"
{{- end}}
	req "{{.Module}}/internal/model/{{.PackageName}}/request"
	_ "{{.Module}}/internal/model/{{.PackageName}}/response"
	"{{.Module}}/internal/vo"
	"{{.Module}}/internal/utils/validator"

)

type {{pascal .FileName}}Api struct{}

{{- range .APIs}}

// {{.HandlerName}}
// @Tags {{.PackageName}}{{.ModelName}}Api
// @Summary {{.HandlerName}} {{.Description}}
// @Description {{.HandlerName}} {{.Description}}
{{- if .NeedAuth}}
// @Security ApiKeyAuth
{{- end}}
{{- if .HasRequest}}
// @Param data body req.{{.RequestStruct}} true "请求参数"
{{- end}}
{{- if .HasResponse}}
// @Success 200 {object} vo.Result{data=_.{{.ServiceMethod}}Res}
{{- else}}
// @Success 200 {object} vo.Result{}
{{- end}}
// @Router {{.Path}} [{{.MethodName}}]
func (s *{{pascal .FileName}}Api) {{.HandlerName}}(c *gin.Context) {

{{- if .HasRequest}}
	var req req.{{.RequestStruct}}

{{- if .IsPathParams}}
	// path 参数
{{- range .PathParams}}
	{
		val := c.Param("{{.JsonTag}}")

{{- if eq .Type "int"}}
		parsed, err := strconv.Atoi(val)
		if err != nil {
			c.JSON(201, vo.Fail(c, "", err))
			return
		}
		req.{{.Name}} = parsed

{{- else if eq .Type "int64"}}
		parsed, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			c.JSON(201, vo.Fail(c, "", err))
			return
		}
		req.{{.Name}} = parsed

{{- else if eq .Type "uint"}}
		parsed, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			c.JSON(201, vo.Fail(c, "", err))
			return
		}
		req.{{.Name}} = uint(parsed)

{{- else if eq .Type "uint64"}}
		parsed, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			c.JSON(201, vo.Fail(c, "", err))
			return
		}
		req.{{.Name}} = parsed

{{- else if eq .Type "float64"}}
		parsed, err := strconv.ParseFloat(val, 64)
		if err != nil {
			c.JSON(201, vo.Fail(c, "", err))
			return
		}
		req.{{.Name}} = parsed

{{- else if eq .Type "bool"}}
		parsed, err := strconv.ParseBool(val)
		if err != nil {
			c.JSON(201, vo.Fail(c, "", err))
			return
		}
		req.{{.Name}} = parsed

{{- else}}
		req.{{.Name}} = val
{{- end}}

	}
{{- end}}
{{- end}}

{{- if .IsQueryParams}}
	// query 参数
{{- range .QueryParams}}
	{
		val := c.Query("{{.JsonTag}}")
		if val != "" {

{{- if eq .Type "int"}}
			parsed, err := strconv.Atoi(val)
			if err != nil {
				c.JSON(201, vo.Fail(c, "", err))
				return
			}
			req.{{.Name}} = parsed

{{- else if eq .Type "int64"}}
			parsed, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				c.JSON(201, vo.Fail(c, "", err))
				return
			}
			req.{{.Name}} = parsed

{{- else if eq .Type "uint"}}
			parsed, err := strconv.ParseUint(val, 10, 64)
			if err != nil {
				c.JSON(201, vo.Fail(c, "", err))
				return
			}
			req.{{.Name}} = uint(parsed)

{{- else if eq .Type "uint64"}}
			parsed, err := strconv.ParseUint(val, 10, 64)
			if err != nil {
				c.JSON(201, vo.Fail(c, "", err))
				return
			}
			req.{{.Name}} = parsed

{{- else if eq .Type "float64"}}
			parsed, err := strconv.ParseFloat(val, 64)
			if err != nil {
				c.JSON(201, vo.Fail(c, "", err))
				return
			}
			req.{{.Name}} = parsed

{{- else if eq .Type "bool"}}
			parsed, err := strconv.ParseBool(val)
			if err != nil {
				c.JSON(201, vo.Fail(c, "", err))
				return
			}
			req.{{.Name}} = parsed

{{- else}}
			req.{{.Name}} = val
{{- end}}

		}
	}
{{- end}}
{{- end}}

{{- if .IsBodyParams}}
	// body 参数
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(201, vo.Fail(c, "", err))
		return
	}
{{- end}}

	if err := validator.Validate(&req); err != nil {
		c.JSON(201, vo.Fail(c, "", err))
		return
	}
{{- end}}

{{- if .HasResponse}}
	data, err := {{lowerCamel .FileName}}Service.{{.ServiceMethod}}(c{{if .HasRequest}}, req{{end}})
	if err != nil {
		c.JSON(201, vo.Fail(c, "", err))
		return
	}
	c.JSON(200, vo.Success(c, data))
{{- else}}
	err := {{lowerCamel .FileName}}Service.{{.ServiceMethod}}(c{{if .HasRequest}}, req{{end}})
	if err != nil {
		c.JSON(201, vo.Fail(c, "", err))
		return
	}
	c.JSON(200, vo.Success(c, "操作成功"))
{{- end}}
}
{{- end}}
