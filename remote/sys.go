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

package remote

import (
	"context"

	"github.com/tutumagi/pitaya/component"
	"github.com/tutumagi/pitaya/constants"
	"github.com/tutumagi/pitaya/protos"
	"github.com/tutumagi/pitaya/session"
)

// type SysService struct {
// 	basepart.Entity
// }

// Sys contains logic for handling sys remotes
type Sys struct {
	component.Base
}

// BindSession binds the local session
func (s *Sys) BindSession(ctx context.Context, sessionData *protos.Session) (*protos.Response, error) {
	sess := session.GetSessionByID(sessionData.Id)
	if sess == nil {
		return nil, constants.ErrSessionNotFound
	}
	if err := sess.Bind(ctx, sessionData.Uid); err != nil {
		return nil, err
	}
	return &protos.Response{Data: []byte("ack")}, nil
}

// Kick kicks a local user
func (s *Sys) Kick(ctx context.Context, msg *protos.KickMsg) (*protos.KickAnswer, error) {
	res := &protos.KickAnswer{
		Kicked: false,
	}
	sess := session.GetSessionByUID(msg.GetUserId())
	if sess == nil {
		return res, constants.ErrSessionNotFound
	}
	err := sess.Kick(ctx)
	if err != nil {
		return res, err
	}
	res.Kicked = true
	return res, nil
}

// SwitchSessionOwner switch session owner
func (s *Sys) SwitchSessionOwner(ctx context.Context, msg *protos.SwitchOwner) (*protos.Response, error) {
	sess := session.GetSessionByUID(msg.Sess.Uid)
	if sess == nil {
		return nil, constants.ErrSessionNotFound
	}
	sess.SwitchOwner(msg.Id, msg.Type)

	return &protos.Response{}, nil
}
