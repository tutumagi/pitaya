package pitaya

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"github.com/tutumagi/pitaya/acceptor"
	"github.com/tutumagi/pitaya/agent"
	"github.com/tutumagi/pitaya/cluster"
	"github.com/tutumagi/pitaya/conn/codec"
	"github.com/tutumagi/pitaya/conn/message"
	"github.com/tutumagi/pitaya/conn/packet"
	"github.com/tutumagi/pitaya/constants"
	pcontext "github.com/tutumagi/pitaya/context"
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

var handlerType = "handler"

type unhandledMessage struct {
	ctx   context.Context
	agent *agent.Agent
	route *route.Route
	msg   *message.Message
}

type EntityManager interface {
	// 通过实体 ID 获取 typeName
	GetTypName(id string) (string, error)
	// 获取实际的实体对象，可能是 cellpart.Entity.Val() ，也可能是 basepart.Entity.Val()，所以用 interface{}
	GetEntityVal(id string, typName string) interface{}
	// 获取实体绑定的pid
	GetEntityPid(id string, typName string) *actor.PID
}

type AppMsgProcessor struct {
	appDieChan      chan bool             // die channel app
	chLocalProcess  chan unhandledMessage // channel of messages that will be processed locally
	chRemoteProcess chan unhandledMessage // channel of messages that will be processed remotely
	serializer      serialize.Serializer  // message serializer

	// 当前server信息
	server *cluster.Server

	metricsReporters []metrics.Reporter

	heartbeatTimeout   time.Duration
	messagesBufferSize int

	// region remote
	rpcServer              cluster.RPCServer
	serviceDiscovery       cluster.ServiceDiscovery
	decoder                codec.PacketDecoder // binary decoder
	encoder                codec.PacketEncoder // binary encoder
	rpcClient              cluster.RPCClient
	router                 *router.Router
	messageEncoder         message.Encoder
	remoteBindingListeners []cluster.RemoteBindingListener
	// region end remote

	entityManager EntityManager
	actorSystem   *actor.ActorSystem
}

var _ protos.PitayaServer = &AppMsgProcessor{}

func NewAppProcessor(
	dieChan chan bool,
	packetDecoder codec.PacketDecoder,
	packetEncoder codec.PacketEncoder,
	serializer serialize.Serializer,
	heartbeatTime time.Duration,
	messagesBufferSize,
	localProcessBufferSize,
	remoteProcessBufferSize int,
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
		chLocalProcess:     make(chan unhandledMessage, localProcessBufferSize),
		chRemoteProcess:    make(chan unhandledMessage, remoteProcessBufferSize),
		decoder:            packetDecoder,
		encoder:            packetEncoder,
		messagesBufferSize: messagesBufferSize,
		serializer:         serializer,
		heartbeatTimeout:   heartbeatTime,
		appDieChan:         dieChan,
		server:             server,
		messageEncoder:     messageEncoder,
		metricsReporters:   metricsReporters,

		rpcClient:              rpcClient,
		rpcServer:              rpcServer,
		serviceDiscovery:       sd,
		router:                 router,
		remoteBindingListeners: make([]cluster.RemoteBindingListener, 0),

		actorSystem: system,
	}

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

// Handle handles messages from a conn
func (h *AppMsgProcessor) Handle(conn acceptor.PlayerConn) {
	// create a client agent and startup write goroutine
	a := agent.NewAgent(conn, h.decoder, h.encoder, h.serializer, h.heartbeatTimeout, h.messagesBufferSize, h.appDieChan, h.messageEncoder, h.metricsReporters)

	// startup agent goroutine
	go a.Handle()
	// go h.processMessage(a)

	logger.Log.Debugf("New session established: %s", a.String())

	// guarantee agent related resource is destroyed
	defer func() {
		// a.Session.Close()
		a.CloseByReason(agent.AgentCloseByMessageEnd)
		logger.Log.Debugf("Session read goroutine exit, SessionID=%d, UID=%d", a.Session.ID(), a.Session.UID())
	}()

	for {
		// logger.Log.Debugf("pitaya.handler begin to get nextmessage for SessionID=%d, UID=%s", a.Session.ID(), a.Session.UID())
		msg, err := conn.GetNextMessage()

		if err != nil {
			logger.Log.Errorf("Error reading next available message: %s", err.Error())
			return
		}

		packets, err := h.decoder.Decode(msg)
		if err != nil {
			logger.Log.Errorf("Failed to decode message: %s", err.Error())
			return
		}

		if len(packets) < 1 {
			logger.Log.Warnf("Read no packets, data: %v", msg)
			continue
		}

		// logger.Log.Debugf("pitaya.handler end to decode nextmessage for SessionID=%d, UID=%s", a.Session.ID(), a.Session.UID())

		// process all packet
		for i := range packets {
			if err := h.processPacket(a, packets[i]); err != nil {
				logger.Log.Errorf("Failed to process packet: %s", err.Error())
				return
			}
		}
	}
}

