package basepart

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/tutumagi/pitaya/cluster"
	"github.com/tutumagi/pitaya/conn/message"
	"github.com/tutumagi/pitaya/engine/bc/metapart"
	"github.com/tutumagi/pitaya/metrics"
	"github.com/tutumagi/pitaya/router"
	"github.com/tutumagi/pitaya/serialize"
)

var baseEntManager *_BaseEntityManager
var msgProcessor *metapart.EntityMsgProcessor
var caller *Caller

func Init(
	dieChan chan bool,
	serializer serialize.Serializer,
	server *cluster.Server,
	messageEncoder message.Encoder,
	metricsReporters []metrics.Reporter,

	rpcClient cluster.RPCClient,
	rpcServer cluster.RPCServer,
	sd cluster.ServiceDiscovery,
	router *router.Router,

	actorSystem *actor.ActorSystem,
) {
	baseEntManager = newBaseEntityManager(actorSystem)

	msgProcessor = metapart.NewEntityProcessor(
		serializer,
	)
	caller = NewAppProcessor(
		dieChan,
		serializer,
		server,
		messageEncoder,
		metricsReporters,

		rpcClient,
		rpcServer,
		sd,
		router,

		actorSystem,
	)

	caller.Start()
}
