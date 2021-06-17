package cellpart

import (
	"context"
	"fmt"
	"time"

	"github.com/tutumagi/pitaya/engine/aoi"
	"github.com/tutumagi/pitaya/engine/math32"
	"github.com/tutumagi/pitaya/logger"
	"github.com/tutumagi/pitaya/protos"
	"gitlab.gamesword.com/nut/dreamcity/game/define"

	"github.com/tutumagi/pitaya"
	"github.com/tutumagi/pitaya/timer"
	"go.uber.org/zap"
)

// 每100ms update一次
const tickMs = 100
const tickDelta = time.Millisecond * tickMs

// var now time.Time

// Space the systems container
type Space struct {
	Entity

	kind int32 // 什么空间的类型

	entities *CellSet

	aoiMgr aoi.Systemer
	I      ISpace
	// ctx         context.Context
	// ctxCancelFn context.CancelFunc
}

// AOIDump dump aoi system nodes
func (s *Space) AOIDump() string {
	return s.aoiMgr.Dump()
}

func (s *Space) String() string {
	if s == nil {
		return "nil"
	}

	return fmt.Sprintf("<Space>(ID:%s kind:%d count:%d)", s.ID, s.kind, s.entities.Count())
}

/******************* IEntity interface ***********************/

// OnInit init 回调
func (s *Space) OnInit() error {
	return nil
}

func (s *Space) CellAttrChanged(keys map[string]struct{}) {}

// OnCreated space 创建后回调
func (s *Space) OnCreated() {
	s.onSpaceCreated()

	if s == nil {
		return
	}
	s.I.OnSpaceCreated()

	// if !CurServerUseSpace() {
	// 	// 在 baseapp 时，通知 cellappmgr 场景创建成功了
	// 	err := pitaya.Send(context.TODO(),
	// 		"cellmgrapp.spaceservice.spaceloaded",
	// 		&pb.SpaceLoadedNotify{
	// 			SpaceKind:    s.kind,
	// 			SpaceID:      s.ID,
	// 			BaseServerID: pitaya.GetServerID(),
	// 			CellServerID: s.initCellServerID,
	// 		},
	// 	)
	// 	if err != nil {
	// 		logger.Warn("notify space load failed", zap.Error(err))
	// 	} else {
	// 		logger.Infof("notify space load success %s", s.String())
	// 	}
	// }
}

// OnDestroy on destroy
func (s *Space) OnDestroy() {
	s.I.OnSpaceDestroy()

	s.entities.ForEach(func(e *Entity) {
		e.Destroy()
	})
	s.entities.Clear()

	spaceManager.delSpace(s.ID)
	logger.Debugf("Space::OnDestroy id=%s", s.ID)
}

func (s *Space) onSpaceCreated() {
	spaceManager.putSpace(s)

	if s.kind == define.MasterSpaceKind {
		logger.Infof("create master space success count:%d", s.entities.Count())
	}
	// add space tick
	s.AddCallback(func() {
		s.tick(tickDelta)
	}, tickDelta, timer.LoopForever)
}

func (s *Space) tick(dt time.Duration) {
	s.I.Tick(tickMs)

	// t := time.Now()
	s.entities.ForEach(func(e *Entity) {
		e.I.Tick(tickMs)
	})
	// logger.Debugf("space tick %s", time.Now().Sub(t))
}

/*********************** ISpace interface *********************/

// OnSpaceInit space 初始化完成回调，可以在这里初始化该场景的数据，比如实体数据
func (s *Space) OnSpaceInit(initDataFromBase map[string]string) error {
	return nil
}

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
func (s *Space) EnableAOI() {
	if s.aoiMgr != nil {
		logger.Warnf("%s is already enable", s)
	}

	if s.entities.Count() > 0 {
		logger.Warnf("%s is already using AOI", s)
	}

	s.aoiMgr = aoi.NewCoordSystem()
}

// enterZeroRadiusEntities 批量进入实体，只适用于初始化地图时，加载没有aoi半径的实体
// entities 里面事先需要设置好position
func (s *Space) enterZeroRadiusEntities(entities []*Entity) {
	t := time.Now()
	logger.Infof("enter batch entities %d", len(entities))

	var stat map[string]int = map[string]int{}

	for _, e := range entities {
		// if e.typeDesc.useAOI || e.typeDesc.aoiDistance != 0 {
		// 	panic(fmt.Sprintf("cannot add the entity in the case %s", e))
		// }

		dPos := e.coord.Position()
		s._beforeEnter(e, dPos)

		e.setWitness(aoi.NewWitness())

		typName := e.TypName()
		// 统计每种类型的实体分别插入了多少个
		if num, ok := stat[typName]; ok {
			stat[typName] = num + 1
		} else {
			stat[typName] = 1
		}
	}

	var aoiEntities []aoi.Entityer = make([]aoi.Entityer, 0, len(entities))
	for _, src := range entities {
		aoiEntities = append(aoiEntities, src)
	}

	s.aoiMgr.InsertZeroRadiusEntities(aoiEntities)

	for _, e := range entities {
		e.witness.InstallViewTrigger()
		s._afterEnter(e)
	}
	logger.Infof("enter batch entities %d done. stat:%v %s", len(entities), stat, time.Now().Sub(t))
}

