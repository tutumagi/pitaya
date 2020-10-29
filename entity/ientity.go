package entity

import "github.com/tutumagi/pitaya/math32"

// IEntity 自定义实体的一些回调方法
type IEntity interface {
	OnInit()    // 实体构造完成后回调
	OnCreated() // 实体数据初始化完成回调
	OnDestroy() // 实体销毁后回调

	OnEnterSpace()         // 实体进入场景回调
	OnLeaveSpace(s *Space) // 实体离开场景回调

	OnEnterSight(other *Entity) // 其他实体进入视野回调
	OnLeaveSight(other *Entity) // 其他实体离开视野回调

	DefaultModel(id string) interface{} // 默认的数据model
	DefaultPos() math32.Vector3         // 默认的位置
}
