// Package entity
//
// @author: xwc1125
package entity

import (
	"github.com/api7/ext-plugin-proto/go/A6"
	hrc "github.com/api7/ext-plugin-proto/go/A6/HTTPReqCall"
	flatbuffers "github.com/google/flatbuffers/go"
)

type Pair struct {
	Name  string
	Value string
}

type ReqOpt struct {
	SrcIP      []byte // 目标IP
	Method     A6.Method
	Path       string // 目标的URL
	Headers    []Pair
	RespHeader []Pair
	Args       []Pair
}

func BuildRequestOpt(reqId uint32, confToken uint32, opt ReqOpt) []byte {
	builder := flatbuffers.NewBuilder(1024)

	var ip flatbuffers.UOffsetT
	if len(opt.SrcIP) > 0 {
		ip = builder.CreateByteVector(opt.SrcIP)
	}

	var path flatbuffers.UOffsetT
	if opt.Path != "" {
		path = builder.CreateString(opt.Path)
	}

	hdrLen := len(opt.Headers)
	var hdrVec, respHdrVec flatbuffers.UOffsetT
	if hdrLen > 0 {
		hdrs := []flatbuffers.UOffsetT{}
		for _, v := range opt.Headers {
			name := builder.CreateString(v.Name)
			value := builder.CreateString(v.Value)
			A6.TextEntryStart(builder)
			A6.TextEntryAddName(builder, name)
			A6.TextEntryAddValue(builder, value)
			te := A6.TextEntryEnd(builder)
			hdrs = append(hdrs, te)
		}
		size := len(hdrs)
		hrc.RewriteStartHeadersVector(builder, size)
		for i := size - 1; i >= 0; i-- {
			te := hdrs[i]
			builder.PrependUOffsetT(te)
		}
		hdrVec = builder.EndVector(size)
	}

	if len(opt.RespHeader) > 0 {
		respHdrs := []flatbuffers.UOffsetT{}
		for _, v := range opt.Headers {
			name := builder.CreateString(v.Name)
			value := builder.CreateString(v.Value)
			A6.TextEntryStart(builder)
			A6.TextEntryAddName(builder, name)
			A6.TextEntryAddValue(builder, value)
			te := A6.TextEntryEnd(builder)
			respHdrs = append(respHdrs, te)
		}
		size := len(respHdrs)
		hrc.RewriteStartRespHeadersVector(builder, size)
		for i := size - 1; i >= 0; i-- {
			te := respHdrs[i]
			builder.PrependUOffsetT(te)
		}
		respHdrVec = builder.EndVector(size)
	}

	argsLen := len(opt.Args)
	var argsVec flatbuffers.UOffsetT
	if argsLen > 0 {
		Args := []flatbuffers.UOffsetT{}
		for _, v := range opt.Args {
			name := builder.CreateString(v.Name)
			value := builder.CreateString(v.Value)
			A6.TextEntryStart(builder)
			A6.TextEntryAddName(builder, name)
			A6.TextEntryAddValue(builder, value)
			te := A6.TextEntryEnd(builder)
			Args = append(Args, te)
		}
		size := len(Args)
		hrc.RewriteStartArgsVector(builder, size)
		for i := size - 1; i >= 0; i-- {
			te := Args[i]
			builder.PrependUOffsetT(te)
		}
		argsVec = builder.EndVector(size)
	}

	hrc.ReqStart(builder)
	hrc.ReqAddId(builder, reqId)
	hrc.ReqAddConfToken(builder, confToken)
	if ip > 0 {
		hrc.ReqAddSrcIp(builder, ip)
	}
	if opt.Method != 0 {
		hrc.ReqAddMethod(builder, opt.Method)
	}
	if path > 0 {
		hrc.ReqAddPath(builder, path)
	}
	if hdrVec > 0 {
		hrc.ReqAddHeaders(builder, hdrVec)
	}
	if respHdrVec > 0 {
		hrc.RewriteAddRespHeaders(builder, respHdrVec)
	}
	if argsVec > 0 {
		hrc.ReqAddArgs(builder, argsVec)
	}
	r := hrc.ReqEnd(builder)
	builder.Finish(r)
	return builder.FinishedBytes()
}
