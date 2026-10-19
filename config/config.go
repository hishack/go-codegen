package config

import (
	"errors"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// CodeGenConfig 代码生成器配置。
// 输出目录指向"目标项目"(被生成脚手架的工程),import 路径由 Module 拼出
type CodeGenConfig struct {
	YamlInPath string // 输入的 yaml 目录
	Module     string // 目标项目的 go module 名,如 "shack",用于生成 import 路径
	Group      string // 路由分组前缀,如 "/api/v1";生成路由、swagger 与前端 url 时统一加上

	ServiceIntOutPath string // service 接口目录,如 {目标项目}/internal/service
	ModelOutPath      string // 模型父目录,其下自动分 request/ response/ 子包,如 {目标项目}/internal/model/gen
	LogicOutPath      string // logic 空实现目录,如 {目标项目}/internal/logic
	HandlerOutPath    string // handler 目录,如 {目标项目}/internal/handler
	RouterOutPath     string // 路由目录,如 {目标项目}/internal/router
	WebApiOutPath     string // 前端 api 目录(可选,留空跳过),如 {目标项目}/web/src/api

	UserName string // 写入生成文件头的作者
	Github   string // 写入生成文件头的 github
}

// LoadCodeGenConfig 从 yaml 配置文件加载
func LoadCodeGenConfig(path string) (*CodeGenConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.New("读取配置失败: " + err.Error())
	}
	var raw struct {
		Yaml       string `yaml:"yaml"`
		Module     string `yaml:"module"`
		Group      string `yaml:"group"`
		ServiceOut string `yaml:"serviceOut"`
		ModelOut   string `yaml:"modelOut"`
		LogicOut   string `yaml:"logicOut"`
		HandlerOut string `yaml:"handlerOut"`
		RouterOut  string `yaml:"routerOut"`
		WebOut     string `yaml:"webOut"`
		Author     string `yaml:"author"`
		Github     string `yaml:"github"`
	}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, errors.New("配置文件格式错误: " + err.Error())
	}
	return NewCodeGenConfig(CodeGenConfig{
		YamlInPath:        raw.Yaml,
		Module:            raw.Module,
		Group:             raw.Group,
		ServiceIntOutPath: raw.ServiceOut,
		ModelOutPath:      raw.ModelOut,
		LogicOutPath:      raw.LogicOut,
		HandlerOutPath:    raw.HandlerOut,
		RouterOutPath:     raw.RouterOut,
		WebApiOutPath:     raw.WebOut,
		UserName:          raw.Author,
		Github:            raw.Github,
	})
}

// NewCodeGenConfig 校验并构造配置
func NewCodeGenConfig(cfg CodeGenConfig) (*CodeGenConfig, error) {
	if cfg.YamlInPath == "" || cfg.Module == "" ||
		cfg.ServiceIntOutPath == "" || cfg.ModelOutPath == "" ||
		cfg.LogicOutPath == "" || cfg.HandlerOutPath == "" || cfg.RouterOutPath == "" {
		return nil, errors.New("config: yaml/module/serviceOut/modelOut/logicOut/handlerOut/routerOut 均不能为空")
	}
	return &CodeGenConfig{
		YamlInPath:        cfg.YamlInPath,
		Module:            cfg.Module,
		Group:             NormalizeGroup(cfg.Group),
		ServiceIntOutPath: cfg.ServiceIntOutPath,
		ModelOutPath:      cfg.ModelOutPath,
		LogicOutPath:      cfg.LogicOutPath,
		HandlerOutPath:    cfg.HandlerOutPath,
		RouterOutPath:     cfg.RouterOutPath,
		WebApiOutPath:     cfg.WebApiOutPath,
		UserName:          cfg.UserName,
		Github:            cfg.Github,
	}, nil
}

// NormalizeGroup 规整分组前缀:补齐前导斜杠、去掉末尾斜杠。
// 空值表示不加前缀,返回空串
func NormalizeGroup(g string) string {
	g = strings.TrimSpace(g)
	if g == "" || g == "/" {
		return ""
	}
	if !strings.HasPrefix(g, "/") {
		g = "/" + g
	}
	return strings.TrimRight(g, "/")
}

// PrefixPath 给路径加上分组前缀。g 为空时原样返回
func (c *CodeGenConfig) PrefixPath(p string) string {
	if c == nil || c.Group == "" {
		return p
	}
	if p == "" || p == "/" {
		return c.Group
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return c.Group + p
}
