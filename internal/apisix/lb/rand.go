package lb

import (
	"github.com/valyala/fasthttp"
	"github.com/valyala/fastrand"
	"github.com/xwc1125/apisix-go/internal/apisix/core/entity"
)

// RandBalance is rand loadBalance impl
type RandBalance struct {
	upstreamDef *entity.UpstreamDef
	weights     []W
}

// newRandBalance create a RandBalance
func newRandBalance(upstreamDef *entity.UpstreamDef, weights []W) LoadBalance {
	lb := RandBalance{
		upstreamDef: upstreamDef,
		weights:     weights,
	}
	return lb
}

func (rb RandBalance) Distribute(req *fasthttp.Request) int64 {
	l := len(rb.weights)
	if 0 >= l {
		return 0
	}
	return int64(fastrand.Uint32n(uint32(l)))
}
