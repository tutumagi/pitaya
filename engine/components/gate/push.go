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

package gate

import (
	"github.com/tutumagi/pitaya/constants"
	"github.com/tutumagi/pitaya/logger"
	"github.com/tutumagi/pitaya/session"
	"github.com/tutumagi/pitaya/util"
)

// SendPushToUsers sends a message to the given list of users
func SendPushToUsers(route string, v interface{}, uids []string) ([]string, error) {
	data, err := util.SerializeOrRaw(app.serializer, v)
	if err != nil {
		return uids, err
	}

	var notPushedUids []string

	// logger.Log.Debugf("Type=PushToUsers Route=%s, Data=%+v, SvType=%s, #Users=%d", route, v, frontendType, len(uids))
	// 注释by 涂飞
	// logger.Log.Debugf("Type=PushToUsers Route=%s, SvType=%s, #Users=%d", route, frontendType, len(uids))

	for _, uid := range uids {
		if s := session.GetSessionByUID(uid); s != nil {
			if err := s.Push(route, data); err != nil {
				notPushedUids = append(notPushedUids, uid)
				logger.Log.Errorf("Session push message error, ID=%d, UID=%d, Error=%s",
					s.ID(), s.UID(), err.Error())
			}
		} else {
			notPushedUids = append(notPushedUids, uid)
		}
	}

	if len(notPushedUids) != 0 {
		return notPushedUids, constants.ErrPushingToUsers
	}

	return nil, nil
}
