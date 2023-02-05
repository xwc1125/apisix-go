# Oauth2-plugin

> 扩展Apisix的认证接口

# 插件使用方式
- 插件编译
```shell
# 构建项目 会在根目录生成go-runner文件
make build

```

- 插件引用

```shell
# 在apisix的 config.yaml 文件中添加
ext-plugin:
  cmd: ["/usr/local/apisix/plugin/finecloud/go-runner", "run"]
```

- 插件配置

```shell
curl http://127.0.0.1:9080/apisix/admin/routes/1 -H 'X-API-KEY: edd1c9f034335f136f87ad84b625c8f1' -X PUT -d '
{
  "uri": "/admin/user/info",
  "plugins": {
    "ext-plugin-pre-req": {
      "conf": [
        { "name": "Oauth2", "value":"{\"api_key\":\"app\",\"password\":\"app\",\"check_url\":\"http://127.0.0.1:9999/auth/oauth/check_token\"}"}
      ]
    }
  },
  "upstream": {
        "type": "roundrobin",
        "nodes": {
            "127.0.0.1:9999": 1
        }
    }
}
'
```
- 测试

```shell
# 无效请求
curl http://127.0.0.1:9080/get

# 有效请求(令牌过期)
curl --location --request GET 'http://127.0.0.1:9080/get' \
--header 'Authorization: Bearer 6a5a8f05-34a9-4a83-8deb-012892c8b08f'
```

## 参数说明

- api_key: 令牌
- password: 令牌密码
- check_url: token检查点url
- name: 扩展插件名
  /usr/local/apisix/plugin/finecloud/go-runner

## 注意

构建时 选择对应OS的架构生成`go-runner`