//
// Copyright (c) TFG Co. All Rights Reserved.
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package metapart

import (
	"context"
	"reflect"

	"github.com/golang/protobuf/proto"

	"github.com/tutumagi/pitaya/agent"
	"github.com/tutumagi/pitaya/cluster"
	"github.com/tutumagi/pitaya/conn/codec"
	"github.com/tutumagi/pitaya/conn/message"
	"github.com/tutumagi/pitaya/constants"
	e "github.com/tutumagi/pitaya/errors"
	"github.com/tutumagi/pitaya/logger"
	"github.com/tutumagi/pitaya/protos"
	"github.com/tutumagi/pitaya/route"
	"github.com/tutumagi/pitaya/router"
	"github.com/tutumagi/pitaya/serialize"
	"github.com/tutumagi/pitaya/util"
)

// EntityMsgProcessor struct
type EntityMsgProcessor struct {
	serviceDiscovery cluster.ServiceDiscovery
	serializer       serialize.Serializer
	encoder          codec.PacketEncoder
	rpcClient        cluster.RPCClient

	router         *router.Router
	messageEncoder message.Encoder
}

// NewEntityProcessor creates and return a new RemoteService
func NewEntityProcessor(
	serviceDiscovery cluster.ServiceDiscovery,
	serializer serialize.Serializer,
	encoder codec.PacketEncoder,
	rpcClient cluster.RPCClient,
	router *router.Router,
	messageEncoder message.Encoder,
) *EntityMsgProcessor {
	return &EntityMsgProcessor{
		serviceDiscovery: serviceDiscovery,
		serializer:       serializer,
		encoder:          encoder,
		rpcClient:        rpcClient,
		router:           router,
		messageEncoder:   messageEncoder,
	}
}

// func (r *EntityMsgProcessor) processMessage(ctx context.Context, req *protos.Request, entity interface{}, routers *Routers) *protos.Response {
func (r *EntityMsgProcessor) ProcessMessage(ctx context.Context, req *protos.Request, entity interface{}, routers *Routers) *protos.Response {
	rt, err := route.Decode(req.GetMsg().GetRoute())
	if err != nil {
		response := &protos.Response{
			Error: &protos.Error{
				Code: e.ErrBadRequestCode,
				Msg:  "cannot decode route",
				Metadata: map[string]string{
					"route": req.GetMsg().GetRoute(),
				},
			},
		}
		return response
	}

	switch {
	case req.Type == protos.RPCType_Sys:
		return r.handleRPCSys(ctx, req, rt, entity, routers)
	case req.Type == protos.RPCType_User:
		return r.handleRPCUser(ctx, req, rt, entity, routers)
	default:
		return &protos.Response{
			Error: &protos.Error{
				Code: e.ErrBadRequestCode,
				Msg:  "invalid rpc type",
				Metadata: map[string]string{
					"route": req.GetMsg().GetRoute(),
				},
			},
		}
	}
}

func (r *EntityMsgProcessor) handleRPCUser(ctx context.Context, req *protos.Request, rt *route.Route, entity interface{}, routers *Routers) *protos.Response {
	response := &protos.Response{}

	remote, err := routers.getRemote(rt)
	// remote, ok := remotes[rt.Short()]
	if err != nil {
		logger.Log.Warnf("pitaya/remote: %s not found", rt.Short())
		response := &protos.Response{
			Error: &protos.Error{
				Code: e.ErrNotFoundCode,
				Msg:  "route not found",
				Metadata: map[string]string{
					"route": rt.Short(),
				},
			},
		}
		return response
	}
	params := []reflect.Value{remote.Receiver, reflect.ValueOf(ctx), reflect.ValueOf(entity)}
	if remote.HasArgs {
		arg, err := unmarshalRemoteArg(remote, req.GetMsg().GetData())
		if err != nil {
			response := &protos.Response{
				Error: &protos.Error{
					Code: e.ErrBadRequestCode,
					Msg:  err.Error(),
				},
			}
			return response
		}
		params = append(params, reflect.ValueOf(arg))
	}

	ret, err := util.Pcall(remote.Method, params)
	if err != nil {
		response := &protos.Response{
			Error: &protos.Error{
				Code: e.ErrUnknownCode,
				Msg:  err.Error(),
			},
		}
		if val, ok := err.(*e.Error); ok {
			response.Error.Code = val.Code
			if val.Metadata != nil {
				response.Error.Metadata = val.Metadata
			}
		}
		return response
	}

	var b []byte
	if ret != nil {
		pb, ok := ret.(proto.Message)
		if !ok {
			response := &protos.Response{
				Error: &protos.Error{
					Code: e.ErrUnknownCode,
					Msg:  constants.ErrWrongValueType.Error(),
				},
			}
			return response
		}
		if b, err = proto.Marshal(pb); err != nil {
			response := &protos.Response{
				Error: &protos.Error{
					Code: e.ErrUnknownCode,
					Msg:  err.Error(),
				},
			}
			return response
		}
	}

	response.Data = b
	return response
}

func (r *EntityMsgProcessor) handleRPCSys(ctx context.Context, req *protos.Request, rt *route.Route, entity interface{}, routers *Routers) *protos.Response {
	reply := req.GetMsg().GetReply()
	response := &protos.Response{}
	// (warning) a new agent is created for every new request
	a, err := agent.NewRemote(
		req.GetSession(),
		reply,
		r.rpcClient,
		r.encoder,
		r.serializer,
		r.serviceDiscovery,
		req.FrontendID,
		r.messageEncoder,
	)
	if err != nil {
		logger.Log.Warn("pitaya/handler: cannot instantiate remote agent")
		response := &protos.Response{
			Error: &protos.Error{
				Code: e.ErrInternalCode,
				Msg:  err.Error(),
			},
		}
		return response
	}

	ret, err := processHandlerMessage(
		ctx,
		rt,
		r.serializer,
		a.Session,
		entity,
		routers,
		req.GetMsg().GetData(),
		req.GetMsg().GetType(),
		true,
	)
	if err != nil {
		logger.Log.Warnf(err.Error())
		response = &protos.Response{
			Error: &protos.Error{
				Code: e.ErrUnknownCode,
				Msg:  err.Error(),
			},
		}
		if val, ok := err.(*e.Error); ok {
			response.Error.Code = val.Code
			if val.Metadata != nil {
				response.Error.Metadata = val.Metadata
			}
		}
	} else {
		response = &protos.Response{Data: ret}
	}
	return response
}
