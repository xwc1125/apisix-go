// Package models
//
// @author: xwc1125
package models

import "encoding/json"

type Response struct {
	ErrorMsg string `json:"error_msg,omitempty"`
}

func (r Response) SetErrMsg(errMsg string) Response {
	r.ErrorMsg = errMsg
	return r
}

func (r Response) String() string {
	bytes, _ := json.Marshal(r)
	return string(bytes)
}
