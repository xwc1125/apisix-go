package polaris

import (
	"encoding/binary"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/polarismesh/polaris-go/api"
	"github.com/polarismesh/polaris-go/pkg/config"
	plog "github.com/polarismesh/polaris-go/pkg/log"
	"github.com/polarismesh/polaris-go/pkg/model"
	"github.com/valyala/fasthttp"
)

const logLevel = plog.InfoLog

func TestPolaris(t *testing.T) {
	go func() {
		log.Println(http.ListenAndServe("localhost:6070", nil))
	}()
	var err error
	seconds := 30
	err = plog.GetBaseLogger().SetLogLevel(logLevel)
	if nil != err {
		log.Fatalf("fail to SetLogLevel, err is %v", err)
	}
	cfg := api.NewConfiguration()
	// 设置负载均衡算法为maglev
	cfg.GetConsumer().GetLoadbalancer().SetType(api.LBPolicyL5CST)
	// 创建consumerAPI实例
	// 注意该实例所有方法都是协程安全，一般用户进程只需要创建一个consumerAPI,重复使用即可
	// 切勿每次调用之前都创建一个consumerAPI
	// consumer, err := api.NewConsumerAPIByConfig(cfg)
	consumer, err := api.NewConsumerAPIByAddress("9.135.57.175:8091")
	if nil != err {
		log.Fatalf("fail to create ConsumerAPI by default configuration, err is %v", err)
	}
	defer consumer.Destroy()

	var token = "IfgzsjSr9w77BaYAdwHd12LiWF4KL28xjmCbQpsbjxMspn2qIWZB0Pd8atYpUgSu8u69LXi15/NvVOgPI5Y="
	deadline := time.Now().Add(time.Duration(seconds) * time.Second)
	hashKey := make([]byte, 8)
	binary.LittleEndian.PutUint64(hashKey, 123)
	for {
		if time.Now().After(deadline) {
			break
		}
		var flowId uint64
		var getInstancesReq = &api.GetOneInstanceRequest{
			GetOneInstanceRequest: model.GetOneInstanceRequest{
				FlowID:    atomic.AddUint64(&flowId, 1),
				Service:   "anxin-nft-auth",
				Namespace: "anxin-nft-dev",
				Metadata: map[string]string{
					"token": token,
				},
				EnableFailOverDefaultMeta: false,
				FailOverDefaultMeta:       model.FailOverDefaultMetaConfig{},
				HashKey:                   hashKey,
				HashValue:                 0,
				SourceService:             nil,
				Timeout:                   nil,
				RetryCount:                nil,
				ReplicateCount:            0,
				LbPolicy:                  config.DefaultLoadBalancerHash,
				Canary:                    "",
			},
		}
		startTime := time.Now()
		// 进行服务发现，获取单一服务实例
		getInstResp, err := consumer.GetOneInstance(getInstancesReq)
		if nil != err {
			log.Fatalf("fail to sync GetOneInstance, err is %v", err)
		}
		consumeDuration := time.Since(startTime)
		log.Printf("success to sync GetOneInstance by maglev hash, count is %d, consume is %v\n",
			len(getInstResp.Instances), consumeDuration)
		targetInstance := getInstResp.Instances[0]
		log.Printf("sync instance is id=%s, address=%s:%d\n", targetInstance.GetId(), targetInstance.GetHost(), targetInstance.GetPort())
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
			log.Fatalf("fail to UpdateServiceCallResult, err is %v", err)
		}

		bytes, err := Http(token, fmt.Sprintf("%s:%d", targetInstance.GetHost(), targetInstance.GetPort()), "/discovery/service/callee/sum?value1=1&value2=2")
		if err != nil {
			log.Fatal(err)
		}
		log.Println(string(bytes))
		time.Sleep(1 * time.Second)
	}
	log.Printf("success to sync get one instance")

}

func Http(token string, addr, reqUri string) ([]byte, error) {
	// if strings.HasPrefix(service, "ip://") {
	// 	addr.addressType = addressTypeIP
	// 	addr.addr = strings.TrimPrefix(service, "ip://")
	// 	return addr, nil
	// } else if strings.HasPrefix(service, "l5://") {
	// 	service = strings.TrimPrefix(service, "l5://")
	// }

	// response, err := http.Get(fmt.Sprintf("http://%s%s", addr, reqUri))
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// bytes, err := io.ReadAll(response.Body)
	// fmt.Println(string(bytes))
	timeoutDura := time.Duration(5) * time.Second
	client := fasthttp.HostClient{
		Addr: addr,
		Name: "",
		// IsTLS:        config.SslVerify,
		ReadTimeout:  timeoutDura,
		WriteTimeout: timeoutDura,
	}
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)
	defer fasthttp.ReleaseRequest(req)
	req.SetRequestURI(fmt.Sprintf("http://%s%s", addr, reqUri))
	// req.Header.Set(`Bearer `+xgateway.HeaderAuthorization, token)
	err := client.DoTimeout(req, resp, timeoutDura)
	if err != nil {
		return nil, err
	}
	return resp.Body(), nil
}

func convertQuery(rawQuery string) map[string]string {
	meta := make(map[string]string)
	if len(rawQuery) == 0 {
		return meta
	}
	tokens := strings.Split(rawQuery, "&")
	if len(tokens) > 0 {
		for _, token := range tokens {
			values := strings.Split(token, "=")
			meta[values[0]] = values[1]
		}
	}
	return meta
}
