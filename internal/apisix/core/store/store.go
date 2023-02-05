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
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sort"
	"sync"
	"time"

	"github.com/chain5j/chain5j-pkg/util/convutil"
	"github.com/chain5j/logger"
	"github.com/xwc1125/apisix-go/internal/apisix/core/entity"
	"github.com/xwc1125/apisix-go/internal/apisix/core/params"
	storage2 "github.com/xwc1125/apisix-go/internal/apisix/core/storage"
	"github.com/xwc1125/apisix-go/internal/apisix/utils/runtime"
)

type Pagination struct {
	PageSize   int `json:"page_size" form:"page_size" auto_read:"page_size"`
	PageNumber int `json:"page" form:"page" auto_read:"page"`
}

type Interface interface {
	Type() HubKey
	Get(ctx context.Context, key string) (interface{}, error)                                // 通过key查询
	List(ctx context.Context, input ListInput) (*ListOutput, error)                          // 查询list
	Create(ctx context.Context, obj interface{}) (interface{}, error)                        // 创建
	Update(ctx context.Context, obj interface{}, createIfNotExist bool) (interface{}, error) // 更新
	BatchDelete(ctx context.Context, keys []string) error                                    // 批量删除
}

type GenericStore struct {
	Stg storage2.Interface

	cache sync.Map
	opt   GenericStoreOption

	cancel context.CancelFunc
}

type WatchEvent interface {
	WatchEventPut(key string, objPtr interface{})
	WatchEventDelete(key string)
}

type GenericStoreOption struct {
	BasePath   string
	ObjType    reflect.Type
	KeyFunc    func(obj interface{}) string
	StockCheck func(obj interface{}, stockObj interface{}) error
	Validator  Validator
	HubKey     HubKey

	WatchEvent WatchEvent
}

func NewGenericStore(opt GenericStoreOption) (*GenericStore, error) {
	if opt.BasePath == "" {
		logger.Error("base path empty")
		return nil, fmt.Errorf("base path can not be empty")
	}
	if opt.ObjType == nil {
		logger.Error("object type is nil")
		return nil, fmt.Errorf("object type can not be nil")
	}
	if opt.KeyFunc == nil {
		logger.Error("key func is nil")
		return nil, fmt.Errorf("key func can not be nil")
	}

	if opt.ObjType.Kind() == reflect.Ptr {
		opt.ObjType = opt.ObjType.Elem()
	}
	if opt.ObjType.Kind() != reflect.Struct {
		logger.Error("obj type is invalid")
		return nil, fmt.Errorf("obj type is invalid")
	}
	s := &GenericStore{
		opt: opt,
	}
	s.Stg = storage2.GenEtcdStorage()

	return s, nil
}

func (s *GenericStore) Init() error {
	lc, lcancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer lcancel()
	ret, err := s.Stg.List(lc, s.opt.BasePath)
	if err != nil {
		return err
	}
	for i := range ret {
		key := ret[i].Key[len(s.opt.BasePath)+1:]
		objPtr, err := s.StringToObjPtr(ret[i].Value, key)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, "Error occurred while initializing logical store: ", s.opt.BasePath)
			return err
		}

		s.cache.Store(s.opt.KeyFunc(objPtr), objPtr)
	}

	c, cancel := context.WithCancel(context.TODO())
	ch := s.Stg.Watch(c, s.opt.BasePath)
	go func() {
		defer runtime.HandlePanic()
		for event := range ch {
			if event.Canceled {
				logger.Warn("watch failed", "err", event.Error)
			}

			for i := range event.Events {
				key := event.Events[i].Key[len(s.opt.BasePath)+1:]
				switch event.Events[i].Type {
				case storage2.EventTypePut:
					objPtr, err := s.StringToObjPtr(event.Events[i].Value, key)
					if err != nil {
						logger.Warn("value convert to obj failed", "basePath", s.opt.BasePath, "key", key, "err", err)
						continue
					}
					logger.Debug("watch event put", "basePath", s.opt.BasePath, "key", key)
					s.cache.Store(key, objPtr)
					if s.opt.WatchEvent != nil {
						s.opt.WatchEvent.WatchEventPut(key, objPtr)
					}
				case storage2.EventTypeDelete:
					logger.Debug("watch event delete", "basePath", s.opt.BasePath, "key", key)
					s.cache.Delete(key)
					if s.opt.WatchEvent != nil {
						s.opt.WatchEvent.WatchEventDelete(key)
					}
				}
			}
		}
	}()
	s.cancel = cancel
	return nil
}

