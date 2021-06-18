package cellpart

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/tutumagi/pitaya/engine/bc/metapart"
	"github.com/tutumagi/pitaya/engine/common"
	"github.com/tutumagi/pitaya/serialize"
)

var cellEntManager *_CellEntityManager
var msgProcessor *metapart.EntityMsgProcessor
var caller metapart.Caller

func Init(
	rootSystem *actor.ActorSystem,
	serializer serialize.Serializer,
	remoteService *common.RemoteService,
	caller metapart.Caller,
) {
	cellEntManager = newCellEntityManager(rootSystem)

	msgProcessor = metapart.NewEntityProcessor(
		serializer,
	)
	caller = caller
}
