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

package agent

import (
	"context"
	"net"

	"github.com/golang/protobuf/proto"
	"github.com/tutumagi/pitaya/cluster"
	"github.com/tutumagi/pitaya/conn/message"
	"github.com/tutumagi/pitaya/constants"
	"github.com/tutumagi/pitaya/logger"
	"github.com/tutumagi/pitaya/protos"
	"github.com/tutumagi/pitaya/route"
	"github.com/tutumagi/pitaya/serialize"
	"github.com/tutumagi/pitaya/util"
)

// Remote corresponding to another server
type Remote struct {
	frontSessID int64
	uid         string
	chDie       chan struct{} // wait for close
	// messageEncoder message.Encoder
	// encoder    codec.PacketEncoder // binary encoder
	frontendID string // the frontend that sent the request

	rpcClient        cluster.RPCClient        // rpc client
	serializer       serialize.Serializer     // message serializer
	serviceDiscovery cluster.ServiceDiscovery // service discovery
}

// NewRemote create new Remote instance
func NewRemote(
	frontSessID int64,
	frontServerID string,
	uid string,
	rpcClient cluster.RPCClient,
	// encoder codec.PacketEncoder,
	serializer serialize.Serializer,
	serviceDiscovery cluster.ServiceDiscovery,
	// frontendID string,
	// messageEncoder message.Encoder,
) (*Remote, error) {
	a := &Remote{
		frontSessID: frontSessID,
		frontendID:  frontServerID,
		uid:         uid,
		chDie:       make(chan struct{}),
		serializer:  serializer,
		// encoder:          encoder,
		rpcClient:        rpcClient,
		serviceDiscovery: serviceDiscovery,

		// messageEncoder: messageEncoder,
	}

	return a, nil
}

// Kick kicks the user
func (a *Remote) Kick(ctx context.Context) error {
	if a.uid == "" {
		return constants.ErrNoUIDBind
	}
	b, err := proto.Marshal(&protos.KickMsg{
		UserId: a.uid,
	})
	if err != nil {
		return err
	}
	_, err = a.SendRequest(ctx, "", "", a.frontendID, constants.KickRoute, b)
	return err
}

// Push pushes the message to the user
func (a *Remote) Push(route string, v interface{}) error {
	if a.uid == "" {
		return constants.ErrNoUIDBind
	}
	switch d := v.(type) {
	case []byte:
		logger.Log.Debugf("Type=Push, ID=%d, UID=%d, Route=%s, Data=%dbytes",
			a.frontSessID, a.uid, route, len(d))
	default:
		logger.Log.Debugf("Type=Push, ID=%d, UID=%d, Route=%s, Data=%+v",
			a.frontSessID, a.uid, route, v)
	}

	sv, err := a.serviceDiscovery.GetServer(a.frontendID)
	if err != nil {
		return err
	}
	return a.sendPush(
		pendingMessage{typ: message.Push, route: route, payload: v},
		a.uid, sv,
	)
}

// Close closes the remote
func (a *Remote) Close() error { return nil }

// RemoteAddr returns the remote address of the user
func (a *Remote) RemoteAddr() net.Addr { return nil }

// func (a *Remote) serialize(m pendingMessage) ([]byte, error) {
// 	payload, err := util.SerializeOrRaw(a.serializer, m.payload)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// construct message and encode
// 	msg := &message.Message{
// 		Type:  m.typ,
// 		Data:  payload,
// 		Route: m.route,
// 		ID:    m.mid,
// 		Err:   m.err,
// 	}

// 	em, err := a.messageEncoder.Encode(msg)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// packet encode
// 	p, err := a.encoder.Encode(packet.Data, em)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return p, err
// }

// func (a *Remote) send(m pendingMessage, to string) (err error) {
// 	p, err := a.serialize(m)
// 	if err != nil {
// 		return err
// 	}
// 	res := &protos.Response{
// 		Data: p,
// 	}
// 	bt, err := proto.Marshal(res)
// 	if err != nil {
// 		return err
// 	}
// 	return a.rpcClient.Send(to, bt)
// }

func (a *Remote) sendPush(m pendingMessage, userID string, sv *cluster.Server) (err error) {
	payload, err := util.SerializeOrRaw(a.serializer, m.payload)
	if err != nil {
		return err
	}
	push := &protos.Push{
		Route: m.route,
		Uid:   a.uid,
		Data:  payload,
	}
	return a.rpcClient.SendPush(userID, sv, push)
}

// SendRequest sends a request to a server
func (a *Remote) SendRequest(ctx context.Context, entityID, entityType, serverID, reqRoute string, v interface{}) (*protos.Response, error) {
	r, err := route.Decode(reqRoute)
	if err != nil {
		return nil, err
	}
	payload, err := util.SerializeOrRaw(a.serializer, v)
	if err != nil {
		return nil, err
	}
	msg := &message.Message{
		Route:      reqRoute,
		Data:       payload,
		EntityID:   entityID,
		EntityType: entityType,
	}
	server, err := a.serviceDiscovery.GetServer(serverID)
	if err != nil {
		return nil, err
	}
	return a.rpcClient.Call(ctx, protos.RPCType_User, r, nil, msg, server)
}

// SendRequest sends a request to a server
func (a *Remote) Bind(ctx context.Context, uid string) error {
	a.uid = uid
	route := constants.SessionBindRoute
	sessionData := &protos.Session{
		Id:  a.frontSessID,
		Uid: a.uid,
	}

	b, err := proto.Marshal(sessionData)
	if err != nil {
		a.uid = ""
		return err
	}
	// TODO 这里没有entityID和entityType
	res, err := a.SendRequest(ctx, "", "", a.frontendID, route, b)
	if err != nil {
		a.uid = ""
		return err
	}
	logger.Log.Debugf("bind uid(%s) response: %+v", uid, res)
	return nil
}
