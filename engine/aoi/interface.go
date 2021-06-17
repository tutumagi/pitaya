package aoi

// Systemer 坐标系统
type Systemer interface {
	Insert(coord *BaseCoord)
	Remove(coord *BaseCoord)

	// Update 坐标移动时调用
	Update(coord *BaseCoord, newX float32, newZ float32)

	// insertWithRef 和 InsertZeroRadiusEntities 都是根据目前的游戏业务做的具体优化

	// InsertWithRef 根据参考点插入coord
	InsertWithRef(coord *BaseCoord, ref *BaseCoord)
	// InsertZeroRadiusEntities 初始化需要插入大量aoi半径为0的实体的优化插入
	InsertZeroRadiusEntities(entities []Entityer)

	// Dump 当前所有节点信息
	Dump() string
}

// Entityer 某个坐标节点绑定的实体
type Entityer interface {
	AoiID() string
	// 实体绑定的坐标节点
	Coord() *EntityCoord

	// 有实体进入视野
	OnEnterAOI(other Entityer)
	// 有实体离开视野
	OnLeaveAOI(other Entityer)

	// 实体的监听者
	Witness() *Witness
	// // 添加订阅者，表示 other 订阅了当前实体进入/离开/更新
	// AddWitnessed(other Entityer)
	// // 删除订阅者，表示 other 取消订阅当前实体的进入/离开/更新
	// DelWitnessed(other Entityer)
}

// ViewCallback 视野回调
type ViewCallback interface {
	onEnter(node *BaseCoord)
	onLeave(node *BaseCoord)
}
