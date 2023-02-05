package lb

import (
	"hash/fnv"

	"github.com/valyala/fasthttp"
	"github.com/xwc1125/apisix-go/internal/apisix/core/entity"
	"github.com/xwc1125/apisix-go/internal/apisix/utils/iputils"
)

// HashIPBalance is hash IP loadBalance impl
type HashIPBalance struct {
	upstreamDef *entity.UpstreamDef
	weights     []W
}

// newHashIPBalance create a HashIPBalance
func newHashIPBalance(upstreamDef *entity.UpstreamDef, weights []W) LoadBalance {
	lb := HashIPBalance{
		upstreamDef: upstreamDef,
		weights:     weights,
	}
	return lb
}

// Distribute select a server from servers using HashIPBalance
func (lb HashIPBalance) Distribute(ctx *fasthttp.Request) int64 {
	l := len(lb.weights)
	if 0 >= l {
		return 0
	}
	hash := fnv.New32a()
	// key is client ip
	key := iputils.ClientIP(ctx)
	hash.Write([]byte(key))
	return int64(hash.Sum32() % uint32(l))
}
