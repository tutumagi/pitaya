package basepart

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/tutumagi/pitaya/cluster"
	"github.com/tutumagi/pitaya/conn/codec"
	"github.com/tutumagi/pitaya/conn/message"
	"github.com/tutumagi/pitaya/engine/bc/metapart"
	"github.com/tutumagi/pitaya/router"
	"github.com/tutumagi/pitaya/serialize"
)

var baseEntManager *_BaseEntityManager
var msgProcessor *metapart.EntityMsgProcessor

func Init(rootSystem *actor.ActorSystem,
	serviceDiscovery cluster.ServiceDiscovery,
	serializer serialize.Serializer,
	encoder codec.PacketEncoder,
	rpcClient cluster.RPCClient,
	router *router.Router,
	messageEncoder message.Encoder,
) {
	baseEntManager = newBaseEntityManager(rootSystem)

	msgProcessor = metapart.NewEntityProcessor(
		serviceDiscovery,
		serializer,
		encoder,
		rpcClient,
		router,
		messageEncoder,
	)

}
