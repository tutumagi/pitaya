package basepart

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/tutumagi/pitaya/engine/bc/metapart"
	"github.com/tutumagi/pitaya/engine/common"
	"github.com/tutumagi/pitaya/serialize"
)

var baseEntManager *_BaseEntityManager
var msgProcessor *metapart.EntityMsgProcessor
var caller *Caller

func Init(
	appDieChan chan bool,
	serializer serialize.Serializer,
	rootSystem *actor.ActorSystem,
	remoteCaller common.EntityRemoteCaller,
) {
	baseEntManager = newBaseEntityManager(rootSystem)

	msgProcessor = metapart.NewEntityProcessor(
		serializer,
	)
	caller = NewAppProcessor(
		appDieChan,
		serializer,
		rootSystem,
		remoteCaller,
	)
}
