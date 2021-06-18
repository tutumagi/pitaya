package app

import (
	"context"
	"fmt"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/golang/protobuf/proto"
	"github.com/tutumagi/pitaya/cluster"
	"github.com/tutumagi/pitaya/conn/codec"
	"github.com/tutumagi/pitaya/conn/message"
	"github.com/tutumagi/pitaya/engine/common"
	e "github.com/tutumagi/pitaya/errors"
	"github.com/tutumagi/pitaya/logger"
	"github.com/tutumagi/pitaya/metrics"
	"github.com/tutumagi/pitaya/protos"
	"github.com/tutumagi/pitaya/route"
	"github.com/tutumagi/pitaya/router"
	"github.com/tutumagi/pitaya/serialize"
	"github.com/tutumagi/pitaya/timer"
)

type EntityManager interface {
	// 通过实体 ID 获取 typeName
	GetTypName(id string) (string, error)
	// 获取实际的实体对象，可能是 cellpart.Entity.Val() ，也可能是 basepart.Entity.Val()，所以用 interface{}
	GetEntityVal(id string, typName string) interface{}
	// 获取实体绑定的pid
	GetEntityPid(id string, typName string) *actor.PID
}

type AppMsgProcessor struct {
	appDieChan chan bool // die channel app

	messageEncoder message.Encoder
	serializer     serialize.Serializer // message serializer
	encoder        codec.PacketEncoder
	decoder        codec.PacketDecoder

	// 当前server信息
	server *cluster.Server

	metricsReporters []metrics.Reporter

	heartbeatTimeout   time.Duration
	messagesBufferSize int

	remote *common.RemoteService

	entityManager EntityManager
	actorSystem   *actor.ActorSystem
}

func NewAppProcessor(
	dieChan chan bool,
	serializer serialize.Serializer,
	server *cluster.Server,
	messageEncoder message.Encoder,
	metricsReporters []metrics.Reporter,

	rpcClient cluster.RPCClient,
	rpcServer cluster.RPCServer,
	sd cluster.ServiceDiscovery,
	router *router.Router,

	system *actor.ActorSystem,
) *AppMsgProcessor {
	p := &AppMsgProcessor{
		serializer:       serializer,
		appDieChan:       dieChan,
		server:           server,
		metricsReporters: metricsReporters,
		messageEncoder:   messageEncoder,

		actorSystem: system,
	}
	remote := common.NewRemoteService(dieChan, serializer, server, metricsReporters, rpcClient, rpcServer, sd, router, system, p)
	p.remote = remote

	return p
}

func (p *AppMsgProcessor) Start() {
	// for i := 0; i < 10; i++ {
	// 	go p.Dispatch()
	// }

	// app.config.GetInt("pitaya.concurrency.handler.dispatch")
	const numberDispatch = 10
	for i := 0; i < numberDispatch; i++ {
		go p.Dispatch(i)
	}
}

func (p *AppMsgProcessor) Dispatch(thread int) {
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

func (r *AppMsgProcessor) Process(ctx context.Context, req *protos.Request) *protos.Response {
	entityID := req.Msg.Eid
	entityType := req.Msg.Typ
	entity := r.entityManager.GetEntityVal(entityID, entityType)
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
		r.entityManager.GetEntityPid(entityID, entityType),
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

func (p *AppMsgProcessor) CallService(ctx context.Context, serviceName string, routeStr string, reply proto.Message, arg proto.Message) error {
	entityID := common.ServiceID(serviceName)
	entityType := common.ServiceTypeName(serviceName)

	return p.call(ctx, entityID, entityType, routeStr, reply, arg)
}

func (p *AppMsgProcessor) SendService(ctx context.Context, serviceName string, routeStr string, arg proto.Message) error {
	entityID := common.ServiceID(serviceName)
	entityType := common.ServiceTypeName(serviceName)

	return p.call(ctx, entityID, entityType, routeStr, nil, arg)
}

func (p *AppMsgProcessor) CallEntity(
	ctx context.Context,
	entityID,
	entityType string,
	routeStr string,
	reply proto.Message,
	arg proto.Message,
) error {
	return p.call(ctx, entityID, entityType, routeStr, reply, arg)
}

func (p *AppMsgProcessor) SendEntity(
	ctx context.Context,
	entityID,
	entityType string,
	routeStr string,
	arg proto.Message,
) error {
	return p.call(ctx, entityID, entityType, routeStr, nil, arg)
}

func (p *AppMsgProcessor) call(
	ctx context.Context,
	entityID,
	entityType string,
	routeStr string,
	reply proto.Message,
	arg proto.Message,
) error {
	var ret interface{}
	var err error

	// 本地没有这个实体
	entity := p.entityManager.GetEntityVal(entityID, entityType)
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
		p.entityManager.GetEntityPid(entityID, entityType),
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

// ret, err := processHandlerMessage(
// 	ctx,
// 	route,
// 	p.serializer,
// 	a.Session,
// 	msg.EntityID,
// 	msg.EntityType,
// 	p,
// 	msg.Data,
// 	msg.Type,
// 	false,
// )
// if msg.Type != message.Notify {
// 	if err != nil {
// 		logger.Log.Errorf("Failed to process handler(route:%s) message: %s", route.Short(), err.Error())
// 		a.AnswerWithError(ctx, mid, err)
// 	} else {
// 		err := a.Session.ResponseMID(ctx, mid, ret)
// 		if err != nil {
// 			tracing.FinishSpan(ctx, err)
// 			metrics.ReportTimingFromCtx(ctx, p.metricsReporters, handlerType, err)
// 		}
// 	}
// } else {
// 	metrics.ReportTimingFromCtx(ctx, p.metricsReporters, handlerType, nil)
// 	tracing.FinishSpan(ctx, err)
// }
