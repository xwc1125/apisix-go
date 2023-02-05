// Package example
//
// @author: xwc1125
package main

import (
	"github.com/xwc1125/apisix-go/internal/pkg/version"
)

func main() {
	version.FilePath = "./logs"
	if version.Build("App: versionApp") {
		return
	}
}
