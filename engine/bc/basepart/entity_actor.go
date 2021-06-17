package basepart

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/tutumagi/pitaya/engine/bc/metapart"
	"github.com/tutumagi/pitaya/logger"
)

// 内置的系统消息
type (
	clientDisconnect struct{}
	clientConnect    struct{}
)

var (
	clientDisconnectMsg = &clientDisconnect{}
	clientConnectMsg    = &clientConnect{}
)

// Receive 实现 actor
func (e *Entity) Receive(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *actor.Started:
		logger.Log.Info("Entity Starting, initialize actor here")
	case *actor.Stopping:
		logger.Log.Info("Entity Stopping, actor is about to shut down")
	case *actor.Stopped:
		logger.Log.Info("Entity Stopped, actor and its children are stopped")
	case *actor.Restarting:
		logger.Log.Info("Entity Restarting, actor is about to restart")
	case *actor.ReceiveTimeout:
		logger.Log.Info("Entity ReceiveTimeout: %v", ctx.Self().String())

	case *clientDisconnect:
		e.I.OnClientDisconnected()
	case *clientConnect:
		e.I.OnClientConnected()
	case *metapart.LocalMessageWrapper:
		// 开始处理业务逻辑消息
		ret := msgProcessor.ProcessMessage(msg.Ctx, msg.Req, e, e.typeDesc.Routers)
		if ret != nil {
			ctx.Respond(ret)
		}
	default:
		logger.Log.Errorf("unknown message %v", msg)
	}
}
