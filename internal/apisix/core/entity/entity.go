/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package entity

import (
	"reflect"
	"time"

	"github.com/chain5j/chain5j-pkg/util/convutil"
	"github.com/chain5j/logger"
	"github.com/xwc1125/apisix-go/internal/apisix/utils/uuid"
)

type BaseInfo struct {
	ID         interface{} `json:"id" comment:"ID"`
	CreateTime int64       `json:"create_time,omitempty" comment:"创建时间"`
	UpdateTime int64       `json:"update_time,omitempty" comment:"更新时间"`
}

func (info *BaseInfo) GetBaseInfo() *BaseInfo {
	return info
}

func (info *BaseInfo) Creating() {
	if info.ID == nil {
		info.ID = uuid.GetFlakeUidStr()
	} else {
		// convert to string if it's not
		if reflect.TypeOf(info.ID).String() != "string" {
			info.ID = convutil.ToString(info.ID)
		}
	}
	info.CreateTime = time.Now().Unix()
	info.UpdateTime = time.Now().Unix()
}

func (info *BaseInfo) Updating(storedInfo *BaseInfo) {
	info.ID = storedInfo.ID
	info.CreateTime = storedInfo.CreateTime
	info.UpdateTime = time.Now().Unix()
}

func (info *BaseInfo) KeyCompat(key string) {
	if info.ID == nil && key != "" {
		info.ID = key
	}
}

type Status uint8

// swagger:model Route
type Route struct {
	BaseInfo
	URI             string                 `json:"uri,omitempty" comment:"单个http请求路径"`
	Uris            []string               `json:"uris,omitempty" comment:"多个http请求路径"`
	Name            string                 `json:"name" validate:"max=100" comment:"名称"`
	Desc            string                 `json:"desc,omitempty" validate:"max=256" default:"描述"`
	Priority        int                    `json:"priority,omitempty" comment:"优先级"`
	Methods         []string               `json:"methods,omitempty" comment:"允许的http 方法"`
	Host            string                 `json:"host,omitempty" comment:"单个域名"`
	Hosts           []string               `json:"hosts,omitempty" comment:"多个域名"`
	RemoteAddr      string                 `json:"remote_addr,omitempty" comment:"单个客户端地址"`
	RemoteAddrs     []string               `json:"remote_addrs,omitempty" comment:"多个客户端地址"`
	Vars            []interface{}          `json:"vars,omitempty" comment:"支持通过请求头，请求参数、Cookie 进行路由匹配，可应用于灰度发布，蓝绿测试等场景。"`
	FilterFunc      string                 `json:"filter_func,omitempty" comment:"FilterFunc"`
	Script          interface{}            `json:"script,omitempty" comment:"Script"`
	ScriptID        interface{}            `json:"script_id,omitempty" comment:"ScriptID"` // For debug and optimization(cache), currently same as Route's ID
	Plugins         map[string]interface{} `json:"plugins,omitempty" comment:"插件集"`
	PluginConfigID  interface{}            `json:"plugin_config_id,omitempty" comment:"插件配置ID"`
	Upstream        *UpstreamDef           `json:"upstream,omitempty" comment:"Upstream"`
	ServiceID       interface{}            `json:"service_id,omitempty" comment:"绑定服务ID"`
	UpstreamID      interface{}            `json:"upstream_id,omitempty" comment:"UpstreamID"`
	ServiceProtocol string                 `json:"service_protocol,omitempty" comment:"ServiceProtocol"`
	Labels          map[string]string      `json:"labels,omitempty" comment:"标签"`
	EnableWebsocket bool                   `json:"enable_websocket,omitempty" comment:"是否为websocket"`
	Status          Status                 `json:"status" comment:"状态：1已发布，0待发布"`
}

// --- structures for upstream start  ---
type TimeoutValue float32
type Timeout struct {
	Connect TimeoutValue `json:"connect,omitempty"`
	Send    TimeoutValue `json:"send,omitempty"`
	Read    TimeoutValue `json:"read,omitempty"`
}

type Node struct {
	Host     string      `json:"host,omitempty"`
	Port     int         `json:"port,omitempty"`
	Weight   int         `json:"weight"`
	Metadata interface{} `json:"metadata,omitempty"`
	Priority int         `json:"priority,omitempty"`
}

type K8sInfo struct {
	Namespace   string `json:"namespace,omitempty"`
	DeployName  string `json:"deploy_name,omitempty"`
	ServiceName string `json:"service_name,omitempty"`
	Port        int    `json:"port,omitempty"`
	BackendType string `json:"backend_type,omitempty"`
}

type Healthy struct {
	Interval     int   `json:"interval,omitempty" comment:"间隔时间:秒"`
	HttpStatuses []int `json:"http_statuses,omitempty" comment:"状态码列表"` // HTTP 状态码列表，当探针在主动健康检查中返回时，视为健康。
	Successes    int   `json:"successes,omitempty" comment:"成功次数"`      // 若达到此值，表示上游服务目标节点是健康的。
}

