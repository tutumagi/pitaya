package metapart

type ViewLayer = int32

const (
	ViewLayerNormal ViewLayer = 1 // 普通的玩家视角
	ViewLayerDrone  ViewLayer = 2 // 无人机视角
)
