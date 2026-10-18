package config

import "errors"

type CodeGenConfig struct {
	YamlInPath        string // 输入的yaml的路径
	ServiceIntOutPath string // 接口的输出位置
	ApiOutPath        string // api的输出位置
	ModelOutPath      string // model的输出位置,即req,res的位置
	LogicOutPath      string // logic的输出位置
	RouterOutPath     string // router的输出位置
	WebApiOutPath     string // webapi的输出位置
	UserName          string // 用户名
	Github            string // github地址
}

func NewCodeGenConfig(cfg CodeGenConfig) (*CodeGenConfig, error) {

	// 校验
	if cfg.YamlInPath == "" || cfg.ServiceIntOutPath == "" || cfg.ApiOutPath == "" ||
		cfg.ModelOutPath == "" || cfg.LogicOutPath == "" || cfg.RouterOutPath == "" ||
		cfg.WebApiOutPath == "" {
		return nil, errors.New("params is empty")
	}
	return &CodeGenConfig{
		YamlInPath:        cfg.YamlInPath,
		ServiceIntOutPath: cfg.ServiceIntOutPath,
		ApiOutPath:        cfg.ApiOutPath,
		ModelOutPath:      cfg.ModelOutPath,
		LogicOutPath:      cfg.LogicOutPath,
		RouterOutPath:     cfg.RouterOutPath,
		WebApiOutPath:     cfg.WebApiOutPath,
		UserName:          cfg.UserName,
		Github:            cfg.Github,
	}, nil
}
