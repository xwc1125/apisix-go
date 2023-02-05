// Package cgw
//
// @author: xwc1125
package cgw

// InterfacePart defines the inner content of cgw-request
type InterfacePart struct {
	InterfaceName string      `json:"interfaceName,omitempty"`
	Param         interface{} `json:"para"` // json params for different products
}
