package basepart

import (
	"github.com/tutumagi/pitaya/engine/math32"
)

// IBaseEntity 自定义实体的一些回调方法
type IBaseEntity interface {
	OnInit() error // 实体构造完成后回调
	OnCreated()    // 实体数据初始化完成回调
	OnDestroy()    // 实体销毁后回调

	EnterSpaceFailed(err error)                        // 实体进入场景失败了
	EnterSpaceSuccess(spaceID string, spaceKind int32) // 只会在业务 app 触发

	// BeforeEnterSpace()     // 实体进入场景前回调，还没触发 aoi
	// AfterEnterSpace()      // 实体进入场景后回调，已经触发aoi 了
	// OnLeaveSpace(s *Space) // 实体离开场景回调

	OnMigrateOut() // 实体迁移出去后
	OnMigrateIn()  // 实体迁移进来后

	// OnEnterSight(other *Entity) // 其他实体进入视野回调
	// OnLeaveSight(other *Entity) // 其他实体离开视野回调

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

	// CellAttrChanged(keys map[string]struct{})

	// 客户端断开连接
	OnClientDisconnected()
	// 客户端连接上
	OnClientConnected()
}
