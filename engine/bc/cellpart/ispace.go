package cellpart

// ISpace 自定义space接口
type ISpace interface {
	ICellEntity

	OnSpaceInit(initDataFromBase map[string]string) error
	OnSpaceCreated()
	OnSpaceDestroy()

	OnEntityEnter(entity *Entity)
	OnEntityLeave(entity *Entity)

	// OnGameReady()
}
