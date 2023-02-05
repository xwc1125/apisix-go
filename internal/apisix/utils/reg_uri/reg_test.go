// Package reg_uri
//
// @author: xwc1125
package reg_uri

import (
	"fmt"
	"strings"
	"testing"
)

func TestReg(t *testing.T) {
	var uris = []string{
		"/foo/bar",
		"/foo/bar/foo",
		"/foo/bar/foo1",
		"foo/bar_project1_foo/",
		"/foo/bar_project1_foo/",
		"foo/123/foo/123",
		"/foo/123/foo/123",
		"foo/123/foo/456",
		"/foo/123/foo/456",
		"/foo/bar?status=1&type=2",
	}
	var regs = []string{
		"/foo/bar",
		"/foo/*",
		"/:foo",
		"/foo/:foo",
		"/foo/{foo}",
		"foo/bar_{project}_foo/",
		"/foo/{id}/foo/{id}",
		"/foo/{id}/foo/{id2}",
	}

	for _, reg := range regs {
		fmt.Println("=====================", reg)
		for _, uri := range uris {
			var buff = new(strings.Builder)
			if KeyMatch(uri, reg) {
				buff.WriteString("KeyMatch1")
				buff.WriteString(",")
			}
			if KeyMatch2(uri, reg) {
				buff.WriteString("KeyMatch2")
				buff.WriteString(",")
			}
			if KeyMatch3(uri, reg) {
				buff.WriteString("KeyMatch3")
				buff.WriteString(",")
			}
			if KeyMatch4(uri, reg) {
				buff.WriteString("KeyMatch4")
				buff.WriteString(",")
			}
			if KeyMatch5(uri, reg) {
				buff.WriteString("KeyMatch5")
				buff.WriteString(",")
			}
			if buff.Len() > 0 {
				fmt.Println("-----------------")
				fmt.Println("reg=", reg, ",uri=", uri)
				fmt.Println(buff.String())
			}
		}
	}
}

func TestDomainMatch(t *testing.T) {
	{
		reg := "*.xwc1125.com"
		domain := "b.xwc1125.com"
		match := DomainMatch(domain, reg)
		t.Log(match)
	}
	{
		reg := "192.168.1.0"
		domain := "192.168.1.0/24"
		match := DomainMatch(domain, reg)
		t.Log(match)
	}
}
