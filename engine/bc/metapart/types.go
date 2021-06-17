package metapart

import (
	"context"

	"github.com/tutumagi/pitaya/protos"
)

type ViewLayer = int32

const (
	ViewLayerNormal ViewLayer = 1 // 普通的玩家视角
	ViewLayerDrone  ViewLayer = 2 // 无人机视角
)

// call给实际实体的参数类型
type LocalMessageWrapper struct {
	Ctx context.Context
	Req *protos.Request
}
