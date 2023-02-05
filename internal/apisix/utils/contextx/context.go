// Package contextx
//
// @author: xwc1125
package contextx

import (
	"context"
	"mime/multipart"
	"net/http"
)

const (
	ContentTypeHeaderKey        = "Content-Type"
	ContentDispositionHeaderKey = "Content-Disposition"
	ContentLengthHeaderKey      = "Content-Length"
	ContentEncodingHeaderKey    = "Content-Encoding"
)

type Handler func(Context)

type Context interface {
	context.Context

	Request() *http.Request
	ResponseWriter() http.ResponseWriter
	RequestId() string

	Next()
	Abort() // 终止。如果授权失败（例如：密码不匹配），请调用Abort以确保不调用此请求的其余处理程序
	AbortWithStatus(code int)
	AbortWithStatusJSON(code int, jsonObj interface{})
	IsAborted() bool // 判断当前context是否已终止

	Param(key string) string
	Query(key string) string
	PostForm(key string) string
	FormFile(name string) (*multipart.FileHeader, error)
	MultipartForm() (*multipart.Form, error)
	SaveUploadedFile(file *multipart.FileHeader, dst string) error
	Bind(obj interface{}, bindType ...string) error
	ContentType() string

	// =========METADATA=======

	Set(key string, value interface{})
	Get(key string) (value interface{}, exists bool)
	GetString(key string) (s string)
	GetBool(key string) (b bool)
	GetInt64(key string) (i64 int64)
	// =========RESPONSE=======

	Status(code int)
	GetStatus() int
	Header(key, value string)
	GetHeader(key string) string
	SetCookie(cookie *http.Cookie)
	GetCookie(name string) (string, error)
	Redirect(code int, location string)
}
