package cellpart

import "github.com/tutumagi/pitaya/engine/math32"

type ICellEntity interface {
	OnInit() error // 实体构造完成后回调
	OnCreated()    // 实体数据初始化完成回调
	OnDestroy()    // 实体销毁后回调

	EnterSpaceFailed(err error)                        // 实体进入场景失败了
	EnterSpaceSuccess(spaceID string, spaceKind int32) // 只会在业务 app 触发

	BeforeEnterSpace()     // 实体进入场景前回调，还没触发 aoi
	AfterEnterSpace()      // 实体进入场景后回调，已经触发aoi 了
	OnLeaveSpace(s *Space) // 实体离开场景回调

	OnMigrateOut() // 实体迁移出去后
	OnMigrateIn()  // 实体迁移进来后

	OnEnterSight(other *Entity) // 其他实体进入视野回调
	OnLeaveSight(other *Entity) // 其他实体离开视野回调

	DefaultPos() math32.Vector3 // 默认的位置，从db load 实体时，加入场景时的位置，会调用到此方法

	// Tick dt 毫秒
	Tick(dt int32)

	// 下面四个回调，每次仅会回调一个

	// OnPositionYawChanged 实体的位置和朝向发生了变化
	OnPositionYawChanged(newPos *math32.Vector3, yaw float32)
	// // OnPositionChanged 实体的位置发生了变化
	// OnPositionChanged(newPos *math32.Vector3)
	// // OnYawChanged 实体的朝向发生了变化
	// OnYawChanged(yaw float32)
	// OnStopMove 实体停止移动了
	OnStopMove(newPos *math32.Vector3, yaw float32)

	CellAttrChanged(keys map[string]struct{})
}

// type Default struct{}

// func (*Default) OnInit() error // 实体构造完成后回调 {}
// func (*Default) OnCreated()    // 实体数据初始化完成回调 {}
// func (*Default) OnDestroy()    // 实体销毁后回调 {}

// func (*Default) EnterSpaceFailed(err error)                        // 实体进入场景失败了 {}
// func (*Default) EnterSpaceSuccess(spaceID string, spaceKind int32) // 只会在业务 app 触发 {}

// func (*Default) BeforeEnterSpace()     // 实体进入场景前回调，还没触发 aoi {}
// func (*Default) AfterEnterSpace()      // 实体进入场景后回调，已经触发aoi 了 {}
// func (*Default) OnLeaveSpace(s *Space) // 实体离开场景回调 {}

// func (*Default) OnMigrateOut() // 实体迁移出去后 {}
// func (*Default) OnMigrateIn()  // 实体迁移进来后 {}

// func (*Default) OnEnterSight(other *CellEntity) // 其他实体进入视野回调 {}
// func (*Default) OnLeaveSight(other *CellEntity) // 其他实体离开视野回调 {}

// func (*Default) DefaultPos() math32.Vector3 // 默认的位置，从db load 实体时，加入场景时的位置，会调用到此方法 {}

// // Tick dt 毫秒
// func (*Default) Tick(dt int32) {}

// // 下面四个回调，每次仅会回调一个

// // OnPositionYawChanged 实体的位置和朝向发生了变化
// func (*Default) OnPositionYawChanged(newPos *math32.Vector3, yaw float32) {}

// // OnStopMove 实体停止移动了
// func (*Default) OnStopMove(newPos *math32.Vector3, yaw float32) {}

// func (*Default) CellAttrChanged(keys map[string]struct{}) {}
