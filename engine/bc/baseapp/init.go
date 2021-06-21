package baseapp

import (
	"strings"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/tutumagi/pitaya/cluster"
	"github.com/tutumagi/pitaya/component"
	"github.com/tutumagi/pitaya/engine/bc"
	"github.com/tutumagi/pitaya/engine/bc/basepart"
	"github.com/tutumagi/pitaya/engine/common"
	"github.com/tutumagi/pitaya/serialize"
)

func Initialize(
	appDieChan chan bool,
	rpcClient cluster.RPCClient,
	serializer serialize.Serializer,
	serviceDiscovery cluster.ServiceDiscovery,
	rootSystem *actor.ActorSystem,
	remoteCaller common.EntityRemoteCaller,
) {
	initEntityService(rpcClient, serializer, serviceDiscovery)
	basepart.Init(appDieChan, serializer, rootSystem, remoteCaller)
}

func initEntityService(rpcClient cluster.RPCClient,
	serializer serialize.Serializer,
	serviceDiscovery cluster.ServiceDiscovery,
) {
	entityServices := basepart.NewRemote(
		rpcClient,
		serializer,
		serviceDiscovery,
	)

	typeDesc := bc.RegisterService("entity", entityServices)

	typeDesc.Routers.RegisterRemote(&basepart.EntityHandler{}, component.WithNameFunc(strings.ToLower))
}
