package proxy

import (
	"crypto/tls"
	"plugin"
	"time"
)

// Option to define all options to reverse http proxy.
type Option interface {
	apply(o *buildOption)
}

// buildOption contains all fields those are used in ReverseProxy.
type buildOption struct {
	// tlsConfig is pointer to tls.Config, will be used if the upstream.
	// need TLS handshake
	tlsConfig *tls.Config

	// timeout specify the timeout context with each request.
	timeout time.Duration

	// disablePathNormalizing disable path normalizing.
	disablePathNormalizing bool
	// plugins 插件集合
	plugins          []plugin.Plugin
	compressionLevel int
}

type funcBuildOption struct {
	f func(o *buildOption)
}

func newFuncBuildOption(f func(o *buildOption)) funcBuildOption { return funcBuildOption{f: f} }
func (fb funcBuildOption) apply(o *buildOption)                 { fb.f(o) }

func WithTLSConfig(config *tls.Config) Option {
	return newFuncBuildOption(func(o *buildOption) {
		o.tlsConfig = config
	})
}

// WithTLS build tls.Config with server certFile and keyFile.
// tlsConfig is nil as default
func WithTLS(certFile, keyFile string) Option {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		panic("" + err.Error())
	}

	return WithTLSConfig(&tls.Config{
		Certificates: []tls.Certificate{cert},
	})
}

// WithTimeout specify the timeout of each request
func WithTimeout(d time.Duration) Option {
	return newFuncBuildOption(func(o *buildOption) {
		o.timeout = d
	})
}

// WithDisablePathNormalizing sets whether disable path normalizing.
func WithDisablePathNormalizing(isDisablePathNormalizing bool) Option {
	return newFuncBuildOption(func(o *buildOption) {
		o.disablePathNormalizing = isDisablePathNormalizing
	})
}

// WithPlugins sets whether disable path normalizing.
func WithPlugins(plugins ...plugin.Plugin) Option {
	return newFuncBuildOption(func(o *buildOption) {
		o.plugins = plugins
	})
}

// WithCompressionLevel
func WithCompressionLevel(compressionLevel int) Option {
	return newFuncBuildOption(func(o *buildOption) {
		o.compressionLevel = compressionLevel
	})
}
