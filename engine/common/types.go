package common

import (
	"context"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/tutumagi/pitaya/protos"
)

const HandlerType = "handler"

type EntityManager interface {
	// 通过实体 ID 获取 typeName
	GetTypName(id string) (string, error)
	// 获取实际的实体对象，可能是 cellpart.Entity.Val() ，也可能是 basepart.Entity.Val()，所以用 interface{}
	GetEntityVal(id string, typName string) interface{}
	// 获取实体绑定的pid
	GetEntityPid(id string, typName string) *actor.PID
}

// call给实际实体的参数类型
type LocalMessageWrapper struct {
	Ctx context.Context
	Req *protos.Request
}
