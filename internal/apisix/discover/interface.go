// Package discover
//
// @author: xwc1125
package discover

import (
	"errors"
	"fmt"
	"sync"

	"github.com/valyala/fasthttp"
)

var (
	discoverRegistry = discoverRegistries{opts: make(map[string]Discover)}

	ErrMissingName = errors.New("missing name")
)

// Discover 服务发现
type Discover interface {
	// Name 插件名称
	Name() string
	// Version 版本信息
	Version() string

	// GetClient 解析服务发现配置,并返回client
	// 如果无法解析，那么返回错误
	GetClient(args map[string]string) (client *fasthttp.HostClient, err error)
}

type discoverRegistries struct {
	sync.Mutex
	opts map[string]Discover
}

func RegisterDiscover(discover Discover) error {
	log().Info("register discover", "name", discover.Name(), "version", discover.Version())

	if discover.Name() == "" {
		return ErrMissingName
	}

	discoverRegistry.Lock()
	defer discoverRegistry.Unlock()
	if _, found := discoverRegistry.opts[discover.Name()]; found {
		return fmt.Errorf("discover %s registered", discover.Name())
	}
	discoverRegistry.opts[discover.Name()] = discover
	return nil
}

func FindDiscover(name string) Discover {
	if opt, found := discoverRegistry.opts[name]; found {
		return opt
	}
	return nil
}
