package basepart

// ISpace 自定义space接口
type ISpace interface {
	IBaseEntity

	OnSpaceInit() error
	OnSpaceCreated()
	OnSpaceDestroy()

	OnCellPartCreated() error
	// PrepareCellData 创建cell时传过去的数据
	PrepareCellData() map[string]string
	// OnEntityEnter(entity *Entity)
	// OnEntityLeave(entity *Entity)

	// OnGameReady()
}
