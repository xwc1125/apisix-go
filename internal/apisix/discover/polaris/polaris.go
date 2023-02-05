// Package polaris
//
// @author: xwc1125
package polaris

import (
	"encoding/binary"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/chain5j/logger"
	"github.com/mitchellh/mapstructure"
	"github.com/polarismesh/polaris-go/api"
	"github.com/polarismesh/polaris-go/pkg/config"
	"github.com/polarismesh/polaris-go/pkg/model"
	"github.com/valyala/fasthttp"
	"github.com/xwc1125/apisix-go/internal/apisix/core/store"
	"github.com/xwc1125/apisix-go/internal/apisix/discover"
	"github.com/xwc1125/apisix-go/internal/apisix/plugins"
)

var (
	_ discover.Discover = new(Polaris)

	consumerMapLock sync.Mutex
	consumerMap     = make(map[string]api.ConsumerAPI) // addr-->api.ConsumerAPI
)

func init() {
	name := "polaris"
	p := &Polaris{
		log:     logger.Log(name),
		name:    name,
		version: "0.1",
	}
	var err error
	if p.validator, err = store.NewSchemaValidator(p.schema()); err != nil {
		p.log.Error(p.schema()+" new schema validator err", "err", err)
		return
	}
	if err = discover.RegisterDiscover(p); err != nil {
		p.log.Error("failed to register plugin"+p.Name(), "err", err)
	}
}

type Polaris struct {
	log       logger.Logger
	validator store.Validator

	plugins.DefaultPlugin
	name    string
	version string
}

type PolarisConf struct {
	// 初始化consumer的参数
	ServerAddr string `json:"server_addr" mapstructure:"server_addr" yaml:"server_addr"` // 服务器地址
	LBPolicy   string `json:"lb_policy" mapstructure:"lb_policy" yaml:"lb_policy"`       // lb策略

	// 获取实例的参数
	Namespace      string `json:"namespace" mapstructure:"namespace" yaml:"namespace"`                   // 命名空间
	Service        string `json:"service" mapstructure:"service" yaml:"service"`                         // 服务名
	Token          string `json:"token" mapstructure:"token" yaml:"token"`                               // 可选，token
	Timeout        int64  `json:"timeout" mapstructure:"timeout" yaml:"timeout"`                         // 可选，单次查询超时时间，默认直接获取全局的超时配置
	RetryCount     int    `json:"retry_count" mapstructure:"retry_count" yaml:"retry_count"`             // 可选，重试次数，默认直接获取全局的超时配置
	ReplicateCount int    `json:"replicate_count" mapstructure:"replicate_count" yaml:"replicate_count"` // 可选，备份节点数,对于一致性hash等有状态的负载均衡方式
	// LbPolicy   string `json:"lb_policy" mapstructure:"lb_policy" yaml:"lb_policy"`   // 可选，负载均衡算法
	Canary                    string `json:"canary" mapstructure:"canary" yaml:"canary"`                                                                      // 可选，金丝雀
	EnableFailOverDefaultMeta bool   `json:"enable_fail_over_default_meta" mapstructure:"enable_fail_over_default_meta" yaml:"enable_fail_over_default_meta"` // 是否开启元数据匹配不到时启用自定义匹配规则，仅用于dstMetadata路由插件
	// Metadata                  map[string]string `json:"metadata" mapstructure:"metadata" yaml:"metadata"`                                                                // 可选，元数据信息，仅用于dstMetadata路由插件的过滤
}

func (p *Polaris) Name() string {
	return p.name
}

func (p *Polaris) Version() string {
	return p.version
}

// 创建consumerAPI实例
// 注意该实例所有方法都是协程安全，一般用户进程只需要创建一个consumerAPI,重复使用即可
// 切勿每次调用之前都创建一个consumerAPI
func getConsumer(conf PolarisConf) (api.ConsumerAPI, error) {
	consumerMapLock.Lock()
	defer consumerMapLock.Unlock()
	addr := conf.ServerAddr
	if consumer, ok := consumerMap[addr]; ok {
		return consumer, nil
	}
	cfg := api.NewConfiguration()
	// 设置负载均衡算法为maglev
	if len(conf.LBPolicy) > 0 {
		cfg.GetConsumer().GetLoadbalancer().SetType(conf.LBPolicy)
	} else {
		cfg.GetConsumer().GetLoadbalancer().SetType(api.LBPolicyL5CST)
	}
	addr = strings.Replace(addr, "ip://", "", 1)
	addr = strings.Replace(addr, "l5://", "", 1)
	consumer, err := api.NewConsumerAPIByAddress(addr)

	consumerMap[conf.ServerAddr] = consumer
	return consumer, err
}

