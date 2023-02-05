-- 在头部引入所需要的模块
local ngx = ngx
local core = require("apisix.core")
local plugin = require("apisix.plugin")
local upstream = require("apisix.upstream")

-- 定义插件 schema 格式，需要对字段进行校验
local schema = {
    type = "object",
}

-- 插件元数据 schema
local metadata_schema = {
    type = "object",
    properties = {
        log_format = log_util.metadata_schema_log_format
    }
}

-- 声明插件名称
local plugin_name = "plugin-demo"

local _M = {
    version = 0.1, -- 插件版本
    priority = 11111, -- 插件优先级，同一阶段，优先级 ( priority ) 值大的插件，会优先执行
    name = plugin_name, -- 插件名称
    schema = schema, -- 插件schema
    metadata_schema = metadata_schema
}

-- 检查插件配置是否正确
function _M.check_schema(conf, schema_type)
    if schema_type == core.schema.TYPE_METADATA then
        return core.schema.check(metadata_schema, conf)
    end
    return core.schema.check(schema, conf)
end

-- 初始化
function _M.init()
end

-- destroy
function _M.destroy()
end

-- 可以拿到请求url, 请求头, 请求体
-- 可以修改响应状态码, 响应体
-- 在 APISIX，只有认证逻辑可以在 rewrite 阶段里面完成，
-- 其他需要在代理到上游之前执行的逻辑都是在 access 阶段完成的。
function _M.rewrite(conf, ctx)
end

-- 同rewrite
function _M.access(conf, ctx)
end

-- 不可以获得响应体
-- 不可以修改响应体
-- 可以获得, 修改响应头, 响应状态码
function _M.header_filter(ctx)
    core.log.warn("hit header_filter phase")
end

-- 可以获得响应体
-- 可以修改响应体
-- 可以获得, 修改响应头, 响应状态码
-- 不可以修改响应头, 响应状态码
function _M.body_filter(ctx)
    core.log.warn("hit body_filter phase")
end

-- 日志阶段
function _M.log(conf, ctx)
    core.log.warn("conf: ", core.json.encode(conf))
    core.log.warn("ctx: ", core.json.encode(ctx, true))
end

return _M