type UnHealthy struct {
	Interval     int   `json:"interval,omitempty" comment:"间隔时间:秒"`
	HTTPStatuses []int `json:"http_statuses,omitempty" comment:"状态码列表"`
	TCPFailures  int   `json:"tcp_failures,omitempty" comment:"TCP 失败次数"`
	Timeouts     int   `json:"timeouts,omitempty" comment:"超时时间"`
	HTTPFailures int   `json:"http_failures,omitempty" comment:"HTTP 失败次数"` // 若达到此值，表示上游服务目标节点是不健康的。
}

// Active 主动检查
type Active struct {
	Type                   string       `json:"type,omitempty" comment:"类型：http/https"` // 是使用 HTTP 或 HTTPS 进行主动健康检查，还是只尝试 TCP 连接
	Timeout                TimeoutValue `json:"timeout,omitempty" comment:"超时时间"`
	Concurrency            int          `json:"concurrency,omitempty" comment:"并行数量"`
	Host                   string       `json:"host,omitempty" comment:"主机名"`
	Port                   int          `json:"port,omitempty" comment:"端口"`
	HTTPPath               string       `json:"http_path,omitempty" comment:"请求路径"`
	HTTPSVerifyCertificate bool         `json:"https_verify_certificate,omitempty"`
	Healthy                Healthy      `json:"healthy,omitempty" comment:"健康状态"`
	UnHealthy              UnHealthy    `json:"unhealthy,omitempty" comment:"不健康状态"`
	ReqHeaders             []string     `json:"req_headers,omitempty" comment:"额外的请求头"` // 示例：User-Agent: curl/7.29.0
}

// Passive 被动检查
type Passive struct {
	Type      string    `json:"type,omitempty"`
	Healthy   Healthy   `json:"healthy,omitempty"`
	UnHealthy UnHealthy `json:"unhealthy,omitempty"`
}

type HealthChecker struct {
	Active  Active  `json:"active,omitempty"`
	Passive Passive `json:"passive,omitempty"`
}

type UpstreamTLS struct {
	ClientCert string `json:"client_cert,omitempty"`
	ClientKey  string `json:"client_key,omitempty"`
}

type UpstreamKeepalivePool struct {
	IdleTimeout *TimeoutValue `json:"idle_timeout,omitempty"`
	Requests    int           `json:"requests,omitempty"`
	Size        int           `json:"size"`
}

type UpstreamDef struct {
	Nodes   interface{} `json:"nodes,omitempty" comment:"目标节点"`
	Retries *int        `json:"retries,omitempty" comment:"重试机制将请求发到下一个上游节点。值为 0 表示禁用重试机制，留空表示使用可用后端节点的数量。"`
	Timeout *Timeout    `json:"timeout,omitempty" comment:"超时"`
	Type    string      `json:"type,omitempty" comment:"lb负载均衡算法，chash,roundrobin"`
	HashOn  string      `json:"hash_on,omitempty" comment:"lb哈希位置"`
	Key     string      `json:"key,omitempty" comment:"lb哈希键"`
	Checks  interface{} `json:"checks,omitempty" comment:"健康检查配置"`
	Scheme  string      `json:"scheme,omitempty" comment:"协议:http,https,grpc,grpcs"`

	DiscoveryType string            `json:"discovery_type,omitempty" comment:"服务发现类型"`
	DiscoveryArgs map[string]string `json:"discovery_args,omitempty" comment:"服务发现参数"`
	ServiceName   string            `json:"服务名称,omitempty"`

	PassHost string `json:"pass_host,omitempty" comment:"Host 请求头"` // node：使用目标节点列表中的主机名活IP，pass：保持与客户端一致的主机名

	UpstreamHost  string                 `json:"upstream_host,omitempty"`
	Name          string                 `json:"name,omitempty" comment:"upstream名称"`
	Desc          string                 `json:"desc,omitempty" comment:"upstream描述"`
	Labels        map[string]string      `json:"labels,omitempty" comment:"标签"`
	TLS           *UpstreamTLS           `json:"tls,omitempty"`
	KeepalivePool *UpstreamKeepalivePool `json:"keepalive_pool,omitempty" comment:"连接池配置"`
	RetryTimeout  TimeoutValue           `json:"retry_timeout,omitempty"`
}

func (u UpstreamDef) GetNodes() []*Node {
	format := NodesFormat(u.Nodes)
	switch p := format.(type) {
	case []*Node:
		return p
	default:
		logger.Fatal("upstream nodes format err")
		return nil
	}
}

// swagger:model Upstream
type Upstream struct {
	BaseInfo
	UpstreamDef
}

