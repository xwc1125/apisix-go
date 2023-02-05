// Package storage
//
// @author: xwc1125
package storage

type Tls struct {
	CaFile   string `json:"ca" mapstructure:"ca" yaml:"ca"`
	CertFile string `json:"cert" mapstructure:"cert" yaml:"cert"`
	KeyFile  string `json:"key" mapstructure:"key" yaml:"key"`
}

type EtcdConfig struct {
	Endpoints []string `json:"endpoints" mapstructure:"endpoints" yaml:"endpoints"`
	Username  string   `json:"username" mapstructure:"username" yaml:"username"`
	Password  string   `json:"password" mapstructure:"password" yaml:"password"`
	Tls       *Tls     `json:"tls" mapstructure:"tls" yaml:"tls"`
	Prefix    string   `json:"prefix" mapstructure:"prefix" yaml:"prefix"`
}
