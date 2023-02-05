// Package lb
//
// @author: xwc1125
package lb

import (
	"hash/fnv"

	"github.com/valyala/fasthttp"
	"github.com/xwc1125/apisix-go/internal/apisix/core/entity"
)

type HashCBalance struct {
	upstreamDef *entity.UpstreamDef
	weights     []W
}

// newHashCBalance
func newHashCBalance(upstreamDef *entity.UpstreamDef, weights []W) LoadBalance {
	lb := HashCBalance{
		upstreamDef: upstreamDef,
		weights:     weights,
	}
	return lb
}

func (lb HashCBalance) Distribute(req *fasthttp.Request) int64 {
	l := len(lb.weights)
	if 0 >= l {
		return 0
	}
	hash := fnv.New32a()
	hashOn := lb.upstreamDef.HashOn
	key := lb.upstreamDef.Key

	var hashKey []byte
	switch hashOn {
	case "vars":
		// todo
		hashKey = req.PostArgs().Peek(key)
	case "header":
		hashKey = req.Header.Peek(key)
	case "cookie":
		// todo
	case "consumer":
		// todo
	case "vars_combinations":
		// todo
	}
	hash.Write(hashKey)
	return int64(hash.Sum32() % uint32(l))
}