func (s *GenericStore) Type() HubKey {
	return s.opt.HubKey
}

func (s *GenericStore) Get(_ context.Context, key string) (interface{}, error) {
	ret, ok := s.cache.Load(key)
	if !ok {
		logger.Warn("data not found by key", "key", key)
		return nil, params.ErrNotFound
	}
	return ret, nil
}

type ListInput struct {
	Predicate func(obj interface{}) bool
	Format    func(obj interface{}) interface{}
	PageSize  int
	// start from 1
	PageNumber int
	Less       func(i, j interface{}) bool
}

type ListOutput struct {
	Rows      []interface{} `json:"rows"`
	TotalSize int           `json:"total_size"`
}

// NewListOutput returns JSON marshalling safe struct pointer for empty slice
func NewListOutput() *ListOutput {
	return &ListOutput{Rows: make([]interface{}, 0)}
}

var defLessFunc = func(i, j interface{}) bool {
	iBase := i.(entity.GetBaseInfo).GetBaseInfo()
	jBase := j.(entity.GetBaseInfo).GetBaseInfo()
	if iBase.CreateTime != jBase.CreateTime {
		return iBase.CreateTime < jBase.CreateTime
	}
	if iBase.UpdateTime != jBase.UpdateTime {
		return iBase.UpdateTime < jBase.UpdateTime
	}
	iID := convutil.ToString(iBase.ID)
	jID := convutil.ToString(jBase.ID)
	return iID < jID
}

func (s *GenericStore) List(_ context.Context, input ListInput) (*ListOutput, error) {
	var ret []interface{}
	// 从缓存中不断读取过滤
	s.cache.Range(func(key, value interface{}) bool {
		if input.Predicate != nil && !input.Predicate(value) {
			return true
		}
		if input.Format != nil {
			value = input.Format(value)
		}
		ret = append(ret, value)
		return true
	})

	// should return an empty array not a null for client
	if ret == nil {
		ret = []interface{}{}
	}

	output := &ListOutput{
		Rows:      ret,
		TotalSize: len(ret),
	}
	if input.Less == nil {
		input.Less = defLessFunc
	}

	sort.Slice(output.Rows, func(i, j int) bool {
		return input.Less(output.Rows[i], output.Rows[j])
	})

	if input.PageSize > 0 && input.PageNumber > 0 {
		skipCount := (input.PageNumber - 1) * input.PageSize
		if skipCount > output.TotalSize {
			output.Rows = []interface{}{}
			return output, nil
		}

		endIdx := skipCount + input.PageSize
		if endIdx >= output.TotalSize {
			output.Rows = ret[skipCount:]
			return output, nil
		}
		output.Rows = ret[skipCount:endIdx]
	}

	return output, nil
}

func (s *GenericStore) Range(_ context.Context, f func(key string, obj interface{}) bool) {
	s.cache.Range(func(key, value interface{}) bool {
		return f(key.(string), value)
	})
}

func (s *GenericStore) ingestValidate(obj interface{}) (err error) {
	if s.opt.Validator != nil {
		if err := s.opt.Validator.Validate(obj); err != nil {
			logger.Error("data validate failed", "err", err, "obj", obj)
			return err
		}
	}

	if s.opt.StockCheck != nil {
		s.cache.Range(func(key, value interface{}) bool {
			if err = s.opt.StockCheck(obj, value); err != nil {
				return false
			}
			return true
		})
	}
	return err
}

