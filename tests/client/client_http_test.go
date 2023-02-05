// Package main
//
// @author: xwc1125
package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"
)

// unix socket确实比tcp socket快上一点，多次测试的也是如此。
// 这里大概快了七分之一左右。
// 也印证了tcp保证的可靠传输(校验和，流量控制等)确实会导致一些效率上的牺牲。
func TestClientGet(t *testing.T) {
	httpc := http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", UNIX_SOCK_PIPE_PATH)
			},
		},
	}

	resp, err := httpc.Get("http://http.sock/testGet")
	if err != nil {
		t.Fatal(err)
		return
	}
	fmt.Println(resp.Status)
	for key, val := range resp.Header {
		fmt.Println(key, "=", val)
	}

	all, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(all))
}
