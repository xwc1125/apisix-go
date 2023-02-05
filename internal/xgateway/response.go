// Package xgateway
//
// @author: xwc1125
package xgateway

// Response represents the HTTP response from the upstream received by APISIX.
// In order to avoid semantic misunderstanding,
// we also use Response to represent the rewritten response from Plugin Runner.
// Therefore, any instance that implements the Response interface will be readable and rewritable.
type Response interface {
	// ID returns the request id
	ID() uint32

	// StatusCode returns the response code
	StatusCode() int

	// Header returns the response header.
	//
	// It allows you to add or set response headers before reaching the client.
	Header() Header

	// Var returns the value of a Nginx variable, like `r.Var("request_time")`
	//
	// To fetch the value, the runner will look up the request's cache first. If not found,
	// the runner will ask it from the APISIX. If the RPC call is failed, an error in
	// pkg/common.ErrConnClosed type is returned.
	Var(name string) ([]byte, error)

	// ReadBody returns origin HTTP response body
	//
	// To fetch the value, the runner will look up the request's cache first. If not found,
	// the runner will ask it from the APISIX. If the RPC call is failed, an error in
	// pkg/common.ErrConnClosed type is returned.
	//
	// It was not named `Body`
	// because `Body` was already occupied in earlier interface implementations.
	ReadBody() ([]byte, error)

	// Write rewrites the origin response data.
	//
	// Unlike `ResponseWriter.Write`, we don't need to WriteHeader(http.StatusOK)
	// before writing the data
	// Because APISIX will convert code 0 to 200.
	Write(b []byte) (int, error)

	// WriteHeader rewrites the origin response StatusCode
	//
	// WriteHeader can't override written status.
	WriteHeader(statusCode int)
}
