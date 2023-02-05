// Package params
//
// @author: xwc1125
package params

import (
	"fmt"

	"github.com/xwc1125/apisix-go/internal/pkg/version"
)

var (
	defaultApp     = "Apisix-Go"
	defaultVersion = "0.0.1"
	defaultWelcome = "\n" +
		"Welcome to %s(%s)"
)

func App() string {
	if len(version.AppName) > 0 {
		return version.AppName
	}
	return defaultApp
}

func Version() string {
	if len(version.GetVersion()) > 0 {
		return version.GetVersion()
	}
	return defaultVersion
}

func Welcome() string {
	return fmt.Sprintf(defaultWelcome, App(), Version())
}
