package main

import (
	"context"
	"fmt"

	"github.com/tutumagi/pitaya/engine/bc/basepart"
	"github.com/tutumagi/pitaya/engine/bc/metapart"
	"github.com/tutumagi/pitaya/engine/components/app"
)

type Player struct {
	basepart.Entity
}

type PlayerHandler struct{}

// Join room
func (r *PlayerHandler) Join(ctx context.Context, entity interface{}, msg []byte) (*JoinResponse, error) {
	player := entity.(*Player)

	player.ID = metapart.NewUUID()

	// rsp := &JoinResponse{}
	// err := player.CallService(ctx, "room", "room.join", rsp, msg)
	// if err != nil {
	// 	return nil, err
	// }

	// return rsp, nil
	return &JoinResponse{Result: "ack"}, nil
}

// Message sync last message to all members
func (r *PlayerHandler) Message(ctx context.Context, msg *UserMessage) {
	err := app.GroupBroadcast(ctx, "chat", "room", "onMessage", msg)
	if err != nil {
		fmt.Println("error broadcasting message", err)
	}
}
