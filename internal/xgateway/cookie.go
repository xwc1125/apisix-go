// Package xgateway
//
// @author: xwc1125
package xgateway

import (
	"io"
	"time"
)

type SameSite int

type Cookie interface {
	Key() []byte
	SetKey(key string)
	Value() []byte
	SetValue(value string)
	Path() []byte
	SetPath(path string)
	Domain() []byte
	SetDomain(domain string)
	Expire() time.Time
	SetExpire(expire time.Time)
	MaxAge() int
	SetMaxAge(seconds int)
	Secure() bool
	SetSecure(secure bool)
	HTTPOnly() bool
	SetHTTPOnly(httpOnly bool)
	SameSite() SameSite
	SetSameSite(mode SameSite)

	WriteTo(w io.Writer) (int64, error)
	Cookie() []byte
	String() string
	Valid() error
}
