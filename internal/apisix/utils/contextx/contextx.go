// Package contextx
//
// @author: xwc1125
package contextx

import "context"

type ContextX struct {
	Ctx    context.Context
	Cancel context.CancelFunc
}

func NewContextX(ctx context.Context) *ContextX {
	ctx1, cancelFunc := context.WithCancel(ctx)
	return NewContextXWithCancel(ctx1, cancelFunc)
}

func NewContextXWithCancel(ctx context.Context, cancel context.CancelFunc) *ContextX {
	return &ContextX{
		Ctx:    ctx,
		Cancel: cancel,
	}
}
