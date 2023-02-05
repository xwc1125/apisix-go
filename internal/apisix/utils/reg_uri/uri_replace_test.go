// Package reg_uri
//
// @author: xwc1125
package reg_uri

import (
	"fmt"
	"net/url"
	"testing"
)

func TestNewRule(t *testing.T) {
	{
		rule, err := NewRule("^/check/(.*)", "/check/rewrite/$1")
		if err != nil {
			t.Fatal(err)
		}
		replace := rule.Replace("/check/getIp")
		fmt.Println(replace)
	}
	{
		rule, err := NewRule("/old", "/new")
		if err != nil {
			t.Fatal(err)
		}
		replace := rule.Replace("/old")
		fmt.Println(replace)
	}
	{
		// rule, err := NewRule("/api/*", "/$1")// out: /
		rule, err := NewRule("/api/(.*)", "/$1")
		if err != nil {
			t.Fatal(err)
		}
		replace := rule.Replace("/api/anything/1")
		fmt.Println(replace)
	}
	{
		rule, err := NewRule("/js/(.*)", "/public/javascripts/$1")
		if err != nil {
			t.Fatal(err)
		}
		replace := rule.Replace("/js/anything/1")
		fmt.Println(replace)
	}
	{
		rule, err := NewRule("/users/(.*)/orders/(.*)", "/user/$1/order/$2")
		if err != nil {
			t.Fatal(err)
		}
		replace := rule.Replace("/users/1/orders/2/anything/1?aa")

		fmt.Println(replace)
		queryUnescape, _ := url.QueryUnescape(replace)
		fmt.Println(queryUnescape)
	}
}
