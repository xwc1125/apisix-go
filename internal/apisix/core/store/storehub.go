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

package store

import (
	"fmt"
	"reflect"

	"github.com/chain5j/chain5j-pkg/util/convutil"
	"github.com/chain5j/logger"
	"github.com/tidwall/gjson"
	"github.com/xwc1125/apisix-go/internal/apisix/core/entity"
	"github.com/xwc1125/apisix-go/internal/apisix/core/storage"
	"github.com/xwc1125/apisix-go/internal/apisix/utils/closer"
)

type HubKey string

const (
	HubKeyConsumer     HubKey = "consumer"
	HubKeyRoute        HubKey = "route"
	HubKeyService      HubKey = "service"
	HubKeySsl          HubKey = "ssl"
	HubKeyUpstream     HubKey = "upstream"
	HubKeyScript       HubKey = "script"
	HubKeyGlobalRule   HubKey = "global_rule"
	HubKeyServerInfo   HubKey = "server_info"
	HubKeyPluginConfig HubKey = "plugin_config"
	HubKeyProto        HubKey = "proto"
	HubKeyStreamRoute  HubKey = "stream_route"
)

var (
	storeHub = map[HubKey]*GenericStore{}
)

func InitStore(schema gjson.Result, key HubKey, opt GenericStoreOption) error {
	hubsNeedCheck := map[HubKey]bool{
		HubKeyConsumer:    true,
		HubKeyRoute:       true,
		HubKeySsl:         true,
		HubKeyService:     true,
		HubKeyUpstream:    true,
		HubKeyGlobalRule:  true,
		HubKeyStreamRoute: true,
	}

	if _, ok := hubsNeedCheck[key]; ok {
		validator, err := NewAPISIXJsonSchemaValidator(schema, "main."+string(key))
		if err != nil {
			return err
		}
		opt.Validator = validator
	}
	opt.HubKey = key
	s, err := NewGenericStore(opt)
	if err != nil {
		logger.Error("NewGenericStore error", "err", err)
		return err
	}
	if err := s.Init(); err != nil {
		logger.Error("GenericStore init error", "err", err)
		return err
	}

	closer.AppendToClosers(s.Close)
	storeHub[key] = s
	return nil
}

func GetStore(key HubKey) *GenericStore {
	if s, ok := storeHub[key]; ok {
		return s
	}
	panic(fmt.Sprintf("no store with key: %s", key))
}

func RangeStore(f func(key HubKey, store *GenericStore) bool) {
	for k, s := range storeHub {
		if k != "" && s != nil {
			if !f(k, s) {
				break
			}
		}
	}
}

