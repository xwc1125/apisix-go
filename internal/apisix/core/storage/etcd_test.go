// Package storage
//
// @author: xwc1125
package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/chain5j/logger"
	entity2 "github.com/xwc1125/apisix-go/internal/apisix/core/entity"
)

func TestNewETCDStorage(t *testing.T) {
	storage, err := NewETCDStorage(&EtcdConfig{
		Endpoints: []string{"127.0.0.1:2379"},
		Username:  "",
		Password:  "",
		Tls:       nil,
		Prefix:    "/apisix",
	})
	if err != nil {
		t.Fatal(err)
	}
	{
		val, err := storage.Get(context.Background(), "/apisix/routes/430024799805244101")
		if err != nil {
			t.Fatal(err)
		}
		t.Log("route", val)
	}
	{
		val, err := storage.List(context.Background(), "/apisix/routes")
		if err != nil {
			t.Fatal(err)
		}
		t.Log("routes", val)
	}
	{
		basePath := "/apisix/routes"
		typeOf := reflect.TypeOf(entity2.Route{})
		// 监听数据变化
		ch := storage.Watch(context.Background(), basePath)
		for event := range ch {
			if event.Canceled {
				logger.Warn("watch failed", "err", event.Error)
			}

			for i := range event.Events {
				switch event.Events[i].Type {
				case EventTypePut:
					key := event.Events[i].Key[len(basePath)+1:]
					objPtr, err := StringToObjPtr(typeOf, event.Events[i].Value, key)
					if err != nil {
						logger.Warn("value convert to obj failed", "err", err)
						continue
					}
					fmt.Println("put==>", objPtr)
					// s.cache.Store(key, objPtr)
				case EventTypeDelete:
					// s.cache.Delete(event.Events[i].Key[len(s.opt.BasePath)+1:])
					fmt.Println("del==>", event.Events[i].Key[len(basePath)+1:])
				}
			}
		}
	}

}

func StringToObjPtr(objType reflect.Type, str, key string) (interface{}, error) {
	objPtr := reflect.New(objType)
	ret := objPtr.Interface()
	err := json.Unmarshal([]byte(str), ret)
	if err != nil {
		logger.Error("json unmarshal failed", "err", err)
		return nil, fmt.Errorf("json unmarshal failed\n\tRelated Key:\t\t%s\n\tError Description:\t%s", key, err)
	}

	if setter, ok := ret.(entity2.GetBaseInfo); ok {
		info := setter.GetBaseInfo()
		info.KeyCompat(key)
	}

	return ret, nil
}
