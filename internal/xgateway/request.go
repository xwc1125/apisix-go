// Package xgateway
//
// @author: xwc1125
package xgateway

import (
	"context"
	"io"
	"mime/multipart"
	"net"
	"net/url"
)

type Request interface {
	ID() uint64
	Context() context.Context
	Header() Header

	Host() []byte
	SetHost(host string)
	RequestURI() []byte
	SetRequestURI(requestURI string)

	SetBody(body []byte)
	Body() ([]byte, error)
	AppendBody(p []byte)
	BodyWriteTo(w io.Writer) error

	WriteTo(w io.Writer) (int64, error)
	MultipartForm() (*multipart.Form, error)

	Cookies() []Cookie
	Cookie(name string) (Cookie, error)
	AddCookie(c Cookie)

	RemoteIP() net.IP
	Method() string
	Path() []byte
	SetPath([]byte)
	Args() url.Values
}
