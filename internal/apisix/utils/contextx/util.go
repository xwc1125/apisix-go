// Package contextx
//
// @author: xwc1125
package contextx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/chain5j/logger"
	"github.com/google/uuid"
)

var (
	RequestIDKey = "X-Request-Id"
	// MaxInMemoryMultipartSize 32 MB in memory max
	MaxInMemoryMultipartSize = int64(32000000)
)

var reqWriteExcludeHeaderDump = map[string]bool{
	"Host":              true,
	"Content-Length":    true,
	"Transfer-Encoding": true,
	"Trailer":           true,
	"Accept-Encoding":   false,
	"Accept-Language":   false,
	"Cache-Control":     false,
	"Connection":        false,
	"Origin":            false,
	"User-Agent":        false,
}

func ReadQueryParams(req *http.Request) map[string]string {
	params := map[string]string{}
	u, err := url.Parse(req.RequestURI)
	if err != nil {
		return params
	}
	for k, v := range u.Query() {
		if len(v) < 1 {
			continue
		}
		// TODO: v is a list, and we should be showing a list of values
		// rather than assuming a single value always, gotta change this
		params[k] = v[0]
	}
	return params
}
func UrlParamMap(ctx Context) map[string]string {
	return ReadQueryParams(ctx.Request())
}

func UrlParamMapWithoutPage(ctx Context) map[string]interface{} {
	return FilterUrlParam(UrlParamMap(ctx))
}
func FilterUrlParam(params map[string]string) map[string]interface{} {
	p := map[string]interface{}{}
	if params == nil || len(params) == 0 {
		return p
	}
	for k, v := range params {
		switch k {
		case "page":
		case "limit":
		case "sort":
		case "order":
		case "startTime":
		case "endTime":
		case "uid":
		default:
			p[k] = v
		}
	}
	return p
}

func PostForm(req *http.Request, key string) string {
	return req.FormValue(key)
}
func ReadMultiPostForm(mpForm *multipart.Form) map[string]string {
	postForm := map[string]string{}
	if mpForm == nil {
		return postForm
	}
	for key, val := range mpForm.Value {
		postForm[key] = val[0]
	}
	return postForm
}
func ReadPostForm(req *http.Request) map[string]string {
	postForm := map[string]string{}
	for _, param := range strings.Split(*ReadBody(req), "&") {
		value := strings.Split(param, "=")
		postForm[value[0]] = value[1]
	}
	return postForm
}

// ReadHeaders 获取头部
func ReadHeaders(req *http.Request) map[string]string {
	b := bytes.NewBuffer([]byte(""))
	err := req.Header.WriteSubset(b, reqWriteExcludeHeaderDump)
	if err != nil {
		return map[string]string{}
	}
	headers := map[string]string{}
	for _, header := range strings.Split(b.String(), "\n") {
		values := strings.Split(header, ":")
		if strings.EqualFold(values[0], "") {
			continue
		}
		headers[values[0]] = values[1]
	}
	return headers
}

// ReadHeadersFromResponse 从响应中读取头部
func ReadHeadersFromResponse(header http.Header) map[string]string {
	headers := map[string]string{}
	for k, v := range header {
		headers[k] = strings.Join(v, " ")
	}
	return headers
}

// ReadBody 获取请求内容
func ReadBody(req *http.Request) *string {
	save := req.Body
	var err error
	if req.Body == nil {
		req.Body = nil
	} else {
		save, req.Body, err = drainBody(req.Body)
		if err != nil {
			return nil
		}
	}
	b := bytes.NewBuffer([]byte(""))
	chunked := len(req.TransferEncoding) > 0 && req.TransferEncoding[0] == "chunked"
	if req.Body == nil {
		return nil
	}
	var dest io.Writer = b
	if chunked {
		dest = httputil.NewChunkedWriter(dest)
	}
	_, err = io.Copy(dest, req.Body)
	if chunked {
		dest.(io.Closer).Close()
	}
	req.Body = save
	body := b.String()
	return &body
}

func drainBody(b io.ReadCloser) (r1, r2 io.ReadCloser, err error) {
	var buf bytes.Buffer
	if _, err = buf.ReadFrom(b); err != nil {
		return nil, nil, err
	}
	if err = b.Close(); err != nil {
		return nil, nil, err
	}
	return ioutil.NopCloser(&buf), ioutil.NopCloser(bytes.NewReader(buf.Bytes())), nil
}

func GetRequestID(request *http.Request) string {
	requestId := request.Header.Get(RequestIDKey)
	if requestId == "" {
		requestId = uuid.New().String()
		requestId = strings.ReplaceAll(requestId, "-", "")
		request.Header.Set(RequestIDKey, requestId)
	}
	return requestId
}

func JSON(ctx Context, code int, obj interface{}) {
	ctx.Header(ContentTypeHeaderKey, "application/json")
	bytes, err := json.Marshal(obj)
	if err != nil {
		http.Error(ctx.ResponseWriter(), err.Error(), 500)
		return
	}
	ctx.Status(code)
	ctx.ResponseWriter().Write(bytes)
}

func Text(ctx Context, code int, format string, values ...interface{}) {
	ctx.Header(ContentTypeHeaderKey, "text/plain")
	Data(ctx, code, []byte(fmt.Sprintf(format, values...)))
}

func HTML(ctx Context, code int, html string) {
	ctx.Header(ContentTypeHeaderKey, "text/html")
	Data(ctx, code, []byte(html))
}

func JSONP(ctx Context, code int, callback string, obj interface{}) {
	ctx.Header(ContentTypeHeaderKey, "application/javascript")
	var result []byte
	if callback != "" {
		result = append(result, []byte(callback+"(")...)
	}
	bytes, err := json.Marshal(obj)
	if err != nil {
		logger.Error("response json.Marshal err", "err", err)
		http.Error(ctx.ResponseWriter(), err.Error(), 500)
		return
	}
	result = append(result, bytes...)
	if callback != "" {
		result = append(result, []byte(");")...)
	}

	Data(ctx, code, bytes)
}

func Data(ctx Context, code int, data []byte) {
	ctx.Status(code)
	ctx.ResponseWriter().Write(data)
}

func SendFile(ctx Context, filename string, destinationName string) error {
	ctx.ResponseWriter().Header().Set(ContentDispositionHeaderKey, "attachment;filename="+destinationName)
	return ServeFile(ctx, filename, false)
}

func ServeFile(ctx Context, filename string, gzipCompression bool) error {
	f, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("%d", 404)
	}
	defer f.Close()
	fi, _ := f.Stat()
	if fi.IsDir() {
		return ServeFile(ctx, path.Join(filename, "index.html"), gzipCompression)
	}

	return ServeContent(ctx, f, fi.Name(), fi.ModTime(), gzipCompression)
}

func ServeContent(ctx Context, content io.ReadSeeker, filename string, modtime time.Time, gzipCompression bool) error {
	ctx.Header(ContentTypeHeaderKey, filename)
	// ctx.SetLastModified(modtime)
	var out io.Writer
	out = ctx.ResponseWriter()
	_, err := io.Copy(out, content)
	if err != nil {
		return err
	}
	return nil
}
