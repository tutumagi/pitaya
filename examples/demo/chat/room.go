package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/tutumagi/pitaya/engine/bc/basepart"
	"github.com/tutumagi/pitaya/engine/components/app"
	"github.com/tutumagi/pitaya/logger"
)

// NewRoom returns a Handler Base implementation
func NewRoom() *Room {
	return &Room{}
}

type RoomService struct {
	basepart.Entity
}

// AfterInit component lifetime callback
func (r *Room) AfterInit() {
	r.timer = app.NewTimer(time.Minute, func() {
		count, err := app.GroupCountMembers(context.Background(), "room")
		logger.Log.Debugf("UserCount: Time=> %s, Count=> %d, Error=> %q", time.Now().String(), count, err)
	})
}

// Join room
func (r *Room) Join(ctx context.Context, entity interface{}, msg []byte) (*JoinResponse, error) {
	s := app.GetSessionFromCtx(ctx)
	fakeUID := s.ID()                              // just use s.ID as uid !!!
	err := s.Bind(ctx, strconv.Itoa(int(fakeUID))) // binding session uid

	if err != nil {
		return nil, app.Error(err, "RH-000", map[string]string{"failed": "bind"})
	}

	uids, err := app.GroupMembers(ctx, "room")
	if err != nil {
		return nil, err
	}
	s.Push("onMembers", &AllMembers{Members: uids})
	// notify others
	app.GroupBroadcast(ctx, "chat", "room", "onNewUser", &NewUser{Content: fmt.Sprintf("New user: %s", s.UID())})
	// new user join group
	app.GroupAddMember(ctx, "room", s.UID()) // add session to group

	// on session close, remove it from group
	s.OnClose(func() {
		app.GroupRemoveMember(ctx, "room", s.UID())
	})

	return &JoinResponse{Result: "success"}, nil
}

// Message sync last message to all members
func (r *Room) Message(ctx context.Context, entity interface{}, msg *UserMessage) {
	err := app.GroupBroadcast(ctx, "chat", "room", "onMessage", msg)
	if err != nil {
		fmt.Println("error broadcasting message", err)
	}
}