func (h *AppMsgProcessor) processPacket(a *agent.Agent, p *packet.Packet) error {
	switch p.Type {
	case packet.Handshake:
		logger.Log.Debug("Received handshake packet")
		// logger.Log.Infof("pitaya.handler end to processPacket :handshake packet for SessionID=%d, UID=%s", a.Session.ID(), a.Session.UID())
		if err := a.SendHandshakeResponse(); err != nil {
			logger.Log.Errorf("Error sending handshake response: %s", err.Error())
			return err
		}
		logger.Log.Debugf("Session handshake Id=%d, Remote=%s", a.Session.ID(), a.RemoteAddr())

		// Parse the json sent with the handshake by the client
		handshakeData := &session.HandshakeData{}
		err := json.Unmarshal(p.Data, handshakeData)
		if err != nil {
			a.SetStatus(constants.StatusClosed)
			return fmt.Errorf("Invalid handshake data. Id=%d", a.Session.ID())
		}

		a.Session.SetHandshakeData(handshakeData)
		a.SetStatus(constants.StatusHandshake)
		// err = a.Session.Set(constants.IPVersionKey, a.IPVersion())
		// if err != nil {
		// 	logger.Log.Warnf("failed to save ip version on session: %q\n", err)
		// }

		logger.Log.Debug("Successfully saved handshake data")

	case packet.HandshakeAck:
		a.SetStatus(constants.StatusWorking)
		logger.Log.Debugf("Receive handshake ACK Id=%d, Remote=%s", a.Session.ID(), a.RemoteAddr())
		// logger.Log.Infof("pitaya.handler end to processPacket :handshake ACK for SessionID=%d, UID=%s", a.Session.ID(), a.Session.UID())

		// 连接连成功了
		// a.ownerEntityID
		// RPC(context.TODO(), )
	case packet.Data:
		if a.GetStatus() < constants.StatusWorking {
			return fmt.Errorf("receive data on socket which is not yet ACK, session will be closed immediately, remote=%s",
				a.RemoteAddr().String())
		}

		msg, err := message.Decode(p.Data)
		if err != nil {
			return err
		}

		// logger.Log.Debugf("pitaya.handler begin to processMessage for SessionID=%d, UID=%s, route=%s", a.Session.ID(), a.Session.UID(), msg.Route)
		h.processMessage(a, msg)
		// logger.Log.Debugf("pitaya.handler end to processMessage for SessionID=%d, UID=%s, route=%s", a.Session.ID(), a.Session.UID(), msg.Route)

	case packet.Heartbeat:
		// expected
	}

	a.SetLastAt()
	return nil
}

