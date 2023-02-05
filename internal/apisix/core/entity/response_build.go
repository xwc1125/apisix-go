// Package entity
//
// @author: xwc1125
package entity

import (
	"github.com/api7/ext-plugin-proto/go/A6"
	hrc "github.com/api7/ext-plugin-proto/go/A6/HTTPRespCall"
	flatbuffers "github.com/google/flatbuffers/go"
)

type RespReqOpt struct {
	Id         uint32
	StatusCode int
	Headers    []Pair
	Token      uint32
}

func BuildRespReqOpt(opt RespReqOpt) []byte {
	builder := flatbuffers.NewBuilder(1024)

	hdrLen := len(opt.Headers)
	var hdrVec flatbuffers.UOffsetT
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
		hrc.ReqStartHeadersVector(builder, size)
		for i := size - 1; i >= 0; i-- {
			te := hdrs[i]
			builder.PrependUOffsetT(te)
		}
		hdrVec = builder.EndVector(size)
	}

	hrc.ReqStart(builder)
	hrc.ReqAddId(builder, uint32(opt.Id))
	hrc.ReqAddConfToken(builder, uint32(opt.Token))

	if opt.StatusCode != 0 {
		hrc.ReqAddStatus(builder, uint16(opt.StatusCode))
	}
	if hdrVec > 0 {
		hrc.ReqAddHeaders(builder, hdrVec)
	}
	r := hrc.ReqEnd(builder)
	builder.Finish(r)
	return builder.FinishedBytes()
}