func (p *Polaris) GetClient(args map[string]string) (client *fasthttp.HostClient, err error) {
	conf := PolarisConf{}
	// 将 map 转换为指定的结构体
	if err := mapstructure.Decode(args, &conf); err != nil {
		p.log.Error("json unmarshal conf err", "err", err)
		return nil, err
	}
	if len(conf.LBPolicy) == 0 {
		conf.LBPolicy = api.LBPolicyL5CST
	}

	// Validate
	err = p.validator.Validate(conf)
	if err != nil {
		p.log.Error("validate conf err", "err", err)
		return nil, err
	}

	// 业务操作
	consumer, err := getConsumer(conf)
	if err != nil {
		p.log.Error("get consumer err", "err", err)
		return nil, err
	}

	hashKey := make([]byte, 8)
	binary.LittleEndian.PutUint64(hashKey, 123)
	var flowId uint64

	var (
		timeout    time.Duration
		retryCount *int
	)
	if conf.Timeout > 0 {
		timeout = time.Duration(conf.Timeout) * time.Second
	}
	if conf.RetryCount > 0 {
		retryCount = &conf.RetryCount
	}

	var getInstancesReq = &api.GetOneInstanceRequest{
		GetOneInstanceRequest: model.GetOneInstanceRequest{
			FlowID:    atomic.AddUint64(&flowId, 1),
			Namespace: conf.Namespace,
			Service:   conf.Service,
			// Metadata:                  conf.Metadata,
			EnableFailOverDefaultMeta: conf.EnableFailOverDefaultMeta,
			HashKey:                   hashKey,
			Timeout:                   &timeout,
			RetryCount:                retryCount,
			ReplicateCount:            conf.ReplicateCount,
			LbPolicy:                  config.DefaultLoadBalancerHash,
			Canary:                    conf.Canary,
		},
	}

	startTime := time.Now()
	// 进行服务发现，获取单一服务实例
	getInstResp, err := consumer.GetOneInstance(getInstancesReq)
	if nil != err {
		p.log.Error("fail to sync GetOneInstance", "err", err)
		return nil, err
	}
	consumeDuration := time.Since(startTime)
	targetInstance := getInstResp.Instances[0]
	// 构造请求，进行服务调用结果上报
	svcCallResult := &api.ServiceCallResult{
		ServiceCallResult: model.ServiceCallResult{
			EmptyInstanceGauge: model.EmptyInstanceGauge{},
			CalledInstance:     targetInstance, // 设置被调的实例信息
			Method:             "",
			RetStatus:          api.RetSuccess,   // 设置服务调用结果，枚举，成功或者失败
			RetCode:            nil,              // 设置服务调用返回码
			Delay:              &consumeDuration, // 设置服务调用时延信息
		},
	}
	// 设置服务调用返回码
	svcCallResult.SetRetCode(0)
	// 执行调用结果上报
	err = consumer.UpdateServiceCallResult(svcCallResult)
	if nil != err {
		p.log.Error("fail to UpdateServiceCallResult", "err", err)
		return nil, err
	}
	hostClient := &fasthttp.HostClient{
		Addr: fmt.Sprintf("%s:%d", targetInstance.GetHost(), targetInstance.GetPort()),
		Name: "",
		// IsTLS:        config.SslVerify,
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
	}
	return hostClient, nil
}

func (p *Polaris) schema() string {
	return `
 {
  "$comment": "this is a mark for our injected discover schema",
  "properties": {
    "server_addr": {
     "pattern": "^[0-9]{1,3}.[0-9]{1,3}.[0-9]{1,3}.[0-9]{1,3}:[0-9]{1,5}$",
	  "type": "string"
    },
    "lb_policy": {
      "maxLength": 1024,
      "minLength": 1,
      "type": "string"
    },
    "namespace": {
      "maxLength": 1024,
      "minLength": 1,
      "type": "string"
    },
    "service": {
      "maxLength": 1024,
      "minLength": 1,
      "type": "string"
    },
    "token": {
      "maxLength": 1024,
      "type": "string"
    },
    "timeout": {
      "default": 5,
      "type": "integer"
    },
    "retry_count": {
      "type": "integer"
    },
    "replicate_count": {
      "type": "integer"
    },
    "canary": {
      "maxLength": 1024,
      "type": "string"
    },
    "enable_fail_over_default_meta": {
      "default": false,
      "type": "boolean"
    }
  },
  "required": [
     "server_addr",
     "namespace",
     "service"
   ],
  "type": "object"
}
`
}