func (s *GenericStore) CreateCheck(obj interface{}) ([]byte, error) {

	if setter, ok := obj.(entity.GetBaseInfo); ok {
		info := setter.GetBaseInfo()
		info.Creating()
	}

	if err := s.ingestValidate(obj); err != nil {
		return nil, err
	}

	key := s.opt.KeyFunc(obj)
	if key == "" {
		return nil, fmt.Errorf("key is required")
	}
	_, ok := s.cache.Load(key)
	if ok {
		logger.Warn("key is conflicted", "key", key)
		return nil, fmt.Errorf("key: %s is conflicted", key)
	}

	bytes, err := json.Marshal(obj)
	if err != nil {
		logger.Error("json marshal failed", "err", err)
		return nil, fmt.Errorf("json marshal failed: %s", err)
	}

	return bytes, nil
}

func (s *GenericStore) Create(ctx context.Context, obj interface{}) (interface{}, error) {
	if setter, ok := obj.(entity.GetBaseInfo); ok {
		info := setter.GetBaseInfo()
		info.Creating()
	}

	bytes, err := s.CreateCheck(obj)
	if err != nil {
		return nil, err
	}

	if err := s.Stg.Create(ctx, s.GetObjStorageKey(obj), string(bytes)); err != nil {
		return nil, err
	}

	return obj, nil
}

func (s *GenericStore) Update(ctx context.Context, obj interface{}, createIfNotExist bool) (interface{}, error) {
	if err := s.ingestValidate(obj); err != nil {
		return nil, err
	}

	key := s.opt.KeyFunc(obj)
	if key == "" {
		return nil, fmt.Errorf("key is required")
	}
	storedObj, ok := s.cache.Load(key)
	if !ok {
		if createIfNotExist {
			return s.Create(ctx, obj)
		}
		logger.Warn("key is not found", "err", key)
		return nil, fmt.Errorf("key: %s is not found", key)
	}

	if setter, ok := obj.(entity.GetBaseInfo); ok {
		storedGetter := storedObj.(entity.GetBaseInfo)
		storedInfo := storedGetter.GetBaseInfo()
		info := setter.GetBaseInfo()
		info.Updating(storedInfo)
	}

	bs, err := json.Marshal(obj)
	if err != nil {
		logger.Error("json marshal failed", "err", err)
		return nil, fmt.Errorf("json marshal failed: %s", err)
	}
	if err := s.Stg.Update(ctx, s.GetObjStorageKey(obj), string(bs)); err != nil {
		return nil, err
	}

	return obj, nil
}

func (s *GenericStore) BatchDelete(ctx context.Context, keys []string) error {
	var storageKeys []string
	for i := range keys {
		storageKeys = append(storageKeys, s.GetStorageKey(keys[i]))
	}

	return s.Stg.BatchDelete(ctx, storageKeys)
}

func (s *GenericStore) Close() error {
	s.cancel()
	return nil
}

func (s *GenericStore) StringToObjPtr(str, key string) (interface{}, error) {
	objPtr := reflect.New(s.opt.ObjType)
	ret := objPtr.Interface()
	err := json.Unmarshal([]byte(str), ret)
	if err != nil {
		logger.Error("json unmarshal failed", "err", err)
		return nil, fmt.Errorf("json unmarshal failed\n\tRelated Key:\t\t%s\n\tError Description:\t%s", key, err)
	}

	if setter, ok := ret.(entity.GetBaseInfo); ok {
		info := setter.GetBaseInfo()
		info.KeyCompat(key)
	}

	return ret, nil
}

func (s *GenericStore) GetObjStorageKey(obj interface{}) string {
	return s.GetStorageKey(s.opt.KeyFunc(obj))
}

func (s *GenericStore) GetStorageKey(key string) string {
	return fmt.Sprintf("%s/%s", s.opt.BasePath, key)
}
