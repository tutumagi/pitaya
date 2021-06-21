package route

import (
	"strings"

	"github.com/tutumagi/pitaya/component"
	"github.com/tutumagi/pitaya/engine/bc"
	"github.com/tutumagi/pitaya/examples/demo/cluster_protobuf/baseapp/entity"
	"github.com/tutumagi/pitaya/examples/demo/cluster_protobuf/baseapp/server/handler"
)

func RegisterRoute() {
	{
		typedesc := bc.RegisterEntity("account", &entity.Account{}, nil, false)
		typedesc.Routers.Register(&handler.AccountHandler{}, component.WithName("account"), component.WithNameFunc(strings.ToLower))
	}

	// {
	// 	typedesc := bc.RegisterService("room", &services)
	// }

}
