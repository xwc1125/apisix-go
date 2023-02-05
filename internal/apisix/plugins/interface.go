// Package plugins
//
// @author: xwc1125
package plugins

import "github.com/valyala/fasthttp"

// Plugin 插件接口
type Plugin interface {
	// Name 插件名称
	Name() string
	// Version 版本信息
	Version() string
	// Priority 优先级
	Priority() int64

	// ParseConf 解析插件配置
	// 如果无法解析，那么跳过改插件
	ParseConf(in []byte) (conf interface{}, err error)

	// RequestFilter 根据conf对象进行request的处理
	// 当err不为nil时，代表执行出错，那么将会直接返回错误。
	// 当w被改写时，即w.StatusCode!=fasthttp.StatusOK时，会跳出插件链的执行
	RequestFilter(conf interface{}, r *fasthttp.Request, w *fasthttp.Response) (err error)

	// ResponseFilter 对相应结果的处理
	ResponseFilter(conf interface{}, w *fasthttp.Response) (err error)
}

type pluginRuntime struct {
	conf   ConfEntry
	plugin Plugin
}
type Plugins []pluginRuntime

func (r Plugins) Len() int {
	return len(r)
}

func (r Plugins) Less(i, j int) bool {
	if r[i].plugin.Priority() < r[j].plugin.Priority() {
		return true
	}
	return false
}

func (r Plugins) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

// DefaultPlugin 插件接口的无操作实现
type DefaultPlugin struct{}

func (*DefaultPlugin) RequestFilter(conf interface{}, r *fasthttp.Request, w *fasthttp.Response) error {
	return nil
}
func (*DefaultPlugin) ResponseFilter(conf interface{}, w *fasthttp.Response) error {
	return nil
}
