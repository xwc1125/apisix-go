// Package plugins
//
// @author: xwc1125
package plugins

import (
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/valyala/fasthttp"
)

var (
	pluginRegistry = pluginRegistries{opts: make(map[string]Plugin)}

	ErrMissingName                 = errors.New("missing name")
	ErrMissingParseConfMethod      = errors.New("missing ParseConf method")
	ErrMissingRequestFilterMethod  = errors.New("missing RequestFilter method")
	ErrMissingResponseFilterMethod = errors.New("missing ResponseFilter method")

	RequestPhase  = requestPhase{}  // 请求阶段
	ResponsePhase = responsePhase{} // 响应阶段
)

type ErrPluginRegistered struct {
	name string
}

func (err ErrPluginRegistered) Error() string {
	return fmt.Sprintf("plugin %s registered", err.name)
}

type pluginRegistries struct {
	sync.Mutex
	opts map[string]Plugin
}

func RegisterPlugin(plugin Plugin) error {
	log().Info("register plugin", "name", plugin.Name(), "version", plugin.Version(), "priority", plugin.Priority())

	if plugin.Name() == "" {
		return ErrMissingName
	}

	pluginRegistry.Lock()
	defer pluginRegistry.Unlock()
	if _, found := pluginRegistry.opts[plugin.Name()]; found {
		return ErrPluginRegistered{plugin.Name()}
	}
	pluginRegistry.opts[plugin.Name()] = plugin
	return nil
}

func findPlugin(name string) Plugin {
	if opt, found := pluginRegistry.opts[name]; found {
		return opt
	}
	return nil
}

func getPluginRuntimes(conf RuleConf) []pluginRuntime {
	plugins := Plugins{}
	for _, c := range conf {
		plugin := findPlugin(c.Name)
		if plugin == nil {
			log().Warn("can't find plugin, skip", "name", c.Name)
			continue
		}

		plugins = append(plugins, pluginRuntime{
			conf:   c,
			plugin: plugin,
		})
	}
	sort.Sort(plugins)
	return plugins
}

type requestPhase struct {
}

func (ph *requestPhase) filter(conf RuleConf, req *fasthttp.Request, resp *fasthttp.Response) error {
	pluginRuntimes := getPluginRuntimes(conf)
	for _, pluginRuntime := range pluginRuntimes {
		log().Debug("request run plugin", "plugin", pluginRuntime.conf.Name)
		err := pluginRuntime.plugin.RequestFilter(pluginRuntime.conf.Value, req, resp)
		if err != nil {
			log().Error("plugin run request filter err", "plugin", pluginRuntime.conf.Name, "err", err)
			return err
		}
		if resp.StatusCode() != fasthttp.StatusOK {
			log().Error("plugin run request filter break", "plugin", pluginRuntime.conf.Name, "statusCode", resp.StatusCode())
			break
		}
	}
	return nil
}

// HTTPReqCall http请求的调用
func HTTPReqCall(key string, req *fasthttp.Request, resp *fasthttp.Response) error {
	conf, err := GetRuleConf(key)
	if err != nil {
		return err
	}
	// 请求阶段
	return RequestPhase.filter(conf, req, resp)
}

type responsePhase struct {
}

func (ph *responsePhase) filter(conf RuleConf, w *fasthttp.Response) error {
	pluginRuntimes := getPluginRuntimes(conf)
	for _, pluginRuntime := range pluginRuntimes {
		err := pluginRuntime.plugin.ResponseFilter(pluginRuntime.conf.Value, w)
		if err != nil {
			log().Error("plugin run response filter err", "plugin", pluginRuntime.conf.Name, "statusCode", w.StatusCode(), "err", err)
			return err
		}
		if w.StatusCode() != fasthttp.StatusOK {
			log().Error("plugin run response filter break", "plugin", pluginRuntime.conf.Name, "statusCode", w.StatusCode())
			break
		}
	}

	return nil
}

// HTTPRespCall http 响应的调用
func HTTPRespCall(key string, resp *fasthttp.Response) error {
	conf, err := GetRuleConf(key)
	if err != nil {
		return err
	}

	err = ResponsePhase.filter(conf, resp)
	if err != nil {
		return err
	}

	return nil
}
