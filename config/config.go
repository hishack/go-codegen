package config

import (
	"errors"
	"os"

	"gopkg.in/yaml.v3"
)

// CodeGenConfig 代码生成器配置。
// 输出目录指向"目标项目"(被生成脚手架的工程),import 路径由 Module 拼出
type CodeGenConfig struct {
	YamlInPath string // 输入的 yaml 目录
	Module     string // 目标项目的 go module 名,如 "shack",用于生成 import 路径
	Prefix     string // api 结构体名前缀(可选),如 "YiKou" -> YiKouCreateUserRequest

	ServiceIntOutPath string // service 接口目录,如 {目标项目}/internal/service
	ApiOutPath        string // 请求结构体+响应别名目录,如 {目标项目}/internal/api
	ModelOutPath      string // 响应数据结构(vo)目录,如 {目标项目}/internal/dal/vo
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
		Prefix     string `yaml:"prefix"`
		ServiceOut string `yaml:"serviceOut"`
		ApiOut     string `yaml:"apiOut"`
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
		Prefix:            raw.Prefix,
		ServiceIntOutPath: raw.ServiceOut,
		ApiOutPath:        raw.ApiOut,
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
		cfg.ServiceIntOutPath == "" || cfg.ApiOutPath == "" ||
		cfg.ModelOutPath == "" || cfg.LogicOutPath == "" ||
		cfg.HandlerOutPath == "" || cfg.RouterOutPath == "" {
		return nil, errors.New("config: yaml/module/serviceOut/apiOut/modelOut/logicOut/handlerOut/routerOut 均不能为空")
	}
	return &CodeGenConfig{
		YamlInPath:        cfg.YamlInPath,
		Module:            cfg.Module,
		Prefix:            cfg.Prefix,
		ServiceIntOutPath: cfg.ServiceIntOutPath,
		ApiOutPath:        cfg.ApiOutPath,
		ModelOutPath:      cfg.ModelOutPath,
		LogicOutPath:      cfg.LogicOutPath,
		HandlerOutPath:    cfg.HandlerOutPath,
		RouterOutPath:     cfg.RouterOutPath,
		WebApiOutPath:     cfg.WebApiOutPath,
		UserName:          cfg.UserName,
		Github:            cfg.Github,
	}, nil
}
