package cors

import (
	"net/http"
	"sync"

	"github.com/rs/cors"
)

type Options struct {
	// AllowedOrigins is a list of origins a cross-domain request can be executed from.
	// If the special "*" value is present in the list, all origins will be allowed.
	// An origin may contain a wildcard (*) to replace 0 or more characters
	// (i.e.: http://*.domain.com). Usage of wildcards implies a small performance penalty.
	// Only one wildcard can be used per origin.
	// Default value is ["*"]
	AllowedOrigins []string `json:"allowed_origins" mapstructure:"allowed_origins" yaml:"allowed_origins"`
	// AllowOriginFunc is a custom function to validate the origin. It take the origin
	// as argument and returns true if allowed or false otherwise. If this option is
	// set, the content of AllowedOrigins is ignored.
	AllowOriginFunc func(origin string) bool
	// AllowOriginFunc is a custom function to validate the origin. It takes the HTTP Request object and the origin as
	// argument and returns true if allowed or false otherwise. If this option is set, the content of `AllowedOrigins`
	// and `AllowOriginFunc` is ignored.
	AllowOriginRequestFunc func(r *http.Request, origin string) bool
	// AllowedMethods is a list of methods the client is allowed to use with
	// cross-domain requests. Default value is simple methods (HEAD, GET and POST).
	AllowedMethods []string `json:"allowed_methods" mapstructure:"allowed_methods" yaml:"allowed_methods"`
	// AllowedHeaders is list of non simple headers the client is allowed to use with
	// cross-domain requests.
	// If the special "*" value is present in the list, all headers will be allowed.
	// Default value is [] but "Origin" is always appended to the list.
	AllowedHeaders []string `json:"allowed_headers" mapstructure:"allowed_headers" yaml:"allowed_headers"`
	// ExposedHeaders indicates which headers are safe to expose to the API of a CORS
	// API specification
	ExposedHeaders []string `json:"exposed_headers" mapstructure:"exposed_headers" yaml:"exposed_headers"`
	// MaxAge indicates how long (in seconds) the results of a preflight request
	// can be cached
	MaxAge int `json:"max_age" mapstructure:"max_age" yaml:"max_age"`
	// AllowCredentials indicates whether the request can include user credentials like
	// cookies, HTTP authentication or client side SSL certificates.
	AllowCredentials bool `json:"allow_credentials" mapstructure:"allow_credentials" yaml:"allow_credentials"`
	// OptionsPassthrough instructs preflight to let other potential next handlers to
	// process the OPTIONS method. Turn this on if your application handles OPTIONS.
	OptionsPassthrough bool `json:"options_passthrough" mapstructure:"options_passthrough" yaml:"options_passthrough"`
	// Debugging flag adds additional output to debug server side CORS issues
	Debug bool `json:"debug" mapstructure:"debug" yaml:"debug"`
}
type Cors struct {
	*cors.Cors
}

var (
	_ Handler = new(Cors)
)

type Handler interface {
	Handler(h http.Handler) http.Handler
	HandlerFunc(w http.ResponseWriter, r *http.Request)
	ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc)
}

var (
	corsOne  *Cors
	corsOnce sync.Once
)

func NewOnce(options Options) *Cors {
	corsOnce.Do(func() {
		corsOne = New(options)
	})
	return corsOne
}

func New(options Options) *Cors {
	return &Cors{
		cors.New(cors.Options{
			AllowedOrigins:         options.AllowedOrigins,
			AllowOriginFunc:        options.AllowOriginFunc,
			AllowOriginRequestFunc: options.AllowOriginRequestFunc,
			AllowedMethods:         options.AllowedMethods,
			AllowedHeaders:         options.AllowedHeaders,
			ExposedHeaders:         options.ExposedHeaders,
			MaxAge:                 options.MaxAge,
			AllowCredentials:       options.AllowCredentials,
			OptionsPassthrough:     options.OptionsPassthrough,
			Debug:                  options.Debug,
		}),
	}
}
func Default() *Cors {
	return New(Options{})
}
func AllowAll() *Cors {
	return &Cors{
		cors.AllowAll(),
	}
}
func (c *Cors) Handler(h http.Handler) http.Handler {
	return c.Cors.Handler(h)
}

func (c *Cors) HandlerFunc(w http.ResponseWriter, r *http.Request) {
	c.Cors.HandlerFunc(w, r)
}

func (c *Cors) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	c.Cors.ServeHTTP(w, r, next)
}
