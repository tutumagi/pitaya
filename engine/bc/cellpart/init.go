package cellpart

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/tutumagi/pitaya/cluster"
	"github.com/tutumagi/pitaya/constants"
	"github.com/tutumagi/pitaya/engine/bc/metapart"
	"github.com/tutumagi/pitaya/logger"
	"github.com/tutumagi/pitaya/protos"
	"github.com/tutumagi/pitaya/serialize"
	"github.com/tutumagi/pitaya/util"
)

var cellEntManager *_CellEntityManager
var msgProcessor *metapart.EntityMsgProcessor
var caller metapart.Caller

var rpcClient cluster.RPCClient
var serializer serialize.Serializer

func Init(
	appDieChan chan bool,
	serializer1 serialize.Serializer,
	rootSystem *actor.ActorSystem,
	// remoteCaller common.EntityRemoteCaller,
	rpcClient1 cluster.RPCClient,
) {
	cellEntManager = newCellEntityManager(rootSystem)

	msgProcessor = metapart.NewEntityProcessor(
		serializer1,
	)
	// caller = caller

	rpcClient = rpcClient1
	serializer = serializer1
}

// SendPushToUsers sends a message to the given list of users
func SendPushToUsers(route string, v interface{}, uids []string, frontendType string) ([]string, error) {
	data, err := util.SerializeOrRaw(serializer, v)
	if err != nil {
		return uids, err
	}

	if frontendType == "" {
		return uids, constants.ErrFrontendTypeNotSpecified
	}

	var notPushedUids []string

	// logger.Log.Debugf("Type=PushToUsers Route=%s, Data=%+v, SvType=%s, #Users=%d", route, v, frontendType, len(uids))
	// 注释by 涂飞
	// logger.Log.Debugf("Type=PushToUsers Route=%s, SvType=%s, #Users=%d", route, frontendType, len(uids))

	for _, uid := range uids {
		if rpcClient != nil {
			push := &protos.Push{
				Route: route,
				Uid:   uid,
				Data:  data,
			}
			if err = rpcClient.SendPush(uid, &cluster.Server{Type: frontendType}, push); err != nil {
				notPushedUids = append(notPushedUids, uid)
				logger.Log.Errorf("RPCClient send message error, UID=%s, SvType=%s, Error=%s", uid, frontendType, err.Error())
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