type UpstreamNameResponse struct {
	ID   interface{} `json:"id"`
	Name string      `json:"name"`
}

func (upstream *Upstream) Parse2NameResponse() (*UpstreamNameResponse, error) {
	nameResp := &UpstreamNameResponse{
		ID:   upstream.ID,
		Name: upstream.Name,
	}
	return nameResp, nil
}

// --- structures for upstream end  ---

// swagger:model Consumer
type Consumer struct {
	Username   string                 `json:"username"`
	Desc       string                 `json:"desc,omitempty"`
	Plugins    map[string]interface{} `json:"plugins,omitempty"`
	Labels     map[string]string      `json:"labels,omitempty"`
	CreateTime int64                  `json:"create_time,omitempty"`
	UpdateTime int64                  `json:"update_time,omitempty"`
}

type SSLClient struct {
	CA    string `json:"ca,omitempty"`
	Depth int    `json:"depth,omitempty"`
}

// swagger:model SSL
type SSL struct {
	BaseInfo
	Cert          string            `json:"cert,omitempty"`
	Key           string            `json:"key,omitempty"`
	Sni           string            `json:"sni,omitempty"`
	Snis          []string          `json:"snis,omitempty"`
	Certs         []string          `json:"certs,omitempty"`
	Keys          []string          `json:"keys,omitempty"`
	ExpTime       int64             `json:"exptime,omitempty"`
	Status        int               `json:"status"`
	ValidityStart int64             `json:"validity_start,omitempty"`
	ValidityEnd   int64             `json:"validity_end,omitempty"`
	Labels        map[string]string `json:"labels,omitempty"`
	Client        *SSLClient        `json:"client,omitempty"`
}

// swagger:model Service
type Service struct {
	BaseInfo
	Name            string                 `json:"name,omitempty" comment:"服务名称"`
	Desc            string                 `json:"desc,omitempty" comment:"服务描述"`
	Upstream        *UpstreamDef           `json:"upstream,omitempty" comment:"上游服务"`
	UpstreamID      interface{}            `json:"upstream_id,omitempty" comment:"上游服务ID"`
	Plugins         map[string]interface{} `json:"plugins,omitempty" comment:"插件"`
	Script          string                 `json:"script,omitempty"`
	Labels          map[string]string      `json:"labels,omitempty" comment:"标签"`
	EnableWebsocket bool                   `json:"enable_websocket,omitempty" comment:"是否是websocket"`
	Hosts           []string               `json:"hosts,omitempty" comment:"域名"`
}

type Script struct {
	ID     string      `json:"id"`
	Script interface{} `json:"script,omitempty"`
}

type RequestValidation struct {
	Type       string      `json:"type,omitempty"`
	Required   []string    `json:"required,omitempty"`
	Properties interface{} `json:"properties,omitempty"`
}

// swagger:model GlobalPlugins
type GlobalPlugins struct {
	BaseInfo
	Plugins map[string]interface{} `json:"plugins"`
}

type ServerInfo struct {
	BaseInfo
	LastReportTime int64  `json:"last_report_time,omitempty"`
	UpTime         int64  `json:"up_time,omitempty"`
	BootTime       int64  `json:"boot_time,omitempty"`
	EtcdVersion    string `json:"etcd_version,omitempty"`
	Hostname       string `json:"hostname,omitempty"`
	Version        string `json:"version,omitempty"`
}

// swagger:model GlobalPlugins
type PluginConfig struct {
	BaseInfo
	Desc    string                 `json:"desc,omitempty" validate:"max=256"`
	Plugins map[string]interface{} `json:"plugins"`
	Labels  map[string]string      `json:"labels,omitempty"`
}

// swagger:model Proto
type Proto struct {
	BaseInfo
	Desc    string `json:"desc,omitempty"`
	Content string `json:"content"`
}

// swagger:model StreamRoute
type StreamRoute struct {
	BaseInfo
	Desc       string                 `json:"desc,omitempty"`
	RemoteAddr string                 `json:"remote_addr,omitempty"`
	ServerAddr string                 `json:"server_addr,omitempty"`
	ServerPort int                    `json:"server_port,omitempty"`
	SNI        string                 `json:"sni,omitempty"`
	Upstream   *UpstreamDef           `json:"upstream,omitempty"`
	UpstreamID interface{}            `json:"upstream_id,omitempty"`
	Plugins    map[string]interface{} `json:"plugins,omitempty"`
}

// swagger:model SystemConfig
type SystemConfig struct {
	ConfigName string                 `json:"config_name"`
	Desc       string                 `json:"desc,omitempty"`
	Payload    map[string]interface{} `json:"payload,omitempty"`
	CreateTime int64                  `json:"create_time,omitempty"`
	UpdateTime int64                  `json:"update_time,omitempty"`
}
