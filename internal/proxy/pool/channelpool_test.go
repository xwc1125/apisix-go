// Copyright 2018 The yeqown Author. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.
package pool

import (
	"testing"

	"github.com/xwc1125/apisix-go/internal/apisix/core/entity"
	"github.com/xwc1125/apisix-go/internal/proxy"
)

func Test_chanPool(t *testing.T) {
	factoty := func(addr string) (*proxy.Proxy, error) {
		p, _ := proxy.NewProxy(entity.Route{}, nil)
		return p, nil
	}

	pool, err := NewChanPool(5, 100, factoty)
	if err != nil {
		t.Fatalf("could not make chan pool: %v", err)
	}

	t.Logf("len of pool is %d", pool.Len())

	p, err := pool.Get("localhost:8080")
	if err != nil {
		t.Fatalf("could not make chan pool: %v", err)
	}

	if p == nil {
		t.Fatalf("could not get one proxy form pool, proxy is nil")
	}

	client, _ := p.GetClient(nil)
	t.Logf("proxy addr: %v and addr is: %s", p, client.Addr)
}

func BenchmarkNewReverseProxyWithPool(b *testing.B) {
	b.StopTimer()
	pool, err := NewChanPool(10, 100, func(addr string) (*proxy.Proxy, error) {
		return proxy.NewProxy(entity.Route{})
	})
	if err != nil {
		b.Fatal(err)
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		proxy, err := pool.Get("locahost:8080")
		if err != nil {
			b.Fatal(err)
		}
		if proxy == nil {
			b.Fatalf("could not get from pool, proxy is nil")
		}
		client, _ := proxy.GetClient(nil)
		if client == nil {
			b.Fatalf("could not get from pool, client is nil")
		}
	}
}
