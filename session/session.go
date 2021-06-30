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

package session

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	nats "github.com/nats-io/nats.go"
	"github.com/tutumagi/pitaya/constants"
	"github.com/tutumagi/pitaya/logger"
	"github.com/tutumagi/pitaya/protos"
)

// NetworkEntity represent low-level network instance
type NetworkEntity interface {
	Push(route string, v interface{}) error
	ResponseMID(ctx context.Context, mid uint, v interface{}, isError ...bool) error
	Close() error
	Kick(ctx context.Context) error
	RemoteAddr() net.Addr
	SendRequest(ctx context.Context, entityID, entityType string, serverID, route string, v interface{}) (*protos.Response, error)
}

var (
	sessionBindCallbacks = make([]func(ctx context.Context, s *Session) error, 0)
	afterBindCallbacks   = make([]func(ctx context.Context, s *Session) error, 0)
	// SessionCloseCallbacks contains global session close callbacks
	SessionCloseCallbacks = make([]func(s *Session), 0)
	sessionsByUID         sync.Map
	sessionsByID          sync.Map
	sessionIDSvc          = newSessionIDService()
	// SessionCount keeps the current number of sessions
	SessionCount int64
)

// HandshakeClientData represents information about the client sent on the handshake.
type HandshakeClientData struct {
	Platform    string `json:"platform"`
	LibVersion  string `json:"libVersion"`
	BuildNumber string `json:"clientBuildNumber"`
	Version     string `json:"clientVersion"`
}

// HandshakeData represents information about the handshake sent by the client.
// `sys` corresponds to information independent from the app and `user` information
// that depends on the app and is customized by the user.
type HandshakeData struct {
	Sys  HandshakeClientData    `json:"sys"`
	User map[string]interface{} `json:"user,omitempty"`
}

// Session represents a client session, which can store data during the connection.
// All data is released when the low-level connection is broken.
// Session instance related to the client will be passed to Handler method in the
// context parameter.
type Session struct {
	sync.RWMutex        // protect data
	id           int64  // session global unique id
	uid          string // binding user id
	// 和哪个 EntityID 绑定在一起
	ownerEntityID   string
	ownerEntityType string

	lastTime         int64                // last heartbeat time
	network          NetworkEntity        // low-level network entity
	handshakeData    *HandshakeData       // handshake data received by the client
	OnCloseCallbacks []func()             //onClose callbacks
	Subscriptions    []*nats.Subscription // subscription created on bind when using nats rpc server
}

type sessionIDService struct {
	sid int64
}

func newSessionIDService() *sessionIDService {
	return &sessionIDService{
		sid: 0,
	}
}

// SessionID returns the session id
func (c *sessionIDService) sessionID() int64 {
	return atomic.AddInt64(&c.sid, 1)
}

// New returns a new session instance
// a NetworkEntity is a low-level network instance
func New(entity NetworkEntity, UID ...string) *Session {
	s := &Session{
		id:      sessionIDSvc.sessionID(),
		network: entity,
		// data:             make(map[string]interface{}),
		handshakeData:    nil,
		lastTime:         time.Now().Unix(),
		OnCloseCallbacks: []func(){},
	}

	sessionsByID.Store(s.id, s)
	atomic.AddInt64(&SessionCount, 1)

	if len(UID) > 0 {
		s.uid = UID[0]
	}
	return s
}

// GetSessionByUID return a session bound to an user id
func GetSessionByUID(uid string) *Session {
	// TODO: Block this operation in backend servers
	if val, ok := sessionsByUID.Load(uid); ok {
		return val.(*Session)
	}
	return nil
}

// GetSessionByID return a session bound to a frontend server id
func GetSessionByID(id int64) *Session {
	// TODO: Block this operation in backend servers
	if val, ok := sessionsByID.Load(id); ok {
		return val.(*Session)
	}
	return nil
}

// OnSessionBind adds a method to be called when a session is bound
// same function cannot be added twice!
func OnSessionBind(f func(ctx context.Context, s *Session) error) {
	// Prevents the same function to be added twice in onSessionBind
	sf1 := reflect.ValueOf(f)
	for _, fun := range sessionBindCallbacks {
		sf2 := reflect.ValueOf(fun)
		if sf1.Pointer() == sf2.Pointer() {
			return
		}
	}
	sessionBindCallbacks = append(sessionBindCallbacks, f)
}

// OnAfterSessionBind adds a method to be called when session is bound and after all sessionBind callbacks
func OnAfterSessionBind(f func(ctx context.Context, s *Session) error) {
	// Prevents the same function to be added twice in onSessionBind
	sf1 := reflect.ValueOf(f)
	for _, fun := range afterBindCallbacks {
		sf2 := reflect.ValueOf(fun)
		if sf1.Pointer() == sf2.Pointer() {
			return
		}
	}
	afterBindCallbacks = append(afterBindCallbacks, f)
}

// OnSessionClose adds a method that will be called when every session closes
func OnSessionClose(f func(s *Session)) {
	sf1 := reflect.ValueOf(f)
	for _, fun := range SessionCloseCallbacks {
		sf2 := reflect.ValueOf(fun)
		if sf1.Pointer() == sf2.Pointer() {
			return
		}
	}
	SessionCloseCallbacks = append(SessionCloseCallbacks, f)
}

// CloseAll calls Close on all sessions
func CloseAll() {
	logger.Log.Infof("closing all sessions, %d sessions", SessionCount)
	sessionsByID.Range(func(_, value interface{}) bool {
		s := value.(*Session)
		s.Close()
		return true
	})
	logger.Log.Info("finished closing sessions")
}

// func (s *Session) updateEncodedData() error {
// 	var b []byte
// 	b, err := dataEncoder(s.data)
// 	if err != nil {
// 		return err
// 	}
// 	s.encodedData = b
// 	return nil
// }

// DataEncoder 自定义session data encoder
type DataEncoder func(data interface{}) ([]byte, error)

// DataDecoder 自定义session data encoder
type DataDecoder func(bytes []byte, data interface{}) error

var dataEncoder DataEncoder = json.Marshal
var dataDecoder DataDecoder = json.Unmarshal

// SetCustomEncodeDecode 设置自定义的session data encoder和decoder
// 	默认为 json.Marshal 和 json.Unmarshal
func SetCustomEncodeDecode(encoder DataEncoder, decoder DataDecoder) {
	dataEncoder = encoder
	dataDecoder = decoder
}

// Push message to client
func (s *Session) Push(route string, v interface{}) error {
	return s.network.Push(route, v)
}

// ResponseMID responses message to client, mid is
// request message ID
func (s *Session) ResponseMID(ctx context.Context, mid uint, v interface{}, err ...bool) error {
	return s.network.ResponseMID(ctx, mid, v, err...)
}

// ID returns the session id
func (s *Session) ID() int64 {
	return s.id
}

// UID returns uid that bind to current session
func (s *Session) UID() string {
	return s.uid
}

func (s *Session) OwnerEntityID() string {
	return s.ownerEntityID
}

func (s *Session) OwnerEntityType() string {
	return s.ownerEntityType
}

func (s *Session) Owner() (ownerID string, ownerType string) {
	return s.ownerEntityID, s.ownerEntityType
}

func (s *Session) SwitchOwner(id string, typ string) {
	logger.Infof("session switch owner old id:%s typ:%s", s.ownerEntityID, s.ownerEntityType)
	s.ownerEntityID = id
	s.ownerEntityType = typ
	logger.Infof("session switch owner new id:%s typ:%s", id, typ)
}

// Bind bind UID to current session
func (s *Session) Bind(ctx context.Context, uid string) error {
	if uid == "" {
		return constants.ErrIllegalUID
	}

	if s.UID() != "" {
		return constants.ErrSessionAlreadyBound
	}

	s.uid = uid
	for _, cb := range sessionBindCallbacks {
		err := cb(ctx, s)
		if err != nil {
			s.uid = ""
			return err
		}
	}

	for _, cb := range afterBindCallbacks {
		err := cb(ctx, s)
		if err != nil {
			s.uid = ""
			return err
		}
	}

	// if code running on frontend server
	sessionsByUID.Store(uid, s)

	return nil
}

// Kick kicks the user
func (s *Session) Kick(ctx context.Context) error {
	err := s.network.Kick(ctx)
	if err != nil {
		return err
	}
	return s.network.Close()
}

// OnClose adds the function it receives to the callbacks that will be called
// when the session is closed
func (s *Session) OnClose(c func()) {
	s.OnCloseCallbacks = append(s.OnCloseCallbacks, c)
}

// Close terminates current session, session related data will not be released,
// all related data should be cleared explicitly in Session closed callback
func (s *Session) Close() {
	atomic.AddInt64(&SessionCount, -1)
	sessionsByID.Delete(s.ID())
	sessionsByUID.Delete(s.UID())
	// TODO: this logic should be moved to nats rpc server
	if s.Subscriptions != nil && len(s.Subscriptions) > 0 {
		// if the user is bound to an userid and nats rpc server is being used we need to unsubscribe
		for _, sub := range s.Subscriptions {
			err := sub.Unsubscribe()
			if err != nil {
				logger.Log.Errorf("error unsubscribing to user's messages channel: %s, this can cause performance and leak issues", err.Error())
			} else {
				logger.Log.Debugf("successfully unsubscribed to user's %s messages channel", s.UID())
			}
		}
	}
	s.network.Close()
}

// RemoteAddr returns the remote network address.
func (s *Session) RemoteAddr() net.Addr {
	return s.network.RemoteAddr()
}

// Clear releases all data related to current session
func (s *Session) Clear() {
	s.Lock()
	defer s.Unlock()

	s.uid = ""
}

// SetHandshakeData sets the handshake data received by the client.
func (s *Session) SetHandshakeData(data *HandshakeData) {
	s.Lock()
	defer s.Unlock()

	s.handshakeData = data
}

// GetHandshakeData gets the handshake data received by the client.
func (s *Session) GetHandshakeData() *HandshakeData {
	return s.handshakeData
}

func (s *Session) DebugString() string {
	return fmt.Sprintf("(ID:%d UID:%s)", s.id, s.uid)
}
