package gate

import (
	"context"
	"encoding/json"
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
	"github.com/tutumagi/pitaya/engine/common"
	services "github.com/tutumagi/pitaya/engine/common"

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
)

const HandlerType = "handler"

type unhandledMessage struct {
	ctx   context.Context
	agent *agent.Agent
	route *route.Route
	msg   *message.Message
}

type GateProcessor struct {
	appDieChan      chan bool             // die channel app
	chLocalProcess  chan unhandledMessage // channel of messages that will be processed locally
	chRemoteProcess chan unhandledMessage // channel of messages that will be processed remotely

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

	entityManager common.EntityManager
	actorSystem   *actor.ActorSystem

	caller *common.Caller
}

func NewGateProcessor(
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
) *GateProcessor {
	remote := services.NewRemoteService(dieChan, serializer, server, metricsReporters, rpcClient, rpcServer, sd, router, system, nil)
	p := &GateProcessor{
		chLocalProcess:     make(chan unhandledMessage, localProcessBufferSize),
		chRemoteProcess:    make(chan unhandledMessage, remoteProcessBufferSize),
		decoder:            packetDecoder,
		encoder:            packetEncoder,
		messagesBufferSize: messagesBufferSize,
		serializer:         serializer,
		heartbeatTimeout:   heartbeatTime,
		appDieChan:         dieChan,
		server:             server,
		metricsReporters:   metricsReporters,
		messageEncoder:     messageEncoder,
		remote:             remote,

		actorSystem: system,
	}

	p.caller = common.NewCaller(p)
	return p
}

func (p GateProcessor) Start() {
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
func (h GateProcessor) Handle(conn acceptor.PlayerConn) {
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

func (h GateProcessor) processPacket(a *agent.Agent, p *packet.Packet) error {
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
		// bootEntityType := app.config.GetString("pitaya.bootentity")
		a.OwnerEntityID = metapart.NewUUID()

		logger.Warnf("创建bootEntity id:%s", a.OwnerEntityID)

		// TODO 这里写死请求到某个服务类型
		err := h.caller.CallService(
			context.TODO(),
			"entity",
			"baseapp.entity.clientconnected",
			&protos.Response{},
			&protos.ClientConnect{
				Sess: &protos.Session{
					Id:       a.Session.ID(),
					Uid:      a.Session.UID(),
					RoleID:   a.Session.RoleID(),
					ServerID: app.server.ID,
				},
				BootEntityID: a.OwnerEntityID,
			},
		)
		if err != nil {
			logger.Warnf("创建bootEntity失败 %s", err)
			// TODO 这里需要关闭此连接
		} else {
			logger.Debugf("创建bootEntity成功")
		}
		// a.ownerEntityID
		// RPC(context.TODO(), )
		// bootEntityID :=
		// h.remote.Call()
	case packet.Data:
		if a.GetStatus() < constants.StatusWorking {
			return fmt.Errorf("receive data on socket which is not yet ACK, session will be closed immediately, remote=%s",
				a.RemoteAddr().String())
		}

		msg, err := message.Decode(p.Data)
		if err != nil {
			return err
		}
		msg.EntityID = a.OwnerEntityID
		msg.EntityType = "account"
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
func (p GateProcessor) processMessage(a *agent.Agent, msg *message.Message) {
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
		if p.remote != nil {
			p.chRemoteProcess <- message
		} else {
			logger.Log.Warnf("request made to another server type but no remoteService running")
		}
	}
}

func (p GateProcessor) Dispatch(thread int) {
	// TODO: This timer is being stopped multiple times, it probably doesn't need to be stopped here
	// defer timer.GlobalTicker.Stop()
	defer func() {
		// logger.Log.Warnf("Go HandlerService::Dispatch(%d) exit", thread)
		logger.Errorf("Go HandlerService::Dispatch(%d) exit", thread)
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
			p.remote.RemoteProcess(rm.ctx, nil, rm.agent, rm.route, rm.msg)

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

func (p GateProcessor) localProcess(ctx context.Context, a *agent.Agent, route *route.Route, msg *message.Message) {
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
		&common.LocalMessageWrapper{
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
				metrics.ReportTimingFromCtx(ctx, p.metricsReporters, common.HandlerType, err)
			}
		}
	} else {
		metrics.ReportTimingFromCtx(ctx, p.metricsReporters, common.HandlerType, nil)
		tracing.FinishSpan(ctx, err)
	}
}

func (p GateProcessor) Call(
	ctx context.Context,
	serverID string,
	entityID,
	entityType string,
	routeStr string,
	reply proto.Message,
	arg proto.Message,
) error {
	droute, err := route.Decode(routeStr)
	if err != nil {
		return err
	}
	if reply != nil {
		logger.Log.Debugf("pitaya/remote call message to entity: entity(id:%s type:%s) route:%s serverID:%s", entityID, entityType, routeStr, serverID)
		return p.remote.RPC(ctx, entityID, entityType, "", droute, reply, arg)
	} else {
		logger.Log.Debugf("pitaya/remote send message to entity: entity(id:%s type:%s) route:%s serverID:%s", entityID, entityType, routeStr, serverID)
		return p.remote.Send(ctx, entityID, entityType, "", droute, arg)
	}
}