func (s *Space) _beforeEnter(e *Entity, pos *math32.Vector3) {
	// 注意，这里的代码 时序很重要
	e.space = s
	s.entities.Add(e)
	// 这里设置 pos 是因为  实体从 A 场景迁移到 B 场景时， 原来在A场景用到的pos 已经不适用了
	e.coord.ResetFlags()
	if pos != nil {
		e.coord.SetVec3(pos)
	}

	e.I.BeforeEnterSpace()
}

func (s *Space) _afterEnter(e *Entity) {
	s.I.OnEntityEnter(e)
	e.I.AfterEnterSpace()

	s.notifyEntityEnterSpaceResult(e)
}

// 通知实体已经进入到场景
func (s *Space) notifyEntityEnterSpaceResult(e *Entity) {
	// TODO 这里只通知玩家进入场景，可以考虑用依赖反转，给具体实体控制是否需要通知回业务 app
	if e.TypName() != define.TypNamePlayer {
		return
	}
	msg := &protos.EnterSpaceResultNotify{
		Ok:            true,
		ErrCode:       "",
		ErrMsg:        "",
		SpaceID:       s.ID,
		SpaceKind:     s.kind,
		SpaceServerID: curServerID(),
		EntityID:      e.ID,
		EntityLabel:   e.TypName(),
	}
	logger.Infof("notify entity enter space %s", msg.String())
	err := pitaya.SendTo(
		context.TODO(),
		"",
		"",
		e.baseServerID,
		"entity.enterspaceresult",
		msg,
	)
	if err != nil {
		logger.Warnf("notify entity enter space(err:%s) msg:%s", err, msg.String())
	}
}

func (s *Space) enter(entity *Entity, pos math32.Vector3) {
	s._beforeEnter(entity, &pos)

	if s.aoiMgr != nil {
		if entity.IsUseAOI() {
			// witness 设置池
			entity.setWitness(aoi.NewWitness())
			s.aoiMgr.Insert(entity.coord.BaseCoord)
			// 只要插入到aoi系统后，才能安装 ViewTrigger
			entity.witness.InstallViewTrigger()
		}
	}

	s._afterEnter(entity)
}

// enterRef 只允许初始化的时候调用
func (s *Space) enterRef(entity *Entity, pos math32.Vector3, ref *Entity) {
	s._beforeEnter(entity, &pos)

	if s.aoiMgr != nil {
		if entity.IsUseAOI() {
			// witness 设置池
			entity.setWitness(aoi.NewWitness())
			s.aoiMgr.InsertWithRef(entity.coord.BaseCoord, ref.coord.BaseCoord)
			// 只要插入到aoi系统后，才能安装 ViewTrigger
			entity.witness.InstallViewTrigger()
		}
	}

	s._afterEnter(entity)
}

func (s *Space) leave(entity *Entity) {
	s.entities.Del(entity)
	entity.space = nil

	if s.aoiMgr != nil && entity.IsUseAOI() {
		s.aoiMgr.Remove(entity.coord.BaseCoord)
		entity.witness.Detach(entity)
		entity.witness = nil
	}

	s.I.OnEntityLeave(entity)
	entity.I.OnLeaveSpace(s)
}

func (s *Space) move(entity *Entity, newPos *math32.Vector3) {
	if s.aoiMgr == nil {
		logger.Warn("move in space, but space not enable aoi", zap.String("space", s.String()))
		return
	}
	entity.coord.SetVec3(newPos)

	entity.Coord().Update()
}

/****************** 实体相关 *********************/

// CreateEntity creates a new local entity in this space，这里会保存到db里面去注意
// // entityID 可以为空字符串
// func (s *Space) CreateEntity(typName string, entityID string, attrs *attr.StrMap, pos math32.Vector3) *CellEntity {
// 	e := CreateEntity(
// 		typName,
// 		entityID,
// 		attrs,
// 		false,
// 	)
// 	s.enter(e, pos)
// 	logger.Debugf("pitaya.handler Space::CreateEntity e = %+v", e)
// 	return e
// }

// EnterBatchEntities 批量加入实体，只允许初始化时调用
func (s *Space) EnterBatchEntities(entities []*Entity) {
	s.enterZeroRadiusEntities(entities)
}

// EnterEntityWithRef 进入单个实体，只允许初始化地图数据时调用
func (s *Space) EnterEntityWithRef(entity *Entity, pos math32.Vector3, ref *Entity) {
	s.enterRef(entity, pos, ref)
}

// EnterSingleEntity 进入单个实体
func (s *Space) EnterSingleEntity(entity *Entity, pos math32.Vector3) {
	s.enter(entity, pos)
}

// // LoadEntity laod entity in this space
// func (s *Space) LoadEntity(typName string, entityID string) (*CellEntity, error) {
// 	e, err := LoadEntity(typName, entityID)
// 	if err != nil {
// 		return nil, err
// 	}
// 	s.enter(e, math32.InvalidVec3)
// 	return e, nil
// }

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
func (s *Space) GetEntityByType(typName string) CellMap {
	entities := GetEntitiesByType(typName)
	return entities.Filter(func(e *Entity) bool {
		return e.typeDesc.TypName() == typName && s.entities.Contains(e)
	})
}
