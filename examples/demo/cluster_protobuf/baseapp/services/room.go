package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tutumagi/pitaya/component"
	"github.com/tutumagi/pitaya/config"
	"github.com/tutumagi/pitaya/examples/demo/cluster_protobuf/protos"
	"github.com/tutumagi/pitaya/groups"
	"github.com/tutumagi/pitaya/timer"

	"github.com/tutumagi/pitaya/engine/components/app"
)

type (
	// Room represents a component that contains a bundle of room related handler
	// like Join/Message
	Room struct {
		component.Base
		timer *timer.Timer
		Stats *Stats
	}

	// Stats exports the room status
	Stats struct {
		outboundBytes int
		inboundBytes  int
	}
)

// Outbound gets the outbound status
func (Stats *Stats) Outbound(ctx context.Context, in []byte) ([]byte, error) {
	Stats.outboundBytes += len(in)
	return in, nil
}

// Inbound gets the inbound status
func (Stats *Stats) Inbound(ctx context.Context, in []byte) ([]byte, error) {
	Stats.inboundBytes += len(in)
	return in, nil
}

// NewRoom returns a new room
func NewRoom() *Room {
	return &Room{
		Stats: &Stats{},
	}
}

// Init runs on service initialization
func (r *Room) Init() {
	gsi := groups.NewMemoryGroupService(config.NewConfig())
	app.InitGroups(gsi)
	app.GroupCreate(context.Background(), "room")
}

// AfterInit component lifetime callback
func (r *Room) AfterInit() {
	r.timer = app.NewTimer(time.Minute, func() {
		count, err := app.GroupCountMembers(context.Background(), "room")
		println("UserCount: Time=>", time.Now().String(), "Count=>", count, "Error=>", err)
		println("OutboundBytes", r.Stats.outboundBytes)
		println("InboundBytes", r.Stats.outboundBytes)
	})
}

func reply(code int32, msg string) *protos.Response {
	return &protos.Response{
		Code: code,
		Msg:  msg,
	}
}

// Entry is the entrypoint
func (r *Room) Entry(ctx context.Context) (*protos.Response, error) {
	fakeUID := uuid.New().String() // just use s.ID as uid !!!
	s := app.GetSessionFromCtx(ctx)
	err := s.Bind(ctx, fakeUID) // binding session uid
	if err != nil {
		return nil, app.Error(err, "ENT-000")
	}
	return reply(200, "ok"), nil
}

// Join room
func (r *Room) Join(ctx context.Context) (*protos.Response, error) {
	s := app.GetSessionFromCtx(ctx)
	members, err := app.GroupMembers(ctx, "room")
	if err != nil {
		return nil, err
	}
	s.Push("onMembers", &protos.AllMembers{Members: members})
	app.GroupBroadcast(ctx, "connector", "room", "onNewUser", &protos.NewUser{Content: fmt.Sprintf("New user: %d", s.ID())})
	app.GroupAddMember(ctx, "room", s.UID())
	s.OnClose(func() {
		app.GroupRemoveMember(ctx, "room", s.UID())
	})
	return &protos.Response{Msg: "success"}, nil
}

// Message sync last message to all members
func (r *Room) Message(ctx context.Context, msg *protos.UserMessage) {
	err := app.GroupBroadcast(ctx, "connector", "room", "onMessage", msg)
	if err != nil {
		fmt.Println("error broadcasting message", err)
	}
}

// SendRPC sends rpc
func (r *Room) SendRPC(ctx context.Context, msg []byte) (*protos.Response, error) {
	ret := protos.Response{}
	err := app.RPC(ctx, "connector.connectorremote.remotefunc", &ret, &protos.RPCMsg{})
	if err != nil {
		return nil, app.Error(err, "RPC-000")
	}
	return reply(200, ret.Msg), nil
}
