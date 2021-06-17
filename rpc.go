// Copyright (c) nano Author and TFG Co. All Rights Reserved.
//
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

package pitaya

import (
	"context"
	"reflect"

	"github.com/golang/protobuf/proto"
	"github.com/tutumagi/pitaya/constants"
	"github.com/tutumagi/pitaya/route"
	"github.com/tutumagi/pitaya/worker"
)

// func CallEntity(ctx context.Context, id string, routerStr string, reply proto.Message, arg proto.Message) error {

// }

// RPC calls a method in a different server
func RPC(ctx context.Context, entityID, entityType string, routeStr string, reply proto.Message, arg proto.Message) error {
	return doSendRPC(ctx, entityID, entityType, "", routeStr, reply, arg)
}

// RPCTo send a rpc to a specific server
func RPCTo(ctx context.Context, entityID, entityType string, serverID, routeStr string, reply proto.Message, arg proto.Message) error {
	return doSendRPC(ctx, entityID, entityType, serverID, routeStr, reply, arg)
}

// Send calls a method in a different server
func Send(ctx context.Context, entityID, entityType string, routeStr string, arg proto.Message) error {
	return doSendRPC(ctx, entityID, entityType, "", routeStr, nil, arg)
}

// SendTo send a rpc to a specific server
func SendTo(ctx context.Context, entityID, entityType string, serverID, routeStr string, arg proto.Message) error {
	return doSendRPC(ctx, entityID, entityType, serverID, routeStr, nil, arg)
}

// ReliableRPC enqueues RPC to worker so it's executed asynchronously
// Default enqueue options are used
func ReliableRPC(
	routeStr string,
	metadata map[string]interface{},
	reply, arg proto.Message,
) (jid string, err error) {
	return app.worker.EnqueueRPC(routeStr, metadata, reply, arg)
}

// ReliableRPCWithOptions enqueues RPC to worker
// Receive worker options for this specific RPC
func ReliableRPCWithOptions(
	routeStr string,
	metadata map[string]interface{},
	reply, arg proto.Message,
	opts *worker.EnqueueOpts,
) (jid string, err error) {
	return app.worker.EnqueueRPCWithOptions(routeStr, metadata, reply, arg, opts)
}

func doSendRPC(ctx context.Context, entityID, entityType string, serverID, routeStr string, reply proto.Message, arg proto.Message) error {
	if app.rpcServer == nil {
		return constants.ErrRPCServerNotInitialized
	}

	if reply != nil {
		if reflect.TypeOf(reply).Kind() != reflect.Ptr {
			return constants.ErrReplyShouldBePtr
		}
	}

	rt, err := route.Decode(routeStr)
	if err != nil {
		return err
	}

	// 如果既没有serverID 又没有 serverType 则返回 by 涂飞
	if serverID == "" && rt.SvType == "" {
		return constants.ErrNoServerTypeChosenForRPC
	}

	if (rt.SvType == app.server.Type && serverID == "") || serverID == app.server.ID {
		// 如果发现是 rpc 的服务是 本地 则直接 call 本地的方法 by 涂飞
		// return constants.ErrNonsenseRPC

		return handlerService.CallEntityFromLocal(ctx, entityID, entityType, routeStr, reply, arg)
	}

	if reply == nil {
		// 如果没有reply 则使用 rpc send
		return handlerService.Send(ctx, entityID, entityType, serverID, rt, reply, arg)
	} else {
		// 如果有reply 则使用 rpc call
		return handlerService.RPC(ctx, entityID, entityType, serverID, rt, reply, arg)
	}
}