func InitStores(schema gjson.Result, conf storage.EtcdConfig, watchEvents map[HubKey]WatchEvent) error {
	if watchEvents == nil {
		watchEvents = make(map[HubKey]WatchEvent)
	}
	err := InitStore(schema, HubKeyConsumer, GenericStoreOption{
		BasePath: conf.Prefix + "/consumers",
		ObjType:  reflect.TypeOf(entity.Consumer{}),
		KeyFunc: func(obj interface{}) string {
			r := obj.(*entity.Consumer)
			return r.Username
		},
		WatchEvent: watchEvents[HubKeyConsumer],
	})
	if err != nil {
		return err
	}

	err = InitStore(schema, HubKeyRoute, GenericStoreOption{
		BasePath: conf.Prefix + "/routes",
		ObjType:  reflect.TypeOf(entity.Route{}),
		KeyFunc: func(obj interface{}) string {
			r := obj.(*entity.Route)
			return convutil.ToString(r.ID)
		},
		WatchEvent: watchEvents[HubKeyRoute],
	})
	if err != nil {
		return err
	}

	err = InitStore(schema, HubKeyService, GenericStoreOption{
		BasePath: conf.Prefix + "/services",
		ObjType:  reflect.TypeOf(entity.Service{}),
		KeyFunc: func(obj interface{}) string {
			r := obj.(*entity.Service)
			return convutil.ToString(r.ID)
		},
		WatchEvent: watchEvents[HubKeyService],
	})
	if err != nil {
		return err
	}

	err = InitStore(schema, HubKeySsl, GenericStoreOption{
		BasePath: conf.Prefix + "/ssl",
		ObjType:  reflect.TypeOf(entity.SSL{}),
		KeyFunc: func(obj interface{}) string {
			r := obj.(*entity.SSL)
			return convutil.ToString(r.ID)
		},
		WatchEvent: watchEvents[HubKeySsl],
	})
	if err != nil {
		return err
	}

	err = InitStore(schema, HubKeyUpstream, GenericStoreOption{
		BasePath: conf.Prefix + "/upstreams",
		ObjType:  reflect.TypeOf(entity.Upstream{}),
		KeyFunc: func(obj interface{}) string {
			r := obj.(*entity.Upstream)
			return convutil.ToString(r.ID)
		},
		WatchEvent: watchEvents[HubKeyUpstream],
	})
	if err != nil {
		return err
	}

	err = InitStore(schema, HubKeyScript, GenericStoreOption{
		BasePath: conf.Prefix + "/scripts",
		ObjType:  reflect.TypeOf(entity.Script{}),
		KeyFunc: func(obj interface{}) string {
			r := obj.(*entity.Script)
			return r.ID
		},
		WatchEvent: watchEvents[HubKeyScript],
	})
	if err != nil {
		return err
	}

	err = InitStore(schema, HubKeyGlobalRule, GenericStoreOption{
		BasePath: conf.Prefix + "/global_rules",
		ObjType:  reflect.TypeOf(entity.GlobalPlugins{}),
		KeyFunc: func(obj interface{}) string {
			r := obj.(*entity.GlobalPlugins)
			return convutil.ToString(r.ID)
		},
		WatchEvent: watchEvents[HubKeyGlobalRule],
	})
	if err != nil {
		return err
	}

	err = InitStore(schema, HubKeyServerInfo, GenericStoreOption{
		BasePath: conf.Prefix + "/data_plane/server_info",
		ObjType:  reflect.TypeOf(entity.ServerInfo{}),
		KeyFunc: func(obj interface{}) string {
			r := obj.(*entity.ServerInfo)
			return convutil.ToString(r.ID)
		},
		WatchEvent: watchEvents[HubKeyServerInfo],
	})
	if err != nil {
		return err
	}

	err = InitStore(schema, HubKeyPluginConfig, GenericStoreOption{
		BasePath: conf.Prefix + "/plugin_configs",
		ObjType:  reflect.TypeOf(entity.PluginConfig{}),
		KeyFunc: func(obj interface{}) string {
			r := obj.(*entity.PluginConfig)
			return convutil.ToString(r.ID)
		},
		WatchEvent: watchEvents[HubKeyPluginConfig],
	})
	if err != nil {
		return err
	}

	err = InitStore(schema, HubKeyProto, GenericStoreOption{
		BasePath: conf.Prefix + "/proto",
		ObjType:  reflect.TypeOf(entity.Proto{}),
		KeyFunc: func(obj interface{}) string {
			r := obj.(*entity.Proto)
			return convutil.ToString(r.ID)
		},
		WatchEvent: watchEvents[HubKeyProto],
	})
	if err != nil {
		return err
	}

	err = InitStore(schema, HubKeyStreamRoute, GenericStoreOption{
		BasePath: conf.Prefix + "/stream_routes",
		ObjType:  reflect.TypeOf(entity.StreamRoute{}),
		KeyFunc: func(obj interface{}) string {
			r := obj.(*entity.StreamRoute)
			return convutil.ToString(r.ID)
		},
		WatchEvent: watchEvents[HubKeyStreamRoute],
	})
	if err != nil {
		return err
	}

	return nil
}
