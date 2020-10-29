package entity

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"
	"unsafe"

	"github.com/tutumagi/pitaya/aoi"
	"github.com/tutumagi/pitaya/logger"

	"github.com/tutumagi/pitaya/common"
	"github.com/tutumagi/pitaya/math32"
	"github.com/tutumagi/pitaya/timer"
	"go.uber.org/zap"
)

type EntityStorager interface {
	Write(tlb string, condKey string, condVal string, data interface{})
	Load(tlb string, condKey string, condVal string, data interface{}) (int, error)
}

var storage EntityStorager

// SaveEntityFunc 保存实体方法
// type SaveEntityFunc func(tlb string, condKey string, condVal string, data interface{})

// var saveEntityFunc SaveEntityFunc

// SetStorage 设置保存持久化读写
func SetStorage(s EntityStorager) {
	if storage == nil {
		storage = s
	}
}

// Entity 实体
type Entity struct {
	ID      string
	typName string
	Data    interface{}

	v        reflect.Value
	I        IEntity
	typeDesc *TypeDesc

	destroyed bool

	Space *Space

	pos math32.Vector3
	yaw float32

	aoi          *aoi.Item
	InterestedIn Set // 关心集合
	InterestedBy Set // 被关心集合

	rawTimers map[*timer.Timer]struct{}

	enteringSpaceRequest struct {
		SpaceID  string
		EnterPos math32.Vector3
		// 请求进入场景时的时间，单位为 time.Duration 纳秒
		RequestTime          int64
		migrateRequestIsSent bool
	}
}

// NewEntity ctor
// func NewEntity(id string, data interface{}, pos math32.Vector3) *Entity {
// 	return &Entity{
// 		ID:           id,
// 		Data:         data,
// 		pos:          pos,
// 		InterestedIn: Set{},
// 		InterestedBy: Set{},
// 	}
// }

func (e *Entity) String() string {
	// return fmt.Sprintf("<Entity>label:%s id:%s data:%+v", e.typName, e.ID, e.Data)
	return fmt.Sprintf("<Entity>label:%s id:%s pos:%s", e.typName, e.ID, e.pos)
}

func (e *Entity) init(typName string, eid string, entityInstance reflect.Value) {
	e.ID = eid
	e.v = entityInstance
	e.I = entityInstance.Interface().(IEntity)
	e.typName = typName

	e.typeDesc = GetTypeDesc(typName)

	e.rawTimers = make(map[*timer.Timer]struct{})

	e.InterestedIn = Set{}
	e.InterestedBy = Set{}

	if e.typeDesc.useAOI {
		e.aoi = aoi.NewItem(aoi.Coord(e.typeDesc.aoiDistance), e, e)
	}

	e.I.OnInit()
}

/************************ Getter/Setter *************************/

// GetPosition get pos
func (e *Entity) GetPosition() math32.Vector3 {
	return e.pos
}

// SetPosition set pos
func (e *Entity) SetPosition(pos math32.Vector3) {
	e.setPosition(pos, e.yaw)
}

func (e *Entity) setPosition(pos math32.Vector3, yaw float32) {
	space := e.Space
	if space == nil {
		logger.Log.Warnf("%s setPosition %s. space is nil", e, pos)
		return
	}

	space.move(e, pos)
	e.yaw = yaw

	// TODO 标记此实体是需要同步位置的
}

// IsUseAOI 该实体是否开启AOI
func (e *Entity) IsUseAOI() bool {
	return e.typeDesc.useAOI
}

// TypName 该实体的类别
func (e *Entity) TypName() string {
	return e.typName
}

// IsSpaceEntity 此实体是否是空间
func (e *Entity) IsSpaceEntity() bool {
	return strings.Contains(e.typName, _SpaceEntityType)
}

// AsSpace 类型转换为space
func (e *Entity) AsSpace() *Space {
	if e.IsSpaceEntity() == false {
		logger.Log.Errorf("%s is not a space", e)
	}

	return (*Space)(unsafe.Pointer(e))
}

