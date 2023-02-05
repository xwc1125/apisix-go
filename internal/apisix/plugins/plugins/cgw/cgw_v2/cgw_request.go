// Package cgw_v2
//
// @author: xwc1125
package cgw_v2

import (
	"math/rand"
	"time"
)

const (
	RequestDefaultVersion = "1.0"
)

func init() {
	rand.Seed(time.Now().Unix())
}

// Request The protocol defined by Request is old cgw protocol, used by interfaces of many other inner products .
type Request struct {
	// httppost.Request
	Version       string        `json:"version"`
	Caller        string        `json:"caller"`
	Password      string        `json:"password"`
	Callee        string        `json:"callee"`
	EventId       int64         `json:"eventId"`
	Timestamp     int64         `json:"timestamp"`
	InterfacePart InterfacePart `json:"interface"`
}

// InterfacePart defines the inner content of cgw-request
type InterfacePart struct {
	InterfaceName string      `json:"interfaceName,omitempty"`
	Param         interface{} `json:"para"` // json params for different products
}

// NewCgwRequest wraps request information in the format of api v2.0
func NewCgwRequest(action string, innerRequest interface{}) (*Request, error) {
	vReq := &Request{
		Version: RequestDefaultVersion,
		// Caller:    captain.GetDomain(),
		EventId:   int64(rand.Int()),
		Timestamp: time.Now().Unix(),
		InterfacePart: InterfacePart{
			InterfaceName: action,
			Param:         innerRequest,
		},
	}
	// vReq.SetAction(innerRequest.GetAction())
	return vReq, nil
}
