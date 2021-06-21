package handler

import (
	"context"
	"fmt"

	"github.com/tutumagi/pitaya/constants"
	"github.com/tutumagi/pitaya/engine/bc/metapart"
	"github.com/tutumagi/pitaya/engine/components/app"
	"github.com/tutumagi/pitaya/examples/demo/cluster_protobuf/baseapp/entity"
	"github.com/tutumagi/pitaya/examples/demo/cluster_protobuf/protos"
	"github.com/tutumagi/pitaya/logger"
)

type AccountHandler struct {
}

// Join room
func (r *AccountHandler) Join(ctx context.Context, msg []byte) (*protos.ResponseV2, error) {
	player := ctx.Value(constants.EntityCtxKey).(*entity.Account)

	logger.Infof("account join :%s", player.ID)
	// rsp := &JoinResponse{}
	// err := player.CallService(ctx, "room", "room.join", rsp, msg)
	// if err != nil {
	// 	return nil, err
	// }

	// return rsp, nil
	return &protos.ResponseV2{Msg: "ack"}, nil
}

// Message sync last message to all members
func (r *AccountHandler) Message(ctx context.Context, msg *protos.UserMessage) {
	err := app.GroupBroadcast(ctx, metapart.GateAppSvr, "room", "onMessage", msg)
	if err != nil {
		fmt.Println("error broadcasting message", err)
	}
}
