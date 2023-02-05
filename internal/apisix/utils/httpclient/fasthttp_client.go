// Package httpclient
//
// @author: xwc1125
package httpclient

import (
	"fmt"

	"github.com/valyala/fasthttp"
)

func NewFastHttpClient(hostClient *fasthttp.HostClient, addr, reqUri string) ([]byte, error) {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)
	defer fasthttp.ReleaseRequest(req)
	req.SetRequestURI(fmt.Sprintf("http://%s%s", addr, reqUri))
	// req.Header.Set(`Bearer `+xgateway.HeaderAuthorization, token)
	var (
		err error
	)

	if hostClient.ReadTimeout > 0 {
		err = hostClient.DoTimeout(req, resp, hostClient.ReadTimeout)
	} else {
		err = hostClient.Do(req, resp)
	}
	if err != nil {
		return nil, err
	}
	return resp.Body(), nil
}