// Val 返回实际用户定义的 实体interface
//	 使用方式：landEntity := e.Val().(*LandEntity)
func (e *Entity) Val() interface{} {
	return e.v.Interface()
}

/************************ IEntity interface default imp ***********************/

// OnInit IEntity 的接口，构造完成后回调
func (e *Entity) OnInit() {

}

// OnCreated IEntity 的接口，构造完成，初始化数据后会回调
func (e *Entity) OnCreated() {

}

// OnDestroy IEntity的接口，销毁时回调
func (e *Entity) OnDestroy() {

}

// OnEnterSpace IEntity的接口，进入场景后回调
func (e *Entity) OnEnterSpace() {

}

// OnLeaveSpace IEntity的接口，离开场景后回调
func (e *Entity) OnLeaveSpace(s *Space) {

}

// OnEnterSight IEntity的接口，其他实体进入视野时回调
func (e *Entity) OnEnterSight(other *Entity) {
}

// OnLeaveSight IEntity的接口，其他实体离开视野时回调
func (e *Entity) OnLeaveSight(other *Entity) {

}

// DefaultModel 默认的数据model
func (e *Entity) DefaultModel(id string) interface{} {
	logger.Log.Warnf("%s entity return nil model", e.String())
	return nil
}

// DefaultPos 默认的位置，不会触发在 space中的 move api。 只会在加载/创建实体时，数据创建完成后调用
func (e *Entity) DefaultPos() math32.Vector3 {
	return math32.ZeroVec3
}

/************************** AOI ***************************/

// OnEnterAOI other enter my view
func (e *Entity) OnEnterAOI(self *aoi.Item, other *aoi.Item) {
	otherEntity := other.Data.(*Entity)
	e.InterestedIn.Add(otherEntity)
	otherEntity.InterestedBy.Add(e)

	e.I.OnEnterSight(otherEntity)
}

// OnLeaveAOI other leave my view
func (e *Entity) OnLeaveAOI(self *aoi.Item, other *aoi.Item) {
	otherEntity := other.Data.(*Entity)
	e.InterestedIn.Del(otherEntity)
	otherEntity.InterestedBy.Del(e)

	e.I.OnLeaveSight(otherEntity)
}

/************************ database **************************/

// Save to db
func (e *Entity) Save() {
	if !e.IsPersistent() {
		return
	}

	storage.Write(common.EntityTableName(e.typName), "id", e.ID, e.Data)
	// req := dbmgr.QueryPara{
	// 	TblName:   common.EntityTableName(e.typName),
	// 	FieldName: "id",
	// 	FieldStr:  e.ID,
	// }
	// dbmgr.Update(req, e.Data)
}

// IsPersistent 实体是否需要持久化
func (e *Entity) IsPersistent() bool {
	return e.typeDesc.IsPersistent
}

/****************************** Destroy ******************************/

// Destroy destroy entity
func (e *Entity) Destroy() {
	if e.destroyed {
		return
	}

	logger.Log.Debugf("%s destroy...", e)

	e.destroyEntity(false)

	// TODO 通知 实体管理器有实体销毁了
}

// TODO 如果涉及到迁移实体的操作，这里还没做
func (e *Entity) destroyEntity(isMigrate bool) {
	if e.Space != nil {
		e.Space.leave(e)
	}

	// TODO 根据迁移与否做响应逻辑
	if !isMigrate {
		e.I.OnDestroy()
	} else {

	}

	e.clearRawTimers()
	e.rawTimers = nil

	if !isMigrate {

	}

	e.destroyed = true

	// TODO 这里销毁时保存数据到服务器，有问题再改
	e.Save()
	entityManager.del(e)
}

func (e *Entity) clearRawTimers() {
	for t := range e.rawTimers {
		t.Stop()
	}
	e.rawTimers = map[*timer.Timer]struct{}{}
}

// IsDestroyed 实体是否已销毁
func (e *Entity) IsDestroyed() bool {
	return e.destroyed
}

/*************************** 切换/进入场景 *****************************/

