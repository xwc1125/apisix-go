// Package iputils
//
// @author: xwc1125
package iputils

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/valyala/fasthttp"
	"github.com/xwc1125/apisix-go/internal/xgateway"
)

func ClientIP(req *fasthttp.Request) string {
	clientIP := string(req.Header.Peek(xgateway.HeaderXForwardedFor))
	if index := strings.IndexByte(clientIP, ','); index >= 0 {
		clientIP = clientIP[0:index]
	}
	clientIP = strings.TrimSpace(clientIP)
	if len(clientIP) > 0 {
		return clientIP
	}
	clientIP = strings.TrimSpace(string(req.Header.Peek(xgateway.HeaderXRealIP)))
	if len(clientIP) > 0 {
		return clientIP
	}
	return ""
}

// Ip2Binary 将IP地址转化为二进制String
func Ip2Binary(ip string) string {
	str := strings.Split(ip, ".")
	var ipstr string
	for _, s := range str {
		i, err := strconv.ParseUint(s, 10, 8)
		if err != nil {
			fmt.Println(err)
		}
		ipstr = ipstr + fmt.Sprintf("%08b", i)
	}
	return ipstr
}

// Match 测试IP地址和地址端是否匹配 变量ip为字符串，例子"192.168.56.4" iprange为地址端"192.168.56.64/26"
func Match(ip, ipRange string) bool {
	ipb := Ip2Binary(ip)
	ipr := strings.Split(ipRange, "/")
	iprb := Ip2Binary(ipr[0])

	if len(ipr) > 1 {
		masklen, err := strconv.ParseUint(ipr[1], 10, 32)
		if err != nil {
			fmt.Println(err)
			return false
		}
		return strings.EqualFold(ipb[0:masklen], iprb[0:masklen])
	}
	return strings.EqualFold(ipb, iprb)
}
