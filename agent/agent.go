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
	"encoding/binary"
	gojson "encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/tutumagi/pitaya/acceptor"
	"github.com/tutumagi/pitaya/conn/codec"
	"github.com/tutumagi/pitaya/conn/message"
	"github.com/tutumagi/pitaya/conn/packet"
	"github.com/tutumagi/pitaya/constants"
	"github.com/tutumagi/pitaya/logger"
	"github.com/tutumagi/pitaya/metrics"
	"github.com/tutumagi/pitaya/protos"
	"github.com/tutumagi/pitaya/serialize"
	"github.com/tutumagi/pitaya/session"
	"github.com/tutumagi/pitaya/tracing"
	"github.com/tutumagi/pitaya/util"
	"github.com/tutumagi/pitaya/util/compression"

	opentracing "github.com/opentracing/opentracing-go"
	e "github.com/tutumagi/pitaya/errors"
)

const handlerType = "handler"

var (
	// hbd contains the heartbeat packet data 8个字节，装64位整型的时间戳
	hbd []byte = make([]byte, 8)
	// hrd contains the handshake response data
	hrd  []byte
	once sync.Once
)

type AgentCloseReason int32

const (
	_ AgentCloseReason = iota
	AgentCloseByWriteEnd
	AgentCloseByHeartBeat
	AgentCloseByHandleEnd
	AgentCloseByMessageEnd
)

func (a AgentCloseReason) String() string {
	switch a {
	case AgentCloseByWriteEnd:
		return "AgentCloseByWriteEnd"
	case AgentCloseByHeartBeat:
		return "AgentCloseByHeartBeat"
	case AgentCloseByHandleEnd:
		return "AgentCloseByHandleEnd"
	case AgentCloseByMessageEnd:
		return "AgentCloseByMessageEnd"
	default:
		return "UnknownReason"
	}
}

type (
	// Agent corresponds to a user and is used for storing raw Conn information
	Agent struct {
		Session         *session.Session  // session
		appDieChan      chan bool         // app die channel
		chDie           chan struct{}     // wait for close
		chSend          chan pendingWrite // push message queue
		chHbSend        chan pendingWrite // push message queue (心跳专用)
		chStopHeartbeat chan struct{}     // stop heartbeats
		chStopWrite     chan struct{}     // stop writing messages
		// ChRoleMessages  chan UnhandledRoleMessage // 用户请求的消息列表(队列)
		closeMutex sync.Mutex
		conn       net.Conn // low-level conn fd
		// conn acceptor.PlayerConn // low-level conn fd

		decoder            codec.PacketDecoder // binary decoder
		encoder            codec.PacketEncoder // binary encoder
		heartbeatTimeout   time.Duration
		lastAt             int64 // last heartbeat unix time stamp
		messageEncoder     message.Encoder
		messagesBufferSize int // size of the pending messages buffer
		metricsReporters   []metrics.Reporter
		serializer         serialize.Serializer // message serializer
		state              int32                // current agent state
	}

	pendingMessage struct {
		ctx     context.Context
		typ     message.Type // message type
		route   string       // message route (push)
		mid     uint         // response message id (response)
		payload interface{}  // payload
		err     bool         // if its an error message
	}

	pendingWrite struct {
		ctx  context.Context
		data []byte
		err  error
	}

	// UnhandledRoleMessage struct {
	// 	Ctx   context.Context
	// 	Route *route.Route
	// 	Msg   *message.Message
	// }

	// unhandledMessage struct {
	// 	ctx   context.Context
	// 	agent *Agent
	// 	route *route.Route
	// 	msg   *message.Message
	// }
)

// func (a *Agent) Receive(ctx actor.Context) {
// 	switch msg := ctx.Message().(type) {
// 	case *actor.Started:
// 		logger.Log.Info("Agent Starting, initialize actor here")

// 		a.Start(ctx)
// 	case *actor.Stopping:
// 		logger.Log.Info("Agent Stopping, actor is about to shut down")
// 	case *actor.Stopped:
// 		logger.Log.Info("Agent Stopped, actor and its children are stopped")
// 	case *actor.Restarting:
// 		logger.Log.Info("Agent Restarting, actor is about to restart")
// 	case *actor.ReceiveTimeout:
// 		logger.Log.Info("Agent ReceiveTimeout: %v", ctx.Self().String())

// 	default:
// 		logger.Log.Errorf("unknown message %v", msg)
// 	}
// }

// func (a *Agent) InitActor(ctx actor.Context) {
// 	props := actor.PropsFromProducer(func() actor.Actor {
// 		return a
// 	})
// 	a.pid = ctx.SpawnPrefix(props, "agent")
// }

