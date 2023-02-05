// Package rule
//
// @author: xwc1125
package rule

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

func TestRule(t *testing.T) {
	var rules = map[string]string{
		// "^/check/(.*)":      "/check/rewrite/$1",
		// "/old":              "/new",
		"/api/*":            "/$1",
		"/js/*":             "/public/javascripts/$1",
		"/users/*/orders/*": "/user/$1/order/$2",
	}
	rulesRegex := map[*regexp.Regexp]string{}
	// Initialize
	for k, v := range rules {
		k = strings.Replace(k, "*", "(.*)", -1)
		k = k + "$"
		rulesRegex[regexp.MustCompile(k)] = v
	}

	// Rewrite
	// pathUri := "http://localhost:3000/old"
	// pathUri:="http://localhost:3000/old/hello"
	// pathUri := "/check/getIp"
	// pathUri := "/api/anything/1"
	// pathUri := "/js/anything/1"
	pathUri := "/users/1/orders/2/anything/1?aa"
	for k, v := range rulesRegex {
		replacer := captureTokens(k, pathUri)
		if replacer != nil {
			fmt.Println(replacer.Replace(v))
		}
	}
}

// https://github.com/labstack/echo/blob/master/middleware/rewrite.go
func captureTokens(pattern *regexp.Regexp, input string) *strings.Replacer {
	groups := pattern.FindAllStringSubmatch(input, -1)
	if groups == nil {
		return nil
	}
	values := groups[0][1:]
	replace := make([]string, 2*len(values))
	for i, v := range values {
		j := 2 * i
		replace[j] = "$" + strconv.Itoa(i+1)
		replace[j+1] = v
	}
	return strings.NewReplacer(replace...)
}