// 处理连接来的消息
func (p *AppMsgProcessor) processMessage(a *agent.Agent, msg *message.Message) {
	requestID := uuid.New()
	ctx := pcontext.AddToPropagateCtx(context.Background(), constants.StartTimeKey, time.Now().UnixNano())
	ctx = pcontext.AddToPropagateCtx(ctx, constants.RouteKey, msg.Route)
	ctx = pcontext.AddToPropagateCtx(ctx, constants.RequestIDKey, requestID.String())
	tags := opentracing.Tags{
		"local.id":   p.server,
		"span.kind":  "server",
		"msg.type":   strings.ToLower(msg.Type.String()),
		"user.id":    a.Session.UID(),
		"request.id": requestID.String(),
	}
	ctx = tracing.StartSpan(ctx, msg.Route, tags)
	ctx = context.WithValue(ctx, constants.SessionCtxKey, a.Session)

	r, err := route.Decode(msg.Route)
	if err != nil {
		logger.Log.Errorf("Failed to decode route: %s", err.Error())
		a.AnswerWithError(ctx, msg.ID, e.NewError(err, e.ErrBadRequestCode))
		return
	}

	if r.SvType == "" {
		r.SvType = p.server.Type
	}

	//该消息由协程池竞争执行
	message := unhandledMessage{
		ctx:   ctx,
		agent: a,
		route: r,
		msg:   msg,
	}
	if r.SvType == p.server.Type {
		p.chLocalProcess <- message
	} else {
		// if p.remoteService != nil {
		if p.rpcServer != nil {
			p.chRemoteProcess <- message
		} else {
			logger.Log.Warnf("request made to another server type but no remoteService running")
		}
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
		case lm := <-p.chLocalProcess:
			metrics.ReportMessageProcessDelayFromCtx(lm.ctx, p.metricsReporters, "local")
			p.localProcess(lm.ctx, lm.agent, lm.route, lm.msg)

		case rm := <-p.chRemoteProcess:
			metrics.ReportMessageProcessDelayFromCtx(rm.ctx, p.metricsReporters, "remote")
			p.remoteProcess(rm.ctx, nil, rm.agent, rm.route, rm.msg)

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

func (p *AppMsgProcessor) CallEntityFromLocal(
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

func (p *AppMsgProcessor) localProcess(ctx context.Context, a *agent.Agent, route *route.Route, msg *message.Message) {
	var mid uint
	switch msg.Type {
	case message.Request:
		mid = msg.ID
	case message.Notify:
		mid = 0
	}

	var ret interface{}
	var err error

	entityID := msg.EntityID
	entityType := msg.EntityType
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
	}

	// 给指定实体的 actor 发送消息
	ret, err = p.actorSystem.Root.RequestFuture(
		p.entityManager.GetEntityPid(entityID, entityType),
		&metapart.LocalMessageWrapper{
			Ctx: ctx,
			Req: &protos.Request{
				Type: protos.RPCType_Sys,
				Session: &protos.Session{
					Id:     a.Session.ID(),
					Uid:    a.Session.UID(),
					RoleID: a.Session.RoleID(),
				},
				Msg: &protos.MsgV2{
					Id:    uint64(mid),
					Route: msg.Route,
					Data:  msg.Data,
					Type:  protos.MsgType(msg.Type),
					Eid:   msg.EntityID,
					Typ:   msg.EntityType,
					// Reply: ,
				},
			},
		},
		//TODO 这里写的2秒
		2*time.Second,
	).Result()

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
	if msg.Type != message.Notify {
		if err != nil {
			logger.Log.Errorf("Failed to process handler(route:%s) message: %s", route.Short(), err.Error())
			a.AnswerWithError(ctx, mid, err)
		} else {
			err := a.Session.ResponseMID(ctx, mid, ret)
			if err != nil {
				tracing.FinishSpan(ctx, err)
				metrics.ReportTimingFromCtx(ctx, p.metricsReporters, handlerType, err)
			}
		}
	} else {
		metrics.ReportTimingFromCtx(ctx, p.metricsReporters, handlerType, nil)
		tracing.FinishSpan(ctx, err)
	}
}

func (p *AppMsgProcessor) localCall(ctx context.Context, entityID, entityType string, routeStr string, reply proto.Message, arg proto.Message, mid int) error {
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
func (r *AppMsgProcessor) RPC(ctx context.Context, entityID, entityType string, serverID string, route *route.Route, reply proto.Message, arg proto.Message) error {
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
func (r *AppMsgProcessor) doRPC(ctx context.Context, entityID, entityType string, serverID string, route *route.Route, protoData []byte) (*protos.Response, error) {
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

func (r *AppMsgProcessor) remoteCall(
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
func (r *AppMsgProcessor) Send(ctx context.Context, entityID, entityType string, serverID string, route *route.Route, reply proto.Message, arg proto.Message) error {
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
func (r *AppMsgProcessor) doSend(ctx context.Context, entityID, entityType string, serverID string, route *route.Route, protoData []byte) error {
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

func (r *AppMsgProcessor) remoteSend(
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

func (r *AppMsgProcessor) remoteProcess(
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
func (r *AppMsgProcessor) Call(ctx context.Context, req *protos.Request) (*protos.Response, error) {
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
func (r *AppMsgProcessor) PushToUser(c context.Context, push *protos.Push) (*protos.Response, error) {
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
func (r *AppMsgProcessor) SessionBindRemote(c context.Context, msg *protos.BindMsg) (*protos.Response, error) {
	for _, r := range r.remoteBindingListeners {
		r.OnUserBind(msg.Uid, msg.Fid)
	}
	return &protos.Response{
		Data: []byte("ack"),
	}, nil
}

// KickUser sends a kick to user
func (r *AppMsgProcessor) KickUser(ctx context.Context, kick *protos.KickMsg) (*protos.KickAnswer, error) {
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
func (r *AppMsgProcessor) AddRemoteBindingListener(bindingListener cluster.RemoteBindingListener) {
	r.remoteBindingListeners = append(r.remoteBindingListeners, bindingListener)
}

func (r *AppMsgProcessor) processRemoteMessage2Entity(ctx context.Context, req *protos.Request) *protos.Response {
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
