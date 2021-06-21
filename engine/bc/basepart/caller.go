package basepart

import (
	"context"
	"fmt"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/golang/protobuf/proto"
	"github.com/tutumagi/pitaya/engine/common"
	e "github.com/tutumagi/pitaya/errors"
	"github.com/tutumagi/pitaya/logger"
	"github.com/tutumagi/pitaya/metrics"
	"github.com/tutumagi/pitaya/protos"
	"github.com/tutumagi/pitaya/route"
	"github.com/tutumagi/pitaya/serialize"
	"github.com/tutumagi/pitaya/timer"
)

type Caller struct {
	appDieChan chan bool // die channel app

	serializer serialize.Serializer

	// // 当前server信息
	// server *cluster.Server

	metricsReporters []metrics.Reporter

	remote common.EntityRemoteCaller

	actorSystem *actor.ActorSystem
}

func NewAppProcessor(
	dieChan chan bool,
	serializer serialize.Serializer,
	system *actor.ActorSystem,
	remote common.EntityRemoteCaller,
) *Caller {
	p := &Caller{
		appDieChan:  dieChan,
		serializer:  serializer,
		remote:      remote,
		actorSystem: system,
	}

	return p
}

func (p *Caller) Start() {
	// for i := 0; i < 10; i++ {
	// 	go p.Dispatch()
	// }

	// app.config.GetInt("pitaya.concurrency.handler.dispatch")
	const numberDispatch = 10
	for i := 0; i < numberDispatch; i++ {
		go p.Dispatch(i)
	}
}

func (p *Caller) Dispatch(thread int) {
	// TODO: This timer is being stopped multiple times, it probably doesn't need to be stopped here
	// defer timer.GlobalTicker.Stop()
	defer func() {
		logger.Log.Warnf("Go HandlerService::Dispatch(%d) exit", thread)
		timer.GlobalTicker.Stop()
		if err := recover(); err != nil {
			logger.Log.Warnf("Go HandlerService::Dispatch(%d) exit by err = %v", thread, err)
		}
	}()

	for {
		// Calls to remote servers block calls to local server
		select {
		// 收到 rpc call/post 后，处理消息
		// case rpcReq := <-p.remoteService.rpcServer.GetUnhandledRequestsChannel():
		// 	// logger.Log.Infof("pitaya.handler Dispatch -> rpc.ProcessSingleMessage <0> for ", zap.Any("rpcReq", rpcReq))
		// 	// logger.Log.Debugf("pitaya.handler Dispatch -> rpc.ProcessSingleMessage <0> for route=%s", rpcReq.Msg.Route)
		// 	p.remoteService.rpcServer.ProcessSingleMessage(rpcReq)
		// logger.Log.Infof("pitaya.handler Dispatch -> rpc.ProcessSingleMessage <1> for ", zap.Any("rpcReq", rpcReq))
		// logger.Log.Debugf("pitaya.handler Dispatch -> rpc.ProcessSingleMessage <1> for route=%s", rpcReq.Msg.Route)
		case <-timer.GlobalTicker.C: // execute cron task
			timer.Cron()

		case t := <-timer.Manager.ChCreatedTimer: // new Timers
			timer.AddTimer(t)

		case id := <-timer.Manager.ChClosingTimer: // closing Timers
			timer.RemoveTimer(id)
		}
	}
}

func (r *Caller) Process(ctx context.Context, req *protos.Request) *protos.Response {
	entityID := req.Msg.Eid
	entityType := req.Msg.Typ
	entity := baseEntManager.get(entityID, entityType)
	if entity == nil {
		logger.Log.Warnf("pitaya/remote process message to entity: entity(id:%s type:%s) not found", entityID, entityType)

		return &protos.Response{
			Error: &protos.Error{
				Code: e.ErrUnknownCode,
				Msg:  fmt.Sprintf("entity(id:%s type:%s) not found", entityID, entityType),
			},
		}
	}

	rsp, err := r.actorSystem.Root.RequestFuture(
		baseEntManager.getPid(entityID, entityType),
		&common.LocalMessageWrapper{
			Ctx: ctx,
			Req: req,
		},
		//TODO 这里写的2秒
		2*time.Second,
	).Result()

	if err != nil {
		return &protos.Response{
			Error: &protos.Error{
				Code: e.ErrUnknownCode,
				Msg:  fmt.Sprintf("actor.proto requestFuture error:%s", err),
			},
		}
	} else {
		if rspp, ok := rsp.(*protos.Response); ok {
			return rspp
		} else {
			return &protos.Response{
				Error: &protos.Error{
					Code: e.ErrUnknownCode,
					Msg:  "rsp type is not *protos.Response",
				},
			}
		}
	}
}

func (p *Caller) Call(
	ctx context.Context,
	serverID string,
	entityID,
	entityType string,
	routeStr string,
	reply proto.Message,
	arg proto.Message,
) error {
	var ret interface{}
	var err error

	// 本地没有这个实体
	entity := baseEntManager.get(entityID, entityType)
	if entity == nil {
		logger.Log.Warnf("pitaya/local process message to entity: entity(id:%s type:%s) not found", entityID, entityType)
		droute, err := route.Decode(routeStr)
		if err != nil {
			return err
		}
		if reply != nil {
			return p.remote.RPC(ctx, entityID, entityType, "", droute, reply, arg)
		} else {
			return p.remote.Send(ctx, entityID, entityType, "", droute, arg)
		}
		// return &protos.Response{
		// 	Error: &protos.Error{
		// 		Code: e.ErrUnknownCode,
		// 		Msg:  fmt.Sprintf("entity(id:%s type:%s) not found", entityID, entityType),
		// 	},
		// }
		// err = e.NewError(fmt.Errorf("entity(id:%s type:%s) not found", entityID, entityType), e.ErrUnknownCode)
		// return err
	}

	// TODO 本地 call 这里marshal了一次，到实际call方法时，又unmarshal一次，重复了
	argBytes, _ := p.serializer.Marshal(arg)

	ret, err = p.actorSystem.Root.RequestFuture(
		baseEntManager.getPid(entityID, entityType),
		&common.LocalMessageWrapper{
			Ctx: ctx,
			Req: &protos.Request{
				Type: protos.RPCType_User,
				Msg: &protos.MsgV2{
					// Id:    uint64(mid),
					Route: routeStr,
					Data:  argBytes,
					// Type:  protos.MsgType(msg.Type),
					Eid: entityID,
					Typ: entityType,
					// Reply: ,
				},
			},
		},
		//TODO 这里写的2秒
		2*time.Second,
	).Result()

	if err != nil {
		return err
	}
	if reply != nil {
		err := p.serializer.Unmarshal(ret.([]byte), reply)
		if err != nil {
			return err
		}
	}

	return nil

}

// func (p *Caller) callFromLocal(
// 	ctx context.Context,
// 	entityID,
// 	entityType string,
// 	routeStr string,
// 	reply proto.Message,
// 	arg proto.Message,
// ) error {

// }
