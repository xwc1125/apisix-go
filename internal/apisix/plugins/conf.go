package plugins

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/api7/ext-plugin-proto/go/A6"
	pc "github.com/api7/ext-plugin-proto/go/A6/PrepareConf"
	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/jellydator/ttlcache/v2"
)

var (
	cache *ConfCache
)

type ConfInfo struct {
	Name  string // 插件名
	Value string // 插件配置信息
}

type ConfEntry struct {
	Name  string      // 插件名
	Value interface{} // 插件配置信息对象
}

type RuleConf []ConfEntry

type ConfCache struct {
	lock     sync.Mutex
	keyCache *ttlcache.Cache
}

func newConfCache(ttl time.Duration) *ConfCache {
	cc := &ConfCache{}
	for _, c := range []**ttlcache.Cache{&cc.keyCache} {
		cache := ttlcache.NewCache()
		err := cache.SetTTL(ttl)
		if err != nil {
			log().Fatal("failed to set global ttl for cache", "err", err)
		}
		cache.SkipTTLExtensionOnHit(false)
		*c = cache
	}
	return cc
}

func (cc *ConfCache) Set(key string, pluginsConf map[string]interface{}) (string, error) {
	cc.lock.Lock()
	defer cc.lock.Unlock()

	if key != "" {
		_, err := cc.keyCache.Get(key)
		if err == nil {
			return key, nil
		}

		if err != ttlcache.ErrNotFound {
			log().Error("failed to get cached token with key", "err", err)
			// recreate the token
		}
	}

	var (
		entries = RuleConf{}
	)

	for name, info := range pluginsConf {
		plugin := findPlugin(name)
		if plugin == nil {
			log().Warn("can't find plugin, skip", "name", name)
			continue
		}

		log().Debug("prepare conf for plugin", "name", name)
		var (
			val []byte
			err error
		)
		val, err = json.Marshal(info)
		conf, err := plugin.ParseConf(val) // 解析value中的数据，即配置信息
		if err != nil {
			log().Error(
				"failed to parse configuration for plugin",
				"name", name, "configuration", val, "err", err)
			continue
		}
		if conf != nil {
			entries = append(entries, ConfEntry{
				Name:  name,
				Value: conf,
			})
		}
	}

	err := cc.keyCache.Set(key, entries)
	if err != nil {
		return "", err
	}
	return key, err
}

func (cc *ConfCache) SetIn(key string, entries RuleConf) error {
	cc.lock.Lock()
	defer cc.lock.Unlock()
	return cc.keyCache.Set(key, entries)
}

func (cc *ConfCache) Get(key string) (RuleConf, error) {
	res, err := cc.keyCache.Get(key)
	if err != nil {
		return nil, err
	}
	return res.(RuleConf), err
}

func (cc *ConfCache) Delete(key string) error {
	cc.lock.Lock()
	defer cc.lock.Unlock()
	return cc.keyCache.Remove(key)
}

func InitConfCache(ttl time.Duration) *ConfCache {
	cache = newConfCache(ttl)
	return cache
}

// PrepareConf 将请求的数据进行缓存
func PrepareConf(key string, pluginsConf map[string]interface{}) (string, error) {
	return cache.Set(key, pluginsConf)
}

// GetRuleConf 从缓存中取数据
func GetRuleConf(key string) (RuleConf, error) {
	return cache.Get(key)
}

// ConfToProto plugin conf to proto bytes
func ConfToProto(pluginName string, pluginConf string) []byte {
	builder := flatbuffers.NewBuilder(1024)

	name := builder.CreateString(pluginName)  // 插件名称
	value := builder.CreateString(pluginConf) // 插件的配置

	// 生成proto
	A6.TextEntryStart(builder)           // entry的开始
	A6.TextEntryAddName(builder, name)   // entry添加名称
	A6.TextEntryAddValue(builder, value) // entry添加值
	te := A6.TextEntryEnd(builder)       // entry的结束
	pc.ReqStartConfVector(builder, 1)    // 设置entry元素个数

	builder.PrependUOffsetT(te)
	v := builder.EndVector(1)

	pc.ReqStart(builder)
	pc.ReqAddConf(builder, v)
	root := pc.ReqEnd(builder)
	builder.Finish(root)
	return builder.FinishedBytes()
}

func CreateRuleConf(uniqueKey string, entries []ConfInfo) []byte {
	builder := flatbuffers.NewBuilder(1024)
	key := builder.CreateString(uniqueKey)

	entriesLen := len(entries)
	var textEntries = make([]flatbuffers.UOffsetT, 0, entriesLen)
	for _, entry := range entries {
		name := builder.CreateString(entry.Name)   // 插件名称
		value := builder.CreateString(entry.Value) // 插件的配置

		// 生成proto
		A6.TextEntryStart(builder)           // entry的开始
		A6.TextEntryAddName(builder, name)   // entry添加名称
		A6.TextEntryAddValue(builder, value) // entry添加值
		te := A6.TextEntryEnd(builder)       // entry的结束
		textEntries = append(textEntries, te)
	}
	pc.ReqStartConfVector(builder, entriesLen) // 设置entry元素个数
	for i := entriesLen - 1; i >= 0; i-- {
		offsetT := textEntries[i]
		builder.PrependUOffsetT(offsetT)
	}
	confVec := builder.EndVector(entriesLen)

	pc.ReqStart(builder)
	pc.ReqAddKey(builder, key)
	if confVec != 0 {
		pc.ReqAddConf(builder, confVec)
	}
	root := pc.ReqEnd(builder)
	builder.Finish(root)
	return builder.FinishedBytes()
}
