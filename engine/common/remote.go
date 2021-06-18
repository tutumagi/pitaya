package common

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/golang/protobuf/proto"
	"github.com/tutumagi/pitaya/agent"
	"github.com/tutumagi/pitaya/cluster"
	"github.com/tutumagi/pitaya/conn/message"
	"github.com/tutumagi/pitaya/constants"
	"github.com/tutumagi/pitaya/engine/bc/metapart"
	e "github.com/tutumagi/pitaya/errors"
	"github.com/tutumagi/pitaya/logger"
	"github.com/tutumagi/pitaya/metrics"
	"github.com/tutumagi/pitaya/protos"
	"github.com/tutumagi/pitaya/route"
	"github.com/tutumagi/pitaya/router"
	"github.com/tutumagi/pitaya/serialize"
	"github.com/tutumagi/pitaya/session"
	"github.com/tutumagi/pitaya/timer"
	"github.com/tutumagi/pitaya/tracing"
	"github.com/tutumagi/pitaya/util"
)

// 用来处理rpc send/call 和 rpc process
type RemoteService struct {
	appDieChan chan bool // die channel app

	serializer serialize.Serializer // message serializer

	// 当前server信息
	server *cluster.Server

	metricsReporters []metrics.Reporter

	rpcServer        cluster.RPCServer
	serviceDiscovery cluster.ServiceDiscovery

	rpcClient              cluster.RPCClient
	router                 *router.Router
	remoteBindingListeners []cluster.RemoteBindingListener

	entityManager EntityManager
	actorSystem   *actor.ActorSystem
}

var _ protos.PitayaServer = &RemoteService{}

func NewRemoteService(
	dieChan chan bool,
	serializer serialize.Serializer,
	server *cluster.Server,
	metricsReporters []metrics.Reporter,

	rpcClient cluster.RPCClient,
	rpcServer cluster.RPCServer,
	sd cluster.ServiceDiscovery,
	router *router.Router,

	system *actor.ActorSystem,
) *RemoteService {

	p := &RemoteService{
		serializer:       serializer,
		appDieChan:       dieChan,
		server:           server,
		metricsReporters: metricsReporters,

		rpcClient:              rpcClient,
		rpcServer:              rpcServer,
		serviceDiscovery:       sd,
		router:                 router,
		remoteBindingListeners: make([]cluster.RemoteBindingListener, 0),

		actorSystem: system,
	}

	return p
}

func (p *RemoteService) Start() {
	// for i := 0; i < 10; i++ {
	// 	go p.Dispatch()
	// }

	// app.config.GetInt("pitaya.concurrency.handler.dispatch")
	const numberDispatch = 10
	for i := 0; i < numberDispatch; i++ {
		go p.dispatch(i)
	}
}

