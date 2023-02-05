// Package fasthttp
//
// @author: xwc1125
package fasthttp

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/valyala/fasthttp"
	"github.com/xwc1125/apisix-go/internal/xgateway"
	"golang.org/x/net/http/httpguts"
)

var (
	_ xgateway.Cookie = new(Cookie)
)

type Cookie struct {
	*fasthttp.Cookie
}

func NewCookie() *Cookie {
	cookie := fasthttp.AcquireCookie()
	return &Cookie{
		Cookie: cookie,
	}
}

func (c *Cookie) SameSite() xgateway.SameSite {
	return xgateway.SameSite(c.Cookie.SameSite())
}

func (c *Cookie) SetSameSite(mode xgateway.SameSite) {
	c.Cookie.SetSameSite(fasthttp.CookieSameSite(mode))
}

func (c *Cookie) Valid() error {
	if c == nil || c.Cookie == nil {
		return errors.New("http: nil Cookie")
	}
	if !isCookieNameValid(string(c.Key())) {
		return errors.New("http: invalid Cookie.Name")
	}
	if !validCookieExpires(c.Expire()) {
		return errors.New("http: invalid Cookie.Expires")
	}
	value := c.Value()
	for i := 0; i < len(value); i++ {
		if !validCookieValueByte(value[i]) {
			return fmt.Errorf("http: invalid byte %q in Cookie.Value", value[i])
		}
	}
	path := c.Path()
	if len(path) > 0 {
		for i := 0; i < len(path); i++ {
			if !validCookiePathByte(path[i]) {
				return fmt.Errorf("http: invalid byte %q in Cookie.Path", path[i])
			}
		}
	}
	if len(c.Domain()) > 0 {
		if !validCookieDomain(string(c.Domain())) {
			return errors.New("http: invalid Cookie.Domain")
		}
	}
	return nil
}

func isCookieNameValid(raw string) bool {
	if raw == "" {
		return false
	}
	return strings.IndexFunc(raw, isNotToken) < 0
}
func isNotToken(r rune) bool {
	return !httpguts.IsTokenRune(r)
}

// validCookieExpires reports whether v is a valid cookie expires-value.
func validCookieExpires(t time.Time) bool {
	// IETF RFC 6265 Section 5.1.1.5, the year must not be less than 1601
	return t.Year() >= 1601
}

func validCookieValueByte(b byte) bool {
	return 0x20 <= b && b < 0x7f && b != '"' && b != ';' && b != '\\'
}
func validCookiePathByte(b byte) bool {
	return 0x20 <= b && b < 0x7f && b != ';'
}

func validCookieDomain(v string) bool {
	if isCookieDomainName(v) {
		return true
	}
	if net.ParseIP(v) != nil && !strings.Contains(v, ":") {
		return true
	}
	return false
}

func isCookieDomainName(s string) bool {
	if len(s) == 0 {
		return false
	}
	if len(s) > 255 {
		return false
	}

	if s[0] == '.' {
		// A cookie a domain attribute may start with a leading dot.
		s = s[1:]
	}
	last := byte('.')
	ok := false // Ok once we've seen a letter.
	partlen := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		default:
			return false
		case 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z':
			// No '_' allowed here (in contrast to package net).
			ok = true
			partlen++
		case '0' <= c && c <= '9':
			// fine
			partlen++
		case c == '-':
			// Byte before dash cannot be dot.
			if last == '.' {
				return false
			}
			partlen++
		case c == '.':
			// Byte before dot cannot be dot, dash.
			if last == '.' || last == '-' {
				return false
			}
			if partlen > 63 || partlen == 0 {
				return false
			}
			partlen = 0
		}
		last = c
	}
	if last == '-' || partlen > 63 {
		return false
	}

	return ok
}
