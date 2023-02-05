package lb

import (
	"github.com/valyala/fasthttp"
	"github.com/xwc1125/apisix-go/internal/apisix/core/entity"
)

const (
	RoundRobin = "roundrobin"
	IPHash     = "iphash"
	CHash      = "chash"
	Rand       = "Rand"
)

// LoadBalance .
type LoadBalance interface {
	Distribute(ctx *fasthttp.Request) int64
}

// W ...
type W interface {
	Weight() int
}

// Weight .
type Weight uint

func (w Weight) Weight() int {
	return int(w)
}

// NewBalancer ...
func NewBalancer(upstreamDef *entity.UpstreamDef, ws []W) LoadBalance {
	switch upstreamDef.Type {
	case RoundRobin:
		return newRoundRobin(upstreamDef, ws)
	case CHash:
		return newHashCBalance(upstreamDef, ws)
	case IPHash:
		return newHashIPBalance(upstreamDef, ws)
	case Rand:
		return newRandBalance(upstreamDef, ws)
	default:
		return newRoundRobin(upstreamDef, ws)
	}
}