func (p *RemoteService) dispatch(thread int) {
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

func (p *RemoteService) CallEntityFromLocal(
	ctx context.Context,
	entityID,
	entityType string,
	routeStr string,
	reply proto.Message,
	arg proto.Message,
) error {
	return p.localCall(ctx, entityID, entityType, routeStr, reply, arg, 0)
	// rt, err := route.Decode(router)
	// if err != nil {
	// 	// logger.Errorf("cannot decode route:%s err:%s", router, err)
	// 	return err
	// }

	// typName, err := p.entityManager.GetTypName(entityID)
	// if err != nil {
	// 	logger.Warnf("找不到该实体的类型 id:%s err:%s", entityID, err)
	// 	return err
	// }
	// routers := rtManager.getRoute(typName)

	// h, err := routers.getHandler(rt)
	// if err != nil {
	// 	return e.NewError(err, e.ErrNotFoundCode)
	// }

	// processHandlerMessage(nil, rt, h, p.remoteService.serializer)
	return nil
}

func (p *RemoteService) localCall(ctx context.Context, entityID, entityType string, routeStr string, reply proto.Message, arg proto.Message, mid int) error {
	var ret interface{}
	var err error

	entity := p.entityManager.GetEntityVal(entityID, entityType)
	if entity == nil {
		logger.Log.Warnf("pitaya/local process message to entity: entity(id:%s type:%s) not found", entityID, entityType)

		// return &protos.Response{
		// 	Error: &protos.Error{
		// 		Code: e.ErrUnknownCode,
		// 		Msg:  fmt.Sprintf("entity(id:%s type:%s) not found", entityID, entityType),
		// 	},
		// }
		err = e.NewError(fmt.Errorf("entity(id:%s type:%s) not found", entityID, entityType), e.ErrUnknownCode)
		return err
	}

	// TODO 本地 call 这里marshal了一次，到实际call方法时，又unmarshal一次，重复了
	argBytes, _ := p.serializer.Marshal(arg)

	ret, err = p.actorSystem.Root.RequestFuture(
		p.entityManager.GetEntityPid(entityID, entityType),
		&metapart.LocalMessageWrapper{
			Ctx: ctx,
			Req: &protos.Request{
				Type: protos.RPCType_User,
				Msg: &protos.MsgV2{
					Id:    uint64(mid),
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
}

// region remote

// RPC makes rpcs
func (r *RemoteService) RPC(ctx context.Context, entityID, entityType string, serverID string, route *route.Route, reply proto.Message, arg proto.Message) error {
	var data []byte
	var err error
	if arg != nil {
		data, err = proto.Marshal(arg)
		if err != nil {
			return err
		}
	}
	res, err := r.doRPC(ctx, entityID, entityType, serverID, route, data)
	if err != nil {
		return err
	}

	if res.Error != nil {
		return &e.Error{
			Code:     res.Error.Code,
			Message:  res.Error.Msg,
			Metadata: res.Error.Metadata,
		}
	}

	if reply != nil {
		err = proto.Unmarshal(res.GetData(), reply)
		if err != nil {
			return err
		}
	}
	return nil
}

// DoRPC do rpc and get answer
func (r *RemoteService) doRPC(ctx context.Context, entityID, entityType string, serverID string, route *route.Route, protoData []byte) (*protos.Response, error) {
	msg := &message.Message{
		Type:       message.Request,
		Route:      route.Short(),
		Data:       protoData,
		EntityID:   entityID,
		EntityType: entityType,
	}

	target, _ := r.serviceDiscovery.GetServer(serverID)
	if serverID != "" && target == nil {
		return nil, constants.ErrServerNotFound
	}

	return r.remoteCall(ctx, target, protos.RPCType_User, route, nil, msg)
}

func (r *RemoteService) remoteCall(
	ctx context.Context,
	server *cluster.Server,
	rpcType protos.RPCType,
	route *route.Route,
	session *session.Session,
	msg *message.Message,
) (*protos.Response, error) {
	svType := route.SvType

	var err error
	target := server

	if target == nil {
		target, err = r.router.Route(ctx, rpcType, svType, route, msg)
		if err != nil {
			return nil, e.NewError(err, e.ErrInternalCode)
		}
	}

	res, err := r.rpcClient.Call(ctx, rpcType, route, session, msg, target)
	if err != nil {
		return nil, err
	}
	return res, err
}

// Send makes sends
func (r *RemoteService) Send(ctx context.Context, entityID, entityType string, serverID string, route *route.Route, reply proto.Message, arg proto.Message) error {
	var data []byte
	var err error
	if arg != nil {
		data, err = proto.Marshal(arg)
		if err != nil {
			return err
		}
	}
	return r.doSend(ctx, entityID, entityType, serverID, route, data)

	// if res.Error != nil {
	// 	return &e.Error{
	// 		Code:     res.Error.Code,
	// 		Message:  res.Error.Msg,
	// 		Metadata: res.Error.Metadata,
	// 	}
	// }
}

// DoSend do send and not wait for reponse
func (r *RemoteService) doSend(ctx context.Context, entityID, entityType string, serverID string, route *route.Route, protoData []byte) error {
	msg := &message.Message{
		Type:       message.Request,
		Route:      route.Short(),
		Data:       protoData,
		EntityID:   entityID,
		EntityType: entityType,
	}

	target, _ := r.serviceDiscovery.GetServer(serverID)
	if serverID != "" && target == nil {
		return constants.ErrServerNotFound
	}

	return r.remoteSend(ctx, target, protos.RPCType_User, route, nil, msg)
}

func (r *RemoteService) remoteSend(
	ctx context.Context,
	server *cluster.Server,
	rpcType protos.RPCType,
	route *route.Route,
	session *session.Session,
	msg *message.Message,
) error {
	svType := route.SvType

	var err error
	target := server

	if target == nil {
		target, err = r.router.Route(ctx, rpcType, svType, route, msg)
		if err != nil {
			return e.NewError(err, e.ErrInternalCode)
		}
	}

	return r.rpcClient.Post(ctx, rpcType, route, session, msg, target)
}

func (r *RemoteService) RemoteProcess(
	ctx context.Context,
	server *cluster.Server,
	a *agent.Agent,
	route *route.Route,
	msg *message.Message,
) {
	switch msg.Type {
	case message.Request:
		res, err := r.remoteCall(ctx, server, protos.RPCType_Sys, route, a.Session, msg)
		if err != nil {
			logger.Log.Errorf("Failed to process remote(%s): %s", route, err.Error())
			a.AnswerWithError(ctx, msg.ID, err)
			return
		}
		err = a.Session.ResponseMID(ctx, msg.ID, res.Data)
		if err != nil {
			logger.Log.Errorf("Failed to respond remote(%s): %s", route, err.Error())
			a.AnswerWithError(ctx, msg.ID, err)
		}
	case message.Notify:
		err := r.remoteSend(ctx, server, protos.RPCType_Sys, route, a.Session, msg)
		defer tracing.FinishSpan(ctx, err)

		if err != nil {
			logger.Log.Errorf("error while sending a notify: %s", err.Error())
		}
	}
}

// end region remote

// pitaya server imp
// Call processes a remote call
func (r *RemoteService) Call(ctx context.Context, req *protos.Request) (*protos.Response, error) {
	c, err := util.GetContextFromRequest(req, r.server.ID)
	c = util.StartSpanFromRequest(c, r.server.ID, req.GetMsg().GetRoute())
	var res *protos.Response
	if err != nil {
		res = &protos.Response{
			Error: &protos.Error{
				Code: e.ErrInternalCode,
				Msg:  err.Error(),
			},
		}
	} else {
		// res = r.processRemoteMessage(c, req)

		res = r.processRemoteMessage2Entity(c, req)
	}

	if res.Error != nil {
		err = errors.New(res.Error.Msg)
	}

	defer tracing.FinishSpan(c, err)
	return res, nil
}

// PushToUser sends a push to user
func (r *RemoteService) PushToUser(c context.Context, push *protos.Push) (*protos.Response, error) {
	// 去掉这个日志打印 by 涂飞
	// logger.Log.Debugf("sending push to user %s: %v", push.GetUid(), string(push.Data))
	s := session.GetSessionByUID(push.GetUid())
	if s != nil {
		err := s.Push(push.Route, push.Data)
		if err != nil {
			return nil, err
		}
		return &protos.Response{
			Data: []byte("ack"),
		}, nil
	}
	return nil, constants.ErrSessionNotFound
}

// SessionBindRemote is called when a remote server binds a user session and want us to acknowledge it
func (r *RemoteService) SessionBindRemote(c context.Context, msg *protos.BindMsg) (*protos.Response, error) {
	for _, r := range r.remoteBindingListeners {
		r.OnUserBind(msg.Uid, msg.Fid)
	}
	return &protos.Response{
		Data: []byte("ack"),
	}, nil
}

// KickUser sends a kick to user
func (r *RemoteService) KickUser(ctx context.Context, kick *protos.KickMsg) (*protos.KickAnswer, error) {
	logger.Log.Debugf("sending kick to user %s", kick.GetUserId())
	s := session.GetSessionByUID(kick.GetUserId())
	if s != nil {
		err := s.Kick(ctx)
		if err != nil {
			return nil, err
		}
		return &protos.KickAnswer{
			Kicked: true,
		}, nil
	}
	return nil, constants.ErrSessionNotFound
}

// end pitaya server imp

// AddRemoteBindingListener adds a listener
func (r *RemoteService) AddRemoteBindingListener(bindingListener cluster.RemoteBindingListener) {
	r.remoteBindingListeners = append(r.remoteBindingListeners, bindingListener)
}

func (r *RemoteService) processRemoteMessage2Entity(ctx context.Context, req *protos.Request) *protos.Response {
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
		&metapart.LocalMessageWrapper{
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
