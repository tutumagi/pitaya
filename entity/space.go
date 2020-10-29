package entity

import (
	"fmt"
	"time"

	"github.com/tutumagi/pitaya/aoi"
	"github.com/tutumagi/pitaya/logger"
	"github.com/tutumagi/pitaya/math32"
)

// 每50ms update一次
const updateDelta = time.Millisecond * 50

var now time.Time

// Space the systems container
type Space struct {
	Entity

	kind int32 // 什么空间的类型

	entities Set
	aoiMgr   *aoi.XZListManager
	I        ISpace

	// ctx         context.Context
	// ctxCancelFn context.CancelFunc
}

func (s *Space) String() string {
	if s == nil {
		return "nil"
	}

	return fmt.Sprintf("<Space>(%s count:%d)", s.ID, len(s.entities))
}

/******************* IEntity interface ***********************/

// OnInit init 回调
func (s *Space) OnInit() {
	// TODO 这里 2000 预估的
	s.entities = make(Set, 2000)
	s.I = s.Entity.I.(ISpace)

	s.I.OnSpaceInit()
}

// OnCreated space 创建后回调
func (s *Space) OnCreated() {
	if s == nil {
		return
	}
	s.onSpaceCreated()

	s.I.OnSpaceCreated()

	// err := pitaya.Send(context.TODO(),
	// 	"cellappmgr.spaceservice.spaceloaded",
	// 	&protos.SpaceLoaded{
	// 		Kind:     s.kind,
	// 		SpaceID:  s.ID,
	// 		ServerID: pitaya.GetServerID(),
	// 	},
	// )
	// if err != nil {
	// 	logger.Log.Warnf("notify space load failed %s", err)
	// }
}

// OnDestroy on destroy
func (s *Space) OnDestroy() {
	s.I.OnSpaceDestroy()

	for e := range s.entities {
		e.Destroy()
	}

	spaceManager.delSpace(s.ID)
}

// DefaultModel 默认的数据
func (s *Space) DefaultModel(id string) interface{} {
	return nil
}

func (s *Space) onSpaceCreated() {
	spaceManager.putSpace(s)

	if s.kind == 0 {
	}
}

/*********************** ISpace interface *********************/

// OnSpaceInit space 初始化完成回调
func (s *Space) OnSpaceInit() {}

// OnSpaceCreated space 创建完成回调（数据已经初始化完成）
func (s *Space) OnSpaceCreated() {}

// OnSpaceDestroy space destroy后回调
func (s *Space) OnSpaceDestroy() {}

// OnEntityEnter 实体进入后回调
func (s *Space) OnEntityEnter(entity *Entity) {}

// OnEntityLeave 实体离开后回调
func (s *Space) OnEntityLeave(entity *Entity) {}

/*********************** AOI ***********************/

// EnableAOI enable aoi
func (s *Space) EnableAOI(defaultAOIDistance int32) {
	if s.aoiMgr != nil {
		logger.Log.Warnf("%s is already enable", s)
	}

	if len(s.entities) > 0 {
		logger.Log.Warnf("%s is already using AOI", s)
	}

	s.aoiMgr = aoi.NewXZListAOIManager(aoi.Coord(defaultAOIDistance))
}

func (s *Space) enableAOI(aoiDis int32) {
	s.aoiMgr = aoi.NewXZListAOIManager(aoi.Coord(aoiDis))
}

func (s *Space) enter(entity *Entity, pos math32.Vector3) {

	entity.Space = s
	s.entities.Add(entity)
	// 这里设置 pos 是因为  实体从 A 场景迁移到 B 场景时， 原来在A场景用到的pos 已经不适用了
	entity.pos = pos

	if s.aoiMgr != nil {
		if entity.IsUseAOI() {
			s.aoiMgr.Enter(entity.aoi, aoi.Coord(pos.X), aoi.Coord(pos.Z))
		}
	}

	s.I.OnEntityEnter(entity)
	entity.I.OnEnterSpace()
}

func (s *Space) leave(entity *Entity) {
	s.entities.Del(entity)
	entity.Space = nil

	if s.aoiMgr != nil && entity.typeDesc.useAOI {
		s.aoiMgr.Leave(entity.aoi)
	}

	s.I.OnEntityLeave(entity)
	entity.I.OnLeaveSpace(s)
}

func (s *Space) move(entity *Entity, newPos math32.Vector3) {
	if s.aoiMgr == nil {
		logger.Log.Warnf("move in space %s, but space not enable aoi", s)
		return
	}
	entity.pos = newPos
	s.aoiMgr.Moved(entity.aoi, aoi.Coord(newPos.X), aoi.Coord(newPos.Z))
}

// func (s *Space) onSpaceCreated() {
// 	spaceManager.putSpace(s)
// }

/****************** 实体相关 *********************/

// CreateEntity creates a new local entity in this space
// entityID 可以为空字符串
func (s *Space) CreateEntity(typName string, entityID string, model interface{}, pos math32.Vector3) *Entity {
	return createEntity(typName, entityID, model, s, pos, true)
}

// LoadEntity laod entity in this space
func (s *Space) LoadEntity(typName string, entityID string) *Entity {
	return loadEntity(typName, entityID, s)
}

/****************** Getter *****************/

// SpaceID the world bind spaceid
func (s *Space) SpaceID() string {
	return s.ID
}

// Kind 空间类型
func (s *Space) Kind() int32 {
	return s.kind
}

// GetEntity 获取实体
func (s *Space) GetEntity(typName string, id string) *Entity {
	entity := GetEntity(typName, id)
	if entity == nil {
		return nil
	}
	if s.entities.Contains(entity) {
		return entity
	}
	return nil
}

// GetEntityByType 根据typName 查询 实体
func (s *Space) GetEntityByType(typName string) Map {
	entities := GetEntitiesByType(typName)
	return entities.Filter(func(e *Entity) bool {
		return e.typName == typName
	})
}