func (a *Agent) Start() {
	// if a.ChRoleMessages == nil {
	// 	// TODO 可配置
	// 	// a.ChRoleMessages = make(chan UnhandledRoleMessage, h.MessageChanSize)
	// 	a.ChRoleMessages = make(chan UnhandledRoleMessage, 100)
	// }

	// startup agent goroutine
	go a.Handle()

	logger.Log.Debugf("New session established: %s", a.String())

	// // guarantee agent related resource is destroyed
	// defer func() {
	// 	// a.Session.Close()
	// 	a.CloseByReason(AgentCloseByMessageEnd)
	// 	logger.Log.Debugf("Session read goroutine exit, Session:", a.Session.DebugString())
	// }()

	// // 处理该连接的 packet 的主循环
	// for {
	// 	// logger.Log.Debugf("pitaya.handler begin to get nextmessage for SessionID=%d, UID=%s", a.Session.ID(), a.Session.UID())
	// 	msg, err := a.conn.GetNextMessage()

	// 	if err != nil {
	// 		logger.Log.Errorf("Error reading next available message(session:) err: %s", a.Session.DebugString(), err.Error())
	// 		return
	// 	}

	// 	packets, err := a.decoder.Decode(msg)
	// 	if err != nil {
	// 		logger.Log.Errorf("Failed to decode message: %s", err.Error())
	// 		return
	// 	}

	// 	if len(packets) < 1 {
	// 		logger.Log.Warnf("Read no packets, data: %v", msg)
	// 		continue
	// 	}

	// 	// logger.Log.Debugf("pitaya.handler end to decode nextmessage for SessionID=%d, UID=%s", a.Session.ID(), a.Session.UID())

	// 	// process all packet
	// 	for i := range packets {
	// 		if err := a.processPacket(packets[i]); err != nil {
	// 			logger.Log.Errorf("Failed to process packet: %s", err.Error())
	// 			return
	// 		}
	// 	}
	// }
}

// NewAgent create new agent instance
func NewAgent(
	conn acceptor.PlayerConn,
	packetDecoder codec.PacketDecoder,
	packetEncoder codec.PacketEncoder,
	serializer serialize.Serializer,
	heartbeatTime time.Duration,
	messagesBufferSize int,
	dieChan chan bool,
	messageEncoder message.Encoder,
	metricsReporters []metrics.Reporter,
) *Agent {
	// initialize heartbeat and handshake data on first user connection
	serializerName := serializer.GetName()

	once.Do(func() {
		hbdEncode(heartbeatTime, packetEncoder, messageEncoder.IsCompressionEnabled(), serializerName)
	})

	a := &Agent{
		appDieChan:         dieChan,
		chDie:              make(chan struct{}),
		chSend:             make(chan pendingWrite, messagesBufferSize),
		chHbSend:           make(chan pendingWrite, messagesBufferSize),
		chStopHeartbeat:    make(chan struct{}),
		chStopWrite:        make(chan struct{}),
		messagesBufferSize: messagesBufferSize,
		conn:               conn,
		decoder:            packetDecoder,
		encoder:            packetEncoder,
		heartbeatTimeout:   heartbeatTime,
		lastAt:             time.Now().Unix(),
		serializer:         serializer,
		state:              constants.StatusStart,
		messageEncoder:     messageEncoder,
		metricsReporters:   metricsReporters,
	}

	// binding session
	s := session.New(a)
	metrics.ReportNumberOfConnectedClients(metricsReporters, session.SessionCount)
	a.Session = s
	return a
}

func (a *Agent) getMessageFromPendingMessage(pm pendingMessage) (*message.Message, error) {
	payload, err := util.SerializeOrRaw(a.serializer, pm.payload)
	if err != nil {
		payload, err = util.GetErrorPayload(a.serializer, err)
		if err != nil {
			return nil, err
		}
	}

	// construct message and encode
	m := &message.Message{
		Type:  pm.typ,
		Data:  payload,
		Route: pm.route,
		ID:    pm.mid,
		Err:   pm.err,
	}

	return m, nil
}

