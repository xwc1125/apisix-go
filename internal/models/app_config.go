// Package models
//
// @author: xwc1125
package models

import "github.com/chain5j/chain5j-pkg/network"

// ApplicationConfig 应用信息
type ApplicationConfig struct {
	Mode       string `json:"mode" mapstructure:"mode"`                // 环境模式(dev,test,prod)
	Name       string `json:"name" mapstructure:"name"`                // 应用名称
	Version    string `json:"version" mapstructure:"version"`          // 应用版本
	BaseSecret string `json:"base_secret" mapstructure:"base_secret"`  // 基础加密秘密
	Debug      bool   `json:"debug" mapstructure:"debug" yaml:"debug"` // 是否开启debug
}

// ServerConfig 服务配置
type ServerConfig struct {
	Host string            `json:"host" mapstructure:"host" yaml:"host"`
	Port int               `json:"port" mapstructure:"port" yaml:"port"`
	Ssl  network.TlsConfig `json:"ssl" mapstructure:"ssl" yaml:"ssl"`
	// Cors cors.Options      `json:"cors" mapstructure:"cors" yaml:"cors"`
}