// GetMigrateData 迁移逻辑
func (e *Entity) GetMigrateData() *MigrateData {
	databytes, _ := json.Marshal(e.Data)
	return &MigrateData{
		ID: e.ID,
		// 实体typName
		TypName: e.typName,
		// 实体序列化后的数据，目前使用json
		DataBytes: databytes,
		// 实体位置
		Pos: e.pos,
		// 实体朝向
		Yaw: e.yaw,
		// 实体所在space
		SpaceID: e.Space.SpaceID(),
	}
}

// EnterSpace 实体进入场景
func (e *Entity) EnterSpace(spaceID string, pos math32.Vector3) {
	if e.isEnteringSpace() {
		logger.Log.Warnf("%s is entering space %s, cannot enter space %s", e, e.enteringSpaceRequest.SpaceID, spaceID)
		e.I.OnEnterSpace()
		return
	}

	localSpace := spaceManager.getSpace(spaceID)
	if localSpace != nil {
		e.enterLocalSpace(localSpace, pos)
	} else {
		e.requestMigrateTo(spaceID, pos)
	}
}

// 进入当前进程服务中的场景
func (e *Entity) enterLocalSpace(space *Space, pos math32.Vector3) {
	if space == e.Space {
		logger.Log.Errorf("%s already in space %s", e, space)
		return
	}

	e.enteringSpaceRequest.SpaceID = space.ID
	e.enteringSpaceRequest.EnterPos = pos
	e.enteringSpaceRequest.RequestTime = time.Now().UnixNano()

	// TODO 这里考虑异步处理
	e.cancelEnterSpace()
	if space.IsDestroyed() {
		logger.Log.Warnf("entity enter space, but space is destroyed, enter space fail", zap.String("entity", e.String()), zap.String("space", space.String()))
		return
	}

	// 离开原来的场景
	if e.Space != nil {
		e.Space.leave(e)
	}
	// 进入现在的场景
	space.enter(e, pos)
}

func (e *Entity) cancelEnterSpace() {
	e.enteringSpaceRequest.SpaceID = ""
	e.enteringSpaceRequest.EnterPos = math32.Vector3{}
	e.enteringSpaceRequest.RequestTime = 0

	if e.enteringSpaceRequest.migrateRequestIsSent {

		e.enteringSpaceRequest.migrateRequestIsSent = false
		// TODO rpc通知取消迁移
	}
}

func (e *Entity) isEnteringSpace() bool {
	now := time.Now().UnixNano()
	return now < (e.enteringSpaceRequest.RequestTime + int64(EnterSpaceRequestTimeout))
}

func (e *Entity) requestMigrateTo(spaceID string, pos math32.Vector3) {
	e.enteringSpaceRequest.SpaceID = spaceID
	e.enteringSpaceRequest.EnterPos = pos
	e.enteringSpaceRequest.RequestTime = time.Now().UnixNano()

	// TODO rpc 请求迁移实体到新空间
}

/***************************** timer *****************************/

// AddCallback 添加timer callback
func (e *Entity) AddCallback(callback func(), interval time.Duration, count int) *timer.Timer {
	t := timer.NewTimer(callback, interval, count)
	e.rawTimers[t] = struct{}{}

	return t
}

// RemoveCallback remove timer callback
func (e *Entity) RemoveCallback(t *timer.Timer) {
	t.Stop()
	delete(e.rawTimers, t)
}

/********************************* move ****************************/

// func (e *Entity) move(pos math32.Vector3) {

// }

// func (e *Entity) stopMove(pos math32.Vector3) {

// }

/**************************** push to client *************************/

// PushOwnClient 给当前entity的客户端推送消息
func (e *Entity) PushOwnClient(router string, msg interface{}) {
	// helper.PushToUser(router, msg, e.ID)
}

// PushNeighbourClient 给周围玩家推送消息
func (e *Entity) PushNeighbourClient(router string, msg interface{}) {
	// e.InterestedBy.ForEach(func(e *Entity) {
	// 	if e.typName == define.TypNamePlayer {
	// 		helper.PushToUser(router, msg, e.ID)
	// 	}
	// })
}