func (a *Agent) packetEncodeMessage(m *message.Message) ([]byte, error) {
	em, err := a.messageEncoder.Encode(m)
	if err != nil {
		return nil, err
	}

	// packet encode
	p, err := a.encoder.Encode(packet.Data, em)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (a *Agent) send(pendingMsg pendingMessage) (err error) {
	defer func() {
		if erro := recover(); erro != nil {
			err = e.NewError(constants.ErrBrokenPipe, e.ErrClientClosedRequest)
		}
	}()
	a.reportChannelSize()

	m, err := a.getMessageFromPendingMessage(pendingMsg)
	if err != nil {
		return err
	}

	// packet encode
	p, err := a.packetEncodeMessage(m)
	if err != nil {
		return err
	}

	pWrite := pendingWrite{
		ctx:  pendingMsg.ctx,
		data: p,
	}

	if pendingMsg.err {
		pWrite.err = util.GetErrorFromPayload(a.serializer, m.Data)
	}

	// chSend is never closed so we need this to don't block if agent is already closed
	select {
	case a.chSend <- pWrite:
	case <-a.chDie:
	}
	return
}

// Push implementation for session.NetworkEntity interface
func (a *Agent) Push(route string, v interface{}) error {
	if a.GetStatus() == constants.StatusClosed {
		return e.NewError(constants.ErrBrokenPipe, e.ErrClientClosedRequest)
	}

	// 注释 by 涂飞，日志打印太多
	// switch d := v.(type) {
	// case []byte:
	// 	logger.Log.Debugf("Type=Push, ID=%d, UID=%d, Route=%s, Data=%dbytes",
	// 		a.Session.ID(), a.Session.UID(), route, len(d))
	// default:
	// 	// logger.Log.Debugf("Type=Push, ID=%d, UID=%d, Route=%s, Data=%+v",
	// 	// 	a.Session.ID(), a.Session.UID(), route, v)
	// 	logger.Log.Debugf("Type=Push, ID=%d, UID=%d, Route=%s",
	// 		a.Session.ID(), a.Session.UID(), route)
	// }
	return a.send(pendingMessage{typ: message.Push, route: route, payload: v})
}

// ResponseMID implementation for session.NetworkEntity interface
// Respond message to session
func (a *Agent) ResponseMID(ctx context.Context, mid uint, v interface{}, isError ...bool) error {
	err := false
	if len(isError) > 0 {
		err = isError[0]
	}
	if a.GetStatus() == constants.StatusClosed {
		return e.NewError(constants.ErrBrokenPipe, e.ErrClientClosedRequest)
	}

	if mid <= 0 {
		return constants.ErrSessionOnNotify
	}

	switch d := v.(type) {
	case []byte:
		logger.Log.Debugf("Type=Response, ID=%d, UID=%d, MID=%d, Data=%dbytes",
			a.Session.ID(), a.Session.UID(), mid, len(d))
	default:
		logger.Log.Infof("Type=Response, ID=%d, UID=%d, MID=%d, Data=%+v",
			a.Session.ID(), a.Session.UID(), mid, v)
	}

	return a.send(pendingMessage{ctx: ctx, typ: message.Response, mid: mid, payload: v, err: err})
}

// Close closes the agent, cleans inner state and closes low-level connection.
// Any blocked Read or Write operations will be unblocked and return errors.
func (a *Agent) Close() error {
	a.closeMutex.Lock()
	defer a.closeMutex.Unlock()
	if a.GetStatus() == constants.StatusClosed {
		return constants.ErrCloseClosedSession
	}
	a.SetStatus(constants.StatusClosed)

	logger.Log.Debugf("Session closed, ID=%d, UID=%s, IP=%s",
		a.Session.ID(), a.Session.UID(), a.conn.RemoteAddr())

	// prevent closing closed channel
	select {
	case <-a.chDie:
		// expect
	default:
		close(a.chStopWrite)
		close(a.chStopHeartbeat)
		close(a.chDie)
		// close(a.ChRoleMessages)
		onSessionClosed(a.Session)
	}

	metrics.ReportNumberOfConnectedClients(a.metricsReporters, session.SessionCount)

	return a.conn.Close()
}

// RemoteAddr implementation for session.NetworkEntity interface
// returns the remote network address.
func (a *Agent) RemoteAddr() net.Addr {
	return a.conn.RemoteAddr()
}

// String, implementation for Stringer interface
func (a *Agent) String() string {
	return fmt.Sprintf("Remote=%s, LastTime=%d", a.conn.RemoteAddr().String(), atomic.LoadInt64(&a.lastAt))
}

// GetStatus gets the status
func (a *Agent) GetStatus() int32 {
	return atomic.LoadInt32(&a.state)
}

func (a *Agent) ChDie() chan struct{} {
	return a.chDie
}

// Kick sends a kick packet to a client
func (a *Agent) Kick(ctx context.Context) error {
	// packet encode
	p, err := a.encoder.Encode(packet.Kick, nil)
	if err != nil {
		return err
	}
	_, err = a.conn.Write(p)
	logger.Warnf("kick agent(sess:%s) err:%s", a.Session.DebugString(), err)
	return err
}

// SetLastAt sets the last at to now
func (a *Agent) SetLastAt() {
	atomic.StoreInt64(&a.lastAt, time.Now().Unix())
}

// SetStatus sets the agent status
func (a *Agent) SetStatus(state int32) {
	atomic.StoreInt32(&a.state, state)
}

// Handle handles the messages from and to a client
func (a *Agent) Handle() {
	// 处理写
	go a.write()
	// 处理心跳
	go a.heartbeat()
	// 处理读
	// 这里只有一个协程在访问 actor.Context，不会有问题
	// go a.processGameMessage(ctx)
	// // 处理读
	// go a.processPackets()
	select {
	case <-a.chDie: // agent closed signal
		return
	}
}

// IPVersion returns the remote address ip version.
// net.TCPAddr and net.UDPAddr implementations of String()
// always construct result as <ip>:<port> on both
// ipv4 and ipv6. Also, to see if the ip is ipv6 they both
// check if there is a colon on the string.
// So checking if there are more than one colon here is safe.
func (a *Agent) IPVersion() string {
	version := constants.IPv4

	ipPort := a.RemoteAddr().String()
	if strings.Count(ipPort, ":") > 1 {
		version = constants.IPv6
	}

	return version
}

func (a *Agent) heartbeat() {
	ticker := time.NewTicker(a.heartbeatTimeout)

	defer func() {
		ticker.Stop()
		a.CloseByReason(AgentCloseByHeartBeat)
		close(a.chHbSend)
		if e := recover(); e != nil {
			logger.Log.Warnf("heartbeat err=%+v", e)
		}
	}()

	for {
		select {
		case <-ticker.C:
			now := time.Now()
			deadline := now.Add(-2 * a.heartbeatTimeout).Unix()
			if atomic.LoadInt64(&a.lastAt) < deadline {
				logger.Log.Debugf("Session heartbeat timeout, LastTime=%d, Deadline=%d", atomic.LoadInt64(&a.lastAt), deadline)
				return
			}

			// 时间戳，毫秒
			ts := uint64(now.UnixNano() / int64(time.Millisecond))
			binary.BigEndian.PutUint64(hbd, ts)

			bytes, err := a.encoder.Encode(packet.Heartbeat, hbd)
			if err != nil {
				logger.Log.Warn("encode heartbeat err %s", err)
			}
			// logger.Log.Debugf("heartbeat chHbSend <-")
			// a.chHbSend <- pendingWrite{data: bytes}
			// chSend is never closed so we need this to don't block if agent is already closed
			select {
			case a.chSend <- pendingWrite{data: bytes}:
			case <-a.chDie:
				return
			case <-a.chStopHeartbeat:
				return
			}
		case <-a.chDie:
			return
		case <-a.chStopHeartbeat:
			return
		}
	}
}

func onSessionClosed(s *session.Session) {
	defer func() {
		if err := recover(); err != nil {
			logger.Log.Errorf("pitaya/onSessionClosed: %v", err)
		}
	}()

	for _, fn1 := range s.OnCloseCallbacks {
		fn1()
	}

	for _, fn2 := range session.SessionCloseCallbacks {
		fn2(s)
	}
}

// SendHandshakeResponse sends a handshake response
func (a *Agent) SendHandshakeResponse() error {
	_, err := a.conn.Write(a.hrdEncodeInner())
	return err
}

func (a *Agent) write() {
	// clean func
	defer func() {
		close(a.chSend)
		a.CloseByReason(AgentCloseByWriteEnd)
	}()

	for {
		select {
		case pWrite := <-a.chSend:
			//TODO 搞明白什么情况下会发生什么事情
			// a.conn.SetWriteDeadline(time.Now().Add(50 * time.Millisecond))
			// close agent if low-level Conn broken
			if _, err := a.conn.Write(pWrite.data); err != nil {
				tracing.FinishSpan(pWrite.ctx, err)
				metrics.ReportTimingFromCtx(pWrite.ctx, a.metricsReporters, handlerType, err)
				logger.Log.Errorf("Failed to write in conn: %s", err.Error())
				return
			}
			var e error
			tracing.FinishSpan(pWrite.ctx, e)
			metrics.ReportTimingFromCtx(pWrite.ctx, a.metricsReporters, handlerType, pWrite.err)
		case pWrite := <-a.chHbSend:
			//TODO 搞明白什么情况下会发生什么事情
			// a.conn.SetWriteDeadline(time.Now().Add(50 * time.Millisecond))
			// logger.Log.Debugf("heartbeat chHbSend ->")
			// close agent if low-level Conn broken
			if _, err := a.conn.Write(pWrite.data); err != nil {
				tracing.FinishSpan(pWrite.ctx, err)
				metrics.ReportTimingFromCtx(pWrite.ctx, a.metricsReporters, handlerType, err)
				logger.Log.Errorf("Failed to write in conn: %s", err.Error())
				return
			}
		case <-a.chStopWrite:
			return
		}
	}
}

// SendRequest sends a request to a server
func (a *Agent) SendRequest(ctx context.Context, entityID, entityType string, serverID, route string, v interface{}) (*protos.Response, error) {
	return nil, errors.New("not implemented")
}

// AnswerWithError answers with an error
func (a *Agent) AnswerWithError(ctx context.Context, mid uint, err error) {
	var e error
	defer func() {
		if e != nil {
			tracing.FinishSpan(ctx, e)
			metrics.ReportTimingFromCtx(ctx, a.metricsReporters, handlerType, e)
		}
	}()
	if ctx != nil && err != nil {
		s := opentracing.SpanFromContext(ctx)
		if s != nil {
			tracing.LogError(s, err.Error())
		}
	}
	p, e := util.GetErrorPayload(a.serializer, err)
	if e != nil {
		logger.Log.Errorf("error answering the user with an error: %s", e.Error())
		return
	}
	e = a.Session.ResponseMID(ctx, mid, p, true)
	if e != nil {
		logger.Log.Errorf("error answering the user with an error: %s", e.Error())
	}
}

func hbdEncode(heartbeatTimeout time.Duration, packetEncoder codec.PacketEncoder, dataCompression bool, serializerName string) {
	hData := map[string]interface{}{
		"code": 200,
		"sys": map[string]interface{}{
			"heartbeat":  heartbeatTimeout.Seconds(),
			"severtime":  uint64(time.Now().UnixNano() / int64(time.Millisecond)), // 时间戳，毫秒
			"dict":       map[string]uint16{},                                     //message.GetDictionary(),
			"serializer": serializerName,
		},
	}
	data, err := gojson.Marshal(hData)
	if err != nil {
		panic(err)
	}

	if dataCompression {
		compressedData, err := compression.DeflateData(data)
		if err != nil {
			panic(err)
		}

		if len(compressedData) < len(data) {
			data = compressedData
		}
	}

	hrd, err = packetEncoder.Encode(packet.Handshake, data)
	if err != nil {
		panic(err)
	}

	// hbd, err = packetEncoder.Encode(packet.Heartbeat, nil)
	// if err != nil {
	// 	panic(err)
	// }
}

//上边函数的一个副本，由于severtime是一个变量，每次握手都是不一样的，不能 once.Do(hbdEncode)
func (a *Agent) hrdEncodeInner() []byte {
	hrdBuff := []byte{}
	hData := map[string]interface{}{
		"code": 200,
		"sys": map[string]interface{}{
			"heartbeat":  a.heartbeatTimeout.Seconds(),
			"severtime":  uint64(time.Now().UnixNano() / int64(time.Millisecond)), // 时间戳，毫秒
			"dict":       map[string]uint16{},                                     //message.GetDictionary(),
			"serializer": a.serializer.GetName(),
		},
	}
	data, err := gojson.Marshal(hData)
	if err != nil {
		logger.Log.Warnf("hrdEncodeInner gojson.Marshal error:%v", err)
		return hrdBuff
	}

	if a.messageEncoder.IsCompressionEnabled() {
		compressedData, err := compression.DeflateData(data)
		if err != nil {
			logger.Log.Warnf("hrdEncodeInner DeflateData error:%v", err)
			return hrdBuff
		}

		if len(compressedData) < len(data) {
			data = compressedData
		}
	}

	hrdBuff, err = a.encoder.Encode(packet.Handshake, data)
	if err != nil {
		logger.Log.Warnf("hrdEncodeInner encoder.Encode error:%v", err)
		return hrdBuff
	}

	return hrdBuff
}

func (a *Agent) reportChannelSize() {
	chSendCapacity := a.messagesBufferSize - len(a.chSend)
	if chSendCapacity == 0 {
		logger.Log.Warnf("chSend is at maximum capacity cap:%d", a.messagesBufferSize)
	}
	for _, mr := range a.metricsReporters {
		if err := mr.ReportGauge(metrics.ChannelCapacity, map[string]string{"channel": "agent_chsend"}, float64(chSendCapacity)); err != nil {
			logger.Log.Warnf("failed to report chSend channel capaacity: %s", err.Error())
		}
	}
}

func (a *Agent) CloseByReason(rs AgentCloseReason) {
	logger.Log.Debugf("CloseByReason rs = %s", rs)
	a.Close()
}
