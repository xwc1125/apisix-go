// Package xgateway
//
// @author: xwc1125
package xgateway

import (
	"io"
)

// Header is like http.Header, but only implements the subset of its methods
type Header interface {
	Add(key, value string)
	Set(key, value string)
	Del(key string)
	Get(key string) string

	ConnectionUpgrade() bool
	ContentLength() int
	SetContentLength(contentLength int)
	ContentType() []byte
	SetContentType(contentType string)
	ContentEncoding() []byte
	SetContentEncoding(contentEncoding string)
	Host() []byte
	SetHost(host string)
	UserAgent() []byte
	SetUserAgent(userAgent string)
	Method() []byte
	SetMethod(method string)
	Protocol() []byte
	SetProtocol(method string)
	RequestURI() []byte
	SetRequestURI(requestURI string)

	Cookie(key string) []byte
	AllCookie(f func(key, value []byte))
	SetCookie(cookie Cookie)
	AddCookie(key, value string)
	DelCookie(key string)

	WriteTo(w io.Writer) (int64, error)
	Headers() []byte

	// resp
	StatusCode() int
	SetStatusCode(statusCode int)
}
