# apisix-go

## 简介

`apisix-go` 是一个试图通过使用golang来实现`apisix`功能的项目。

## 功能说明

### 插件列表

- [ ] real-ip `-->` priority: 23000
- [ ] client-control `-->` priority: 22000
- [ ] proxy-control `-->` priority: 21990
- [x] request-id `-->` priority: 12015 // todo 实现了uuid，雪花算法还未实现
- [ ] zipkin `-->` priority: 12011
- [ ] skywalking `-->` priority: 12010
- [ ] opentelemetry `-->` priority: 12009
- [ ] ext-plugin-pre-req `-->` priority: 12000
- [ ] fault-injection `-->` priority: 11000
- [ ] mocking `-->` priority: 10900
- [ ] serverless-pre-function `-->` priority: 10000
- [ ] batch-requests `-->` priority: 4010
- [x] cors `-->` priority: 4000
- [x] ip-restriction `-->` priority: 3000
- [ ] ua-restriction `-->` priority: 2999
- [ ] referer-restriction `-->` priority: 2990
- [ ] csrf `-->` priority: 2980
- [ ] uri-blocker `-->` priority: 2900
- [x] rpc-to-rest `-->` priority: 2801 // rpc api转为restful接口
- [ ] request-validation `-->` priority: 2800
- [ ] openid-connect `-->` priority: 2599
- [ ] authz-casbin `-->` priority: 2560
- [ ] authz-casdoor `-->` priority: 2559
- [ ] wolf-rbac `-->` priority: 2555
- [ ] ldap-auth `-->` priority: 2540
- [ ] hmac-auth `-->` priority: 2530
- [ ] basic-auth `-->` priority: 2520
- [x] oauth2 `-->` priority: 2515
- [ ] jwt-auth `-->` priority: 2510
- [ ] key-auth `-->` priority: 2500
- [ ] consumer-restriction `-->` priority: 2400
- [ ] forward-auth `-->` priority: 2002
- [ ] opa `-->` priority: 2001
- [ ] authz-keycloak `-->` priority: 2000
- [ ] error-log-logger `-->` priority: 1091
- [ ] proxy-mirror `-->` priority: 1010
- [ ] proxy-cache `-->` priority: 1009
- [x] proxy-rewrite `-->` priority: 1008
- [ ] api-breaker `-->` priority: 1005
- [ ] limit-conn `-->` priority: 1003
- [ ] limit-count `-->` priority: 1002
- [ ] limit-req `-->` priority: 1001
- [ ] node-status `-->` priority: 1000
- [ ] gzip `-->` priority: 995
- [ ] server-info `-->` priority: 990
- [ ] traffic-split `-->` priority: 966
- [ ] redirect `-->` priority: 900
- [ ] response-rewrite `-->` priority: 899
- [ ] kafka-proxy `-->` priority: 508
- [ ] dubbo-proxy `-->` priority: 507
- [ ] grpc-transcode `-->` priority: 506
- [ ] grpc-web `-->` priority: 505
- [ ] public-api `-->` priority: 501
- [ ] prometheus `-->` priority: 500
- [ ] datadog `-->` priority: 495
- [ ] echo `-->` priority: 412
- [ ] loggly `-->` priority: 411
- [ ] http-logger `-->` priority: 410
- [ ] splunk-hec-logging `-->` priority: 409
- [ ] skywalking-logger `-->` priority: 408
- [ ] google-cloud-logging `-->` priority: 407
- [ ] sls-logger `-->` priority: 406
- [ ] tcp-logger `-->` priority: 405
- [ ] kafka-logger `-->` priority: 403
- [ ] rocketmq-logger `-->` priority: 402
- [ ] syslog `-->` priority: 401
- [ ] udp-logger `-->` priority: 400
- [ ] file-logger `-->` priority: 399
- [ ] clickhouse-logger `-->` priority: 398
- [ ] log-rotate `-->` priority: 100
  <-- recommend to use priority (0, 100) for your custom plugins
- [ ] example-plugin `-->` priority: 0
- [ ] aws-lambda `-->` priority: -1899
- [ ] azure-functions `-->` priority: -1900
- [ ] openwhisk `-->` priority: -1901
- [ ] serverless-post-function `-->` priority: -2000
- [ ] ext-plugin-post-req `-->` priority: -3000
- [ ] ext-plugin-post-resp `-->` priority: -4000

## 使用

```
apisix-go server -c=conf/config.yaml
```

## 证书

`apisix-go` 的源码允许用户在遵循 [Apache 2.0 开源证书](LICENSE) 规则的前提下使用。

## 版权

Copyright@2023 xwc1125

![xwc1125](./logo.png)