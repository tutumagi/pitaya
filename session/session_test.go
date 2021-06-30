// Copyright (c) TFG Co. All Rights Reserved.
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
	"errors"
	"flag"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	nats "github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/tutumagi/pitaya/constants"
	"github.com/tutumagi/pitaya/helpers"
	"github.com/tutumagi/pitaya/session/mocks"
)

var update = flag.Bool("update", false, "update .golden files")

type someStruct struct {
	A int
	B string
}

type unregisteredStruct struct{}

type mockAddr struct{}

func (ma *mockAddr) Network() string { return "tcp" }
func (ma *mockAddr) String() string  { return "192.0.2.1:25" }

func getEncodedEmptyMap() []byte {
	b, _ := json.Marshal(map[string]interface{}{})
	return b
}

func TestNewSessionIDService(t *testing.T) {
	t.Parallel()

	sessionIDService := newSessionIDService()
	assert.NotNil(t, sessionIDService)
	assert.EqualValues(t, 0, sessionIDService.sid)
}

func TestSessionIDServiceSessionID(t *testing.T) {
	t.Parallel()

	sessionIDService := newSessionIDService()
	sessionID := sessionIDService.sessionID()
	assert.EqualValues(t, 1, sessionID)
}

func TestCloseAll(t *testing.T) {
	var (
		entity *mocks.MockNetworkEntity
	)

	tables := map[string]struct {
		sessions func() []*Session
		mock     func()
	}{
		"test_close_many_sessions": {
			sessions: func() []*Session {
				return []*Session{
					New(entity, uuid.New().String()),
					New(entity, uuid.New().String()),
					New(entity, uuid.New().String()),
				}
			},
			mock: func() {
				entity.EXPECT().Close().Times(3)
			},
		},

		"test_close_no_sessions": {
			sessions: func() []*Session { return []*Session{} },
			mock:     func() {},
		},
	}

	for name, table := range tables {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			entity = mocks.NewMockNetworkEntity(ctrl)
			for _, s := range table.sessions() {
				sessionsByID.Store(s.ID(), s)
				sessionsByUID.Store(s.UID(), s)
			}

			table.mock()

			CloseAll()
		})
	}
}

func TestNew(t *testing.T) {
	tables := []struct {
		name string
		uid  string
	}{
		{"test_frontend", ""},
		{"test_frontend_with_uid", uuid.New().String()},
	}

	for _, table := range tables {
		t.Run(table.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			entity := mocks.NewMockNetworkEntity(ctrl)
			var ss *Session
			if table.uid != "" {
				ss = New(entity, table.uid)
			} else {
				ss = New(entity)
			}
			assert.NotZero(t, ss.id)
			assert.Equal(t, entity, ss.network)
			// assert.Empty(t, ss.data)
			assert.InDelta(t, time.Now().Unix(), ss.lastTime, 1)
			assert.Empty(t, ss.OnCloseCallbacks)

			if len(table.uid) > 0 {
				assert.Equal(t, table.uid[0], ss.uid[0])
			}

			val, ok := sessionsByID.Load(ss.id)
			assert.True(t, ok)
			assert.Equal(t, val, ss)
		})
	}
}

func TestGetSessionByIDExists(t *testing.T) {
	t.Parallel()

	expectedSS := New(nil)
	ss := GetSessionByID(expectedSS.id)
	assert.Equal(t, expectedSS, ss)
}

func TestGetSessionByIDDoenstExist(t *testing.T) {
	t.Parallel()

	ss := GetSessionByID(123456) // huge number to make sure no session with this id
	assert.Nil(t, ss)
}

func TestGetSessionByUIDExists(t *testing.T) {
	uid := uuid.New().String()
	expectedSS := New(nil, uid)
	sessionsByUID.Store(uid, expectedSS)

	ss := GetSessionByUID(uid)
	assert.Equal(t, expectedSS, ss)
}

func TestGetSessionByUIDDoenstExist(t *testing.T) {
	t.Parallel()

	ss := GetSessionByUID(uuid.New().String())
	assert.Nil(t, ss)
}

func TestKick(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	entity := mocks.NewMockNetworkEntity(ctrl)
	ss := New(entity)
	c := context.Background()
	entity.EXPECT().Kick(c)
	entity.EXPECT().Close()
	err := ss.Kick(c)
	assert.NoError(t, err)
}

func TestSessionPush(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockEntity := mocks.NewMockNetworkEntity(ctrl)
	ss := New(mockEntity)
	route := uuid.New().String()
	v := someStruct{A: 1, B: "aaa"}

	mockEntity.EXPECT().Push(route, v)
	err := ss.Push(route, v)
	assert.NoError(t, err)
}

func TestSessionResponseMID(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockEntity := mocks.NewMockNetworkEntity(ctrl)
	ss := New(mockEntity)
	mid := uint(rand.Int())
	v := someStruct{A: 1, B: "aaa"}
	ctx := context.Background()

	mockEntity.EXPECT().ResponseMID(ctx, mid, v)
	err := ss.ResponseMID(ctx, mid, v)
	assert.NoError(t, err)
}

func TestSessionID(t *testing.T) {
	t.Parallel()

	ss := New(nil)
	ss.id = int64(rand.Uint64())

	id := ss.ID()
	assert.Equal(t, ss.id, id)
}

func TestSessionUID(t *testing.T) {
	t.Parallel()

	ss := New(nil)
	ss.uid = uuid.New().String()

	uid := ss.UID()
	assert.Equal(t, ss.uid, uid)
}

func TestSessionBindFailsWithoutUID(t *testing.T) {
	t.Parallel()

	ss := New(nil)
	assert.NotNil(t, ss)

	err := ss.Bind(nil, "")
	assert.Equal(t, constants.ErrIllegalUID, err)
}

func TestSessionBindFailsIfAlreadyBound(t *testing.T) {
	t.Parallel()

	ss := New(nil)
	ss.uid = uuid.New().String()
	assert.NotNil(t, ss)

	err := ss.Bind(nil, uuid.New().String())
	assert.Equal(t, constants.ErrSessionAlreadyBound, err)
}

func TestSessionBindRunsOnSessionBind(t *testing.T) {
	affectedVar := ""
	err := errors.New("some error occured")
	tables := []struct {
		name          string
		onSessionBind func(ctx context.Context, s *Session) error
		err           error
	}{
		{"successful_on_session_bind", func(ctx context.Context, s *Session) error {
			affectedVar = s.uid
			return nil
		}, nil},
		{"failed_on_session_bind", func(ctx context.Context, s *Session) error { return err }, err},
	}

	for _, table := range tables {
		t.Run(table.name, func(t *testing.T) {
			affectedVar = ""
			ss := New(nil)
			assert.NotNil(t, ss)

			OnSessionBind(table.onSessionBind)
			defer func() { sessionBindCallbacks = make([]func(ctx context.Context, s *Session) error, 0) }()

			uid := uuid.New().String()
			err := ss.Bind(nil, uid)

			if table.err != nil {
				assert.Equal(t, table.err, err)
				assert.Empty(t, affectedVar)
				assert.Empty(t, ss.uid)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, uid, affectedVar)
				assert.Equal(t, uid, ss.uid)
			}
		})
	}
}

func TestSessionBindFrontend(t *testing.T) {
	ss := New(nil)
	assert.NotNil(t, ss)

	uid := uuid.New().String()
	err := ss.Bind(nil, uid)
	assert.NoError(t, err)
	assert.Equal(t, uid, ss.uid)

	val, ok := sessionsByUID.Load(uid)
	assert.True(t, ok)
	assert.Equal(t, val, ss)
}

func TestSessionOnClose(t *testing.T) {
	t.Parallel()

	ss := New(nil)
	assert.NotNil(t, ss)

	expected := false
	f := func() { expected = true }
	ss.OnClose(f)
	assert.Len(t, ss.OnCloseCallbacks, 1)

	ss.OnCloseCallbacks[0]()
	assert.True(t, expected)
}

func TestSessionClose(t *testing.T) {
	tables := []struct {
		name string
		uid  string
	}{
		{"close", ""},
		{"close_bound", uuid.New().String()},
	}

	for _, table := range tables {
		t.Run(table.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockEntity := mocks.NewMockNetworkEntity(ctrl)
			ss := New(mockEntity)
			assert.NotNil(t, ss)

			if table.uid != "" {
				sessionsByUID.Store(table.uid, ss)
				ss.uid = table.uid
			}

			mockEntity.EXPECT().Close()
			ss.Close()

			_, ok := sessionsByID.Load(ss.id)
			assert.False(t, ok)

			if table.uid != "" {
				_, ok = sessionsByUID.Load(table.uid)
				assert.False(t, ok)
			}
		})
	}
}

func TestSessionCloseFrontendWithSubscription(t *testing.T) {
	s := helpers.GetTestNatsServer(t)
	defer s.Shutdown()
	var initialSubs uint32 = s.NumSubscriptions()
	conn, err := nats.Connect(fmt.Sprintf("nats://%s", s.Addr()))
	assert.NoError(t, err)
	defer conn.Close()

	subs, err := conn.Subscribe(uuid.New().String(), func(msg *nats.Msg) {})
	assert.NoError(t, err)
	helpers.ShouldEventuallyReturn(t, s.NumSubscriptions, uint32(initialSubs+1))
	helpers.ShouldEventuallyReturn(t, conn.NumSubscriptions, int(1))

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEntity := mocks.NewMockNetworkEntity(ctrl)
	ss := New(mockEntity)
	assert.NotNil(t, ss)
	ss.Subscriptions = []*nats.Subscription{subs}

	mockEntity.EXPECT().Close()
	ss.Close()

	helpers.ShouldEventuallyReturn(t, s.NumSubscriptions, uint32(initialSubs))
	helpers.ShouldEventuallyReturn(t, conn.NumSubscriptions, int(0))
}

func TestSessionRemoteAddr(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEntity := mocks.NewMockNetworkEntity(ctrl)
	ss := New(mockEntity)
	assert.NotNil(t, ss)

	expectedAddr := &mockAddr{}
	mockEntity.EXPECT().RemoteAddr().Return(expectedAddr)
	addr := ss.RemoteAddr()
	assert.Equal(t, expectedAddr, addr)
}

func TestOnSessionBind(t *testing.T) {
	expected := false
	f := func(context.Context, *Session) error {
		expected = true
		return nil
	}
	OnSessionBind(f)
	defer func() { sessionBindCallbacks = make([]func(ctx context.Context, s *Session) error, 0) }()
	assert.NotNil(t, OnSessionBind)

	sessionBindCallbacks[0](context.Background(), nil)
	assert.True(t, expected)
}

// func TestSessionPushToFrontFailsIfFrontend(t *testing.T) {
// 	t.Parallel()

// 	ss := New(nil, true)
// 	assert.NotNil(t, ss)

// 	err := ss.PushToFront(nil)
// 	assert.Equal(t, constants.ErrFrontSessionCantPushToFront, err)
// }

// func TestSessionPushToFront(t *testing.T) {
// 	t.Parallel()
// 	tables := []struct {
// 		name string
// 		err  error
// 	}{
// 		{"successful_request", nil},
// 		{"failed_request", errors.New("failed bind in front")},
// 	}

// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	for _, table := range tables {
// 		t.Run(table.name, func(t *testing.T) {
// 			mockEntity := mocks.NewMockNetworkEntity(ctrl)
// 			ss := New(mockEntity, false)
// 			assert.NotNil(t, ss)
// 			uid := uuid.New().String()
// 			ss.uid = uid

// 			expectedSessionData := &protos.Session{
// 				Id:  ss.frontendSessionID,
// 				Uid: uid,
// 				// Data: ss.encodedData,
// 			}
// 			expectedRequestData, err := proto.Marshal(expectedSessionData)
// 			assert.NoError(t, err)
// 			ctx := context.Background()
// 			mockEntity.EXPECT().SendRequest(ctx, "", "", ss.frontendID, constants.SessionPushRoute, expectedRequestData).Return(nil, table.err)

// 			// err = ss.PushToFront(ctx)
// 			// assert.Equal(t, table.err, err)
// 		})
// 	}
// }

func TestSessionClear(t *testing.T) {
	t.Parallel()

	ss := New(nil)
	assert.NotNil(t, ss)

	ss.uid = uuid.New().String()
	ss.Clear()
}

func TestSessionGetHandshakeData(t *testing.T) {
	t.Parallel()

	data1 := &HandshakeData{
		Sys: HandshakeClientData{
			Platform:    "macos",
			LibVersion:  "2.3.2",
			BuildNumber: "20",
			Version:     "14.0.2",
		},
		User: make(map[string]interface{}),
	}
	data2 := &HandshakeData{
		Sys: HandshakeClientData{
			Platform:    "windows",
			LibVersion:  "2.3.10",
			BuildNumber: "",
			Version:     "ahaha",
		},
		User: map[string]interface{}{
			"ababa": make(map[string]interface{}),
			"pepe":  1,
		},
	}
	tables := []struct {
		name string
		data *HandshakeData
	}{
		{"test_1", data1},
		{"test_2", data2},
	}

	for _, table := range tables {
		t.Run(table.name, func(t *testing.T) {
			ss := New(nil)

			assert.Nil(t, ss.GetHandshakeData())

			ss.handshakeData = table.data

			assert.Equal(t, ss.GetHandshakeData(), table.data)
		})
	}
}

func TestSessionSetHandshakeData(t *testing.T) {
	t.Parallel()

	data1 := &HandshakeData{
		Sys: HandshakeClientData{
			Platform:    "macos",
			LibVersion:  "2.3.2",
			BuildNumber: "20",
			Version:     "14.0.2",
		},
		User: make(map[string]interface{}),
	}
	data2 := &HandshakeData{
		Sys: HandshakeClientData{
			Platform:    "windows",
			LibVersion:  "2.3.10",
			BuildNumber: "",
			Version:     "ahaha",
		},
		User: map[string]interface{}{
			"ababa": make(map[string]interface{}),
			"pepe":  1,
		},
	}
	tables := []struct {
		name string
		data *HandshakeData
	}{
		{"testSessionSetData_1", data1},
		{"testSessionSetData_2", data2},
	}

	for _, table := range tables {
		t.Run(table.name, func(t *testing.T) {
			ss := New(nil)
			ss.SetHandshakeData(table.data)
			assert.Equal(t, table.data, ss.handshakeData)
		})
	}
}
