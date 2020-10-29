package entity

// ISpace 自定义space接口
type ISpace interface {
	IEntity

	OnSpaceInit()
	OnSpaceCreated()
	OnSpaceDestroy()

	OnEntityEnter(entity *Entity)
	OnEntityLeave(entity *Entity)

	// OnGameReady()
}
