package cellpart

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"
	"unsafe"

	"github.com/golang/protobuf/proto"

	"github.com/tutumagi/pitaya/engine/aoi"
	"github.com/tutumagi/pitaya/engine/bc/internal/consts"
	"github.com/tutumagi/pitaya/engine/bc/metapart"
	"github.com/tutumagi/pitaya/engine/components/app"
	"github.com/tutumagi/pitaya/engine/math32"
	err "github.com/tutumagi/pitaya/errors"
	"github.com/tutumagi/pitaya/logger"
	"github.com/tutumagi/pitaya/protos"
	"github.com/tutumagi/pitaya/timer"
	"gitlab.gamesword.com/nut/entitygen/attr"
	"go.uber.org/zap"
)

type Entity struct {
	UID string // UID 只有在是玩家连接的实体，才有UID，消息收发是靠UID来标识的
	ID  string // 实体ID，如果是玩家实体，则ID也表示角色ID

	data *attr.StrMap

	v        interface{}
	I        ICellEntity
	typeDesc *metapart.TypeDesc

	destroyed     bool
	destroyReason int32

	baseServerID string

	space *Space // 在 cell server 里面使用此 space

	yaw float32

	Viewlayer metapart.ViewLayer // 视野类型
	coord     *aoi.EntityCoord
	witness   *aoi.Witness

	rawTimers map[int64]*timer.Timer

	enteringSpaceRequest *protos.EnterSpaceReq
}

func (e *Entity) resetEnterSpaceRequest() {
	e.enteringSpaceRequest.Reset()
	e.enteringSpaceRequest.FromServerID = curServerID()
	if e.enteringSpaceRequest.Pos == nil {
		e.enteringSpaceRequest.Pos = &protos.Vec3{X: 0, Y: 0, Z: 0}
	} else {
		e.enteringSpaceRequest.Pos.X = 0
		e.enteringSpaceRequest.Pos.Y = 0
		e.enteringSpaceRequest.Pos.Z = 0
	}
}

func (e *Entity) Data() *attr.StrMap {
	return e.data
}

// Destroy destroy entity
func (e *Entity) Destroy(reason ...int32) {
	if e.destroyed {
		return
	}

	if e.typeDesc.TypName() == metapart.TypNamePlayer {
		logger.Debugf("%s destroy...%v", e, reason)
	}

	if len(reason) > 0 {
		e.destroyReason = reason[0]
	}
	e.destroyEntity(false)

	// TODO 告诉baseapp 实体的cellpart销毁了
	// // 如果是在 base server destroy，通知对应的space 玩家离开了
	// if e.SpaceCreated() {
	// 	if erre := app.SendTo(
	// 		context.Background(),
	// 		e.SpaceServerID(),
	// 		"cellremote.destroyentity",
	// 		&protos.Entity{
	// 			Id:     e.ID,
	// 			Label:  e.TypName(),
	// 			Reason: e.destroyReason,
	// 		},
	// 	); erre != nil {
	// 		logger.Error("leave rpc error", zap.Error(erre))
	// 	}
	// }

}

func (e *Entity) Space() *Space {
	return e.space
}

func (e *Entity) destroyEntity(isMigrate bool) {
	if e.space != nil {
		e.space.leave(e)
	}

	if isMigrate {
		e.I.OnMigrateOut()
	} else {
		e.I.OnDestroy()
	}

	e.destroyed = true

	cellEntManager.del(e)
}

// BaseServerID 返回baseappId
func (e *Entity) BaseServerID() string {
	return e.baseServerID
}

func (e *Entity) SetBaseServerID(id string) {
	e.baseServerID = id
}

func (e *Entity) isEnteringSpace() bool {
	now := time.Now().UnixNano()
	return now < (e.enteringSpaceRequest.Time + int64(consts.EnterSpaceRequestTimeout))
}

// PrepareEnterSpace 实体准备进入场景
func (e *Entity) PrepareEnterSpace(spaceID string, spaceKind int32, pos math32.Vector3, viewLayer int32) {
	if e.isEnteringSpace() {
		logger.Warnf("%s is entering space %s, cannot enter space %s", e, e.enteringSpaceRequest.SpaceID, spaceID)
		// e.I.OnEnterSpace()
		e.PushEnterSceneErrorIfNeed(metapart.ErrEnteringSpace)
		return
	}
	// 这个不会触发 positionChanged的回调
	e.coord.SetPosition(pos.X, pos.Y, pos.Z)
	// 在 cell server 上，先找下本地有没有该 space，有则直接进入，否则请求 cellmgrapp 去拿 space server id

	localSpace := spaceManager.getSpace(spaceID)
	if localSpace != nil {
		e.enterLocalSpace(localSpace, pos, viewLayer)
	} else {
		// TODO 这个分支还没测试到(多个cell的情况才有可能走到这里)
		e.requestMigrateTo(spaceID, spaceKind, pos, viewLayer)
	}

}

func (e *Entity) GetPosition() *math32.Vector3 {
	return e.coord.Position()
}

// SpaceKind 返回实体所在的spaceKind
func (e *Entity) SpaceKind() int32 {
	if e.space == nil {
		return metapart.NilSpaceKind
	}

	return e.space.kind
}

func (e *Entity) requestMigrateTo(spaceID string, spaceKind int32, pos math32.Vector3, viewLayer int32) {
	e.enteringSpaceRequest.EntityID = e.ID
	e.enteringSpaceRequest.EntityLabel = e.typeDesc.TypName()
	e.enteringSpaceRequest.SpaceID = spaceID
	e.enteringSpaceRequest.SpaceKind = spaceKind
	e.enteringSpaceRequest.Pos.X = pos.X
	e.enteringSpaceRequest.Pos.Y = pos.Y
	e.enteringSpaceRequest.Pos.Z = pos.Z
	e.enteringSpaceRequest.Time = time.Now().UnixNano()
	e.enteringSpaceRequest.ViewLayer = viewLayer

	// 请求进入场景
	err := app.Send(context.Background(), "", "", "cellmgrapp.spaceservice.enterspace", e.enteringSpaceRequest)
	if err != nil {
		logger.Errorf("%s enter space(%s) err(%s)", e, spaceID, err)
		e.PushEnterSceneErrorIfNeed(metapart.ErrSpaceRequestFailed)
	}
}

// 进入当前进程服务中的场景
func (e *Entity) enterLocalSpace(space *Space, pos math32.Vector3, viewLayer int32) {
	logger.Debugf("e.baseId=%s, space.baseId=%s ", e.BaseServerID(), space.BaseServerID())
	if space == e.space {
		logger.Errorf("%s already in space %s", e, space)
		if e.BaseServerID() != "" && e.BaseServerID() != space.BaseServerID() {
			logger.Warnf("%s already in space %s, but 。。。", e, space)
		} else {
			e.PushEnterSceneErrorIfNeed(metapart.ErrAlreadyInSpace)
			return
		}
	}

	e.enteringSpaceRequest.SpaceID = space.ID
	e.enteringSpaceRequest.SpaceKind = space.kind
	e.enteringSpaceRequest.Pos.X = pos.X
	e.enteringSpaceRequest.Pos.Y = pos.Y
	e.enteringSpaceRequest.Pos.Z = pos.Z
	e.enteringSpaceRequest.Time = time.Now().UnixNano()
	e.enteringSpaceRequest.ViewLayer = viewLayer

	// 这里是将之前的进入场景的消息给取消掉
	e.cancelEnterSpace()
	if space.IsDestroyed() {
		logger.Warnf("entity enter space, but space is destroyed, enter space fail", zap.String("entity", e.String()), zap.String("space", space.String()))
		e.PushEnterSceneErrorIfNeed(metapart.ErrSpaceDestroyed)
		return
	}

	// 离开原来的场景
	if e.space != nil {
		e.space.leave(e)
	}

	// 这里把第一次进场景的视野参数给保存住
	e.Viewlayer = viewLayer
	// 进入现在的场景
	space.enter(e, pos)
}

// IsDestroyed 实体是否已销毁
func (e *Entity) IsDestroyed() bool {
	return e.destroyed
}

func (e *Entity) PushEnterSceneErrorIfNeed(eee error) {
	if e.UID != "" && eee != nil {
		logger.Errorf("push enter space err. entity:%s err:%s", e, eee)
		retErr, ok := eee.(*err.Error)
		if !ok {
			retErr = metapart.ErrSpaceUnknown(eee)
		}
		e.PushOwnClient(metapart.RouterAvatarEnterScene, &protos.RspEnter{
			Code: retErr.Code,
			Msg:  retErr.Message,
		})

		if e.enteringSpaceRequest != nil {
			e.enteringSpaceRequest.Time = 0
		}
	}
}

func (e *Entity) cancelEnterSpace() {
	requestIsSend := e.enteringSpaceRequest.IsSend
	e.resetEnterSpaceRequest()

	if requestIsSend {
		// TODO rpc通知取消迁移
	}
}

// MigrateToSpaceFromCell 迁移到另一个space，这个只会在拿到 space server id 后调用到
func (e *Entity) MigrateToSpaceFromCell(spaceID string, spaceKind int32, pos math32.Vector3, viewLayer int32) {
	localSpace := spaceManager.getSpace(spaceID)
	if localSpace != nil {
		// 如果本地服有该场景，则直接进入该场景
		e.enterLocalSpace(localSpace, pos, viewLayer)
	} else {
		logger.Warnf("这里应该能拿到此场景，但是没有拿到",
			zap.String("spaceID", spaceID),
			zap.Int32("spaceKind", spaceKind),
			zap.Any("spaces", spaceManager.spaces))
		// TODO 这里需要通知该实体 进入场景失败
		e.PushEnterSceneErrorIfNeed(metapart.ErrFindSpace(spaceID))
	}
}

// PushOwnClient 给当前entity的客户端推送消息
func (e *Entity) PushOwnClient(router string, msg interface{}) {
	// helper.PushToUser(router, msg, e.ID)
	if e.UID == "" {
		return
	}
	_, err := app.SendPushToUsers(router, msg, []string{e.UID}, metapart.GateAppSvr)
	if err != nil {
		logger.Errorf("push own client(uid:%s) err:%s", e.UID, err)
	}
}

// PushNeighbourClient 给周围玩家推送消息
func (e *Entity) PushNeighbourClient(router string, msg interface{}) {
	if router == "entity.stopsyncpos" || router == "entity.syncpos" || router == "entity.syncmoveto" {

	} else {
		logger.Debugf("aoi %s %v", router, msg.(proto.Message).String())
	}

	if e.witness == nil {
		return
	}
	for entityer := range e.witness.InterestedBy {
		other := entityer.(*Entity)
		if other.typeDesc.TypName() == metapart.TypNamePlayer {
			if other.UID == "" {
				logger.Warn("typename is player, but uid is empty", zap.String("entity", other.String()))
				return
			}
			_, err := app.SendPushToUsers(router, msg, []string{other.UID}, metapart.GateAppSvr)
			if err != nil {
				logger.Errorf("push neighbour client(uid:%s) err:%s", other.UID, err)
			}
		}
	}
}

// AoiID string return aoi id
func (e *Entity) AoiID() string {
	return e.ID
}

// Coord get aoi coord
func (e *Entity) Coord() *aoi.EntityCoord {
	return e.coord
}

func (e *Entity) setWitness(w *aoi.Witness) {
	e.witness = w
	w.Attach(e)
	w.SetViewRadius(e.typeDesc.AOIDistance(), 0)
}

// Witness get witness
func (e *Entity) Witness() *aoi.Witness {
	return e.witness
}

// InterestedBy return interestedBy entities
func (e *Entity) InterestedBy() map[aoi.Entityer]struct{} {
	if e.witness != nil {
		return e.witness.InterestedBy
	}
	return nil
}

// InterestIn return InterestIn entities
func (e *Entity) InterestIn() map[aoi.Entityer]struct{} {
	if e.witness != nil {
		return e.witness.InterestIn
	}
	return nil
}

// OnEnterAOI other enter my view
func (e *Entity) OnEnterAOI(other aoi.Entityer) {
	otherEntity := other.(*Entity)
	e.I.OnEnterSight(otherEntity)
}

// OnLeaveAOI other leave my view
func (e *Entity) OnLeaveAOI(other aoi.Entityer) {
	otherEntity := other.(*Entity)
	e.I.OnLeaveSight(otherEntity)
}

// // ma 变更的key所属的map
// //	key 变化的key
// //	val 变化的值
// func (e *CellEntity) sendMapAttrChangeToClients(ma *AttrMap, key string, val interface{}) {
// 	var flag AttrFlag
// 	if ma == e.data {
// 		flag = e.getAttrFlag(key)
// 	} else {
// 		// flag = ma.def.flag
// 	}

// 	if svrConfig.curServerUseSpace {
// 		// 是cellapp
// 		// TODO 需要通知周围的其他客户端
// 		if flag&afOtherClient > 0 {

// 		}
// 		// TODO 需要通知自己的客户端
// 		if flag&afClient > 0 {

// 		}
// 	} else {
// 		// 不是cellapp
// 		// TODO 需要通知到cellapp
// 		if flag&afCell > 0 {

// 		}
// 		// TODO 需要通知给周围的cellapp
// 		if flag&afOtherCell > 0 {

// 		}
// 	}
// }

// func (e *CellEntity) getAttrFlag(key string) AttrFlag {
// 	return e.typeDesc.attrsDef[key].flag
// }

// PushInterestin 推送 该实体 关心的实体列表消息
func (e *Entity) PushInterestin(router string, msg interface{}) {
	logger.Debugf("aoi Interestin %s %v", router, msg.(proto.Message).String())

	if e.witness == nil {
		return
	}
	for entityer := range e.witness.InterestIn {
		other := entityer.(*Entity)
		if other.typeDesc.TypName() == metapart.TypNamePlayer {
			if other.UID == "" {
				logger.Warn("typename is player, but uid is empty", zap.String("entity", other.String()))
				return
			}
			app.SendPushToUsers(router, msg, []string{other.UID}, metapart.GateAppSvr)
		}
	}
}

func (e *Entity) String() string {
	if e == nil {
		return ""
	}
	// return fmt.Sprintf("<Entity>label:%s id:%s data:%+v", e.typName, e.ID, e.Data)
	name := "no typedesc"
	if e.typeDesc != nil {
		name = e.typeDesc.TypName()
	}
	return fmt.Sprintf("<Entity>label:%s id:%s pos:%s", name, e.ID, e.coord.Position())
}

func (e *Entity) init(id string, typName string, entityInstance reflect.Value) {
	e.ID = id
	e.v = entityInstance.Interface()
	e.I = entityInstance.Interface().(ICellEntity)

	e.typeDesc = metapart.GetTypeDesc(typName)

	e.rawTimers = make(map[int64]*timer.Timer)

	e.enteringSpaceRequest = &protos.EnterSpaceReq{}
	e.resetEnterSpaceRequest()

	// if e.typeDesc.aoiFlag.Valid() && svrConfig.curServerUseSpace {
	// if svrConfig.curServerUseSpace {
	// e.coord = math32.NewVector3(0, 0, 0)
	e.coord = aoi.NewEntityNode(e)
	// }

}

// IsUseAOI 该实体是否开启AOI
func (e *Entity) IsUseAOI() bool {
	return e.typeDesc.UseAOI()
}

// TypName 该实体的类别
func (e *Entity) TypName() string {
	return e.typeDesc.TypName()
}

// IsSpaceEntity 此实体是否是空间
func (e *Entity) IsSpaceEntity() bool {
	return strings.Contains(e.typeDesc.TypName(), consts.SpaceEntityType)
}

// AsSpace 类型转换为space
func (e *Entity) AsSpace() *Space {
	if !e.IsSpaceEntity() {
		logger.Panicf("%s is not a space", e)
	}

	return (*Space)(unsafe.Pointer(e))
}

// Val 返回实际用户定义的 实体interface
//	 使用方式：landEntity := e.Val().(*LandEntity)
func (e *Entity) Val() interface{} {
	return e.v
}

func (e *Entity) DestroyReason() int32 {
	return e.destroyReason
}

// OnInit IEntity 的接口，构造完成后回调
func (e *Entity) OnInit() error {
	return nil
}

// OnCreated IEntity 的接口，构造完成，初始化数据后会回调
func (e *Entity) OnCreated() {}

// OnDestroy IEntity的接口，销毁时回调
func (e *Entity) OnDestroy() {}

// OnMigrateOut 实体迁移出去后
func (e *Entity) OnMigrateOut() {}

// OnMigrateIn 实体迁移进来后
func (e *Entity) OnMigrateIn() {}

// EnterSpaceFailed 实体进入场景失败了 在业务 app 触发
func (e *Entity) EnterSpaceFailed(error) {}

// EnterSpaceFailed 实体进入场景失败了 在业务 app 触发
func (e *Entity) EnterSpaceSuccess(spaceID string, spaceKind int32) {}

// BeforeEnterSpace IEntity的接口，进入场景前回调
func (e *Entity) BeforeEnterSpace() {}

// AfterEnterSpace IEntity的接口，进入场景后回调
func (e *Entity) AfterEnterSpace() {}

// OnLeaveSpace IEntity的接口，离开场景后回调
func (e *Entity) OnLeaveSpace(s *Space) {}

// OnEnterSight IEntity的接口，其他实体进入视野时回调
func (e *Entity) OnEnterSight(other *Entity) {}

// OnLeaveSight IEntity的接口，其他实体离开视野时回调
func (e *Entity) OnLeaveSight(other *Entity) {}

// DefaultPos 默认的位置，不会触发在 space中的 move api。 只会在加载/创建实体时，数据创建完成后调用
func (e *Entity) DefaultPos() math32.Vector3 {
	return math32.ZeroVec3
}

// Tick 所有实体的主循环
func (e *Entity) Tick(dt int32) {}

// OnPositionYawChanged 实体的位置和朝向发生了变化
func (e *Entity) OnPositionYawChanged(newPos *math32.Vector3, yaw float32) {}

// // OnPositionChanged 实体的位置发生了变化
// func (e *CellEntity) OnPositionChanged(newPos *math32.Vector3) {}

// // OnYawChanged 实体的朝向发生了变化
// func (e *CellEntity) OnYawChanged(yaw float32) {}

// OnStopMove 实体停止移动了
func (e *Entity) OnStopMove(newPos *math32.Vector3, yaw float32) {}

// StopMove 停止移动
func (e *Entity) StopMove(pos math32.Vector3, yaw float32) {
	// vec3 := vec3pool.Get().(*math32.Vector3)
	// vec3.X = pos.X
	// vec3.Y = pos.Y
	// vec3.Z = pos.Z

	// e.setPosition(vec3, yaw)

	// e.I.OnStopMove(vec3, yaw)

	// vec3pool.Put(vec3)

	e.setPosition(&pos, yaw)
	e.I.OnStopMove(&pos, yaw)
}

// Move set pos
func (e *Entity) Move(pos math32.Vector3) {
	// vec3 := vec3pool.Get().(*math32.Vector3)
	// vec3.X = pos.X
	// vec3.Y = pos.Y
	// vec3.Z = pos.Z

	// e.setPosition(vec3, e.yaw)

	// e.I.OnPositionChanged(vec3)
	// vec3pool.Put(vec3)

	e.setPosition(&pos, e.yaw)
	e.I.OnPositionYawChanged(&pos, e.yaw)
}

// Rot 设置朝向,yaw 绕Y轴旋转弧度
func (e *Entity) Rot(yaw float32) {
	// e.setPosition(nil, yaw)

	// e.I.OnYawChanged(yaw)

	e.setPosition(nil, yaw)
	e.I.OnPositionYawChanged(e.coord.Position(), yaw)
}

// MoveAndRot set pos & yaw
func (e *Entity) MoveAndRot(pos math32.Vector3, yaw float32) {
	e.setPosition(&pos, yaw)
	e.I.OnPositionYawChanged(&pos, yaw)
}

func (e *Entity) setPosition(pos *math32.Vector3, yaw float32) {
	space := e.space
	if space == nil {
		logger.Errorf("%s setPosition %s. space is nil", e, pos)
		return
	}
	if pos != nil {
		space.move(e, pos)
	} else {
		// pos = e.coord.Position()
	}
	e.yaw = yaw
}

// GetYaw get yaw
func (e *Entity) GetYaw() float32 {
	return e.yaw
}

// SpaceID 返回实体所在的spaceID
func (e *Entity) SpaceID() string {
	if e.space == nil {
		return ""
	}
	return e.space.SpaceID()
}

func (e *Entity) LeaveSpace() error {
	if e.isEnteringSpace() {
		logger.Warnf("%s is entering space %s, prepare cancel enter %s", e, e.enteringSpaceRequest.SpaceID, e.SpaceID())

		return metapart.ErrEnteringSpace
	}

	if e.space != nil {
		e.space.leave(e)
		e.Destroy(int32(protos.EntityLeaveReason_ELR_LEAVESPACE))
	}
	return nil
}

// Save to db
// Save 表示有数据变化，
func (e *Entity) Save() {
	// // TODO 考虑多并发的情况
	// if !e.data.HasChange() {
	// 	return
	// }
	// 先进行数据变化key 的处理
	e.attrChanged()
}

func (e *Entity) attrChanged() {
	if !e.data.HasChange() {
		return
	}
	logger.Debugf("attr change label:%s id:%s %v", e.TypName(), e.ID, e.data.ChangeKey())
	var cellkeys map[string]struct{} = map[string]struct{}{}
	// var basekeys []string = []string{}
	for k := range e.data.ChangeKey() {
		f := e.typeDesc.GetDef(k)
		if f.HasCell() {
			cellkeys[k] = struct{}{}
		}
		// if f.HasCell() {
		// 	basekeys = append(basekeys, k)
		// }
	}

	if len(cellkeys) > 0 {
		e.I.CellAttrChanged(cellkeys)
	}

	// 处理完 数据变化key回调，和db操作后，清理掉变化key
	e.data.ClearChangeKey()
}

// AttrChangedFromBytes 只会在 cellapp 调用
func (e *Entity) AttrChangedFromBytes(attrBytes map[string][]byte) {
	if len(attrBytes) == 0 {
		return
	}
	e.unmarshalChangedKey(attrBytes)
	e.attrChanged()
}

// UnmarshalChangedKey 为了 bc package 的测试代码，后面要改为私有
func (e *Entity) unmarshalChangedKey(changeAttrBytes map[string][]byte) {
	for k, bb := range changeAttrBytes {
		def := e.typeDesc.GetDef(k)
		if def == nil {
			logger.Warn("cell attr change but not have the key def continue",
				zap.String("entity", e.String()),
				zap.String("key", k),
			)
			continue
		}
		if def.IsPrimary() {
			newV, erro := def.UnmarshalPrimary(bb)
			if erro != nil {
				logger.Warn("cell attr change unmarshal primary key error",
					zap.String("entity", e.String()),
					zap.String("key", k),
					zap.Any("val", newV),
					zap.Error(erro),
				)
				continue
			}
			e.data.Set(k, newV)
		} else {
			// TODO 这里每次都会分配内存，会有 gc 问题，这里考虑看怎么处理，能否可以跟上面的条件分支一样。
			// 使用 e.data.attrs 里面已经有 的值来 Unmarshal
			var a = reflect.New(def.P()).Interface()
			// var a = e.data.Value(k)
			erro := json.Unmarshal(bb, &a)
			if erro != nil {
				logger.Warn("cell attr change unmarshal not primary key error",
					zap.String("entity", e.String()),
					zap.String("key", k),
					// zap.Any("val", val),
					zap.Error(erro),
				)
				continue
			}
			val := reflect.ValueOf(a).Elem().Interface()
			e.data.Set(k, val)
		}

	}
}

// AddCallback 添加timer callback
func (e *Entity) AddCallback(callback func(), interval time.Duration, count int) *timer.Timer {
	if e.rawTimers == nil {
		return nil
	}
	e.RemoveOverTimerIfHas()

	t := timer.NewTimer(callback, interval, count)
	e.rawTimers[t.ID] = t
	// logger.Infof("AddCallback. rawTimers.size = %d, uid = %s, timer.id = %d", len(e.rawTimers), e.UID, t.ID)

	return t
}

// RemoveCallback remove timer callback
func (e *Entity) RemoveCallback(t *timer.Timer) {
	t.Stop()
	delete(e.rawTimers, t.ID)
}

// 清除已完成（closed == 1）的Timer
func (e *Entity) RemoveOverTimerIfHas() {
	if e.rawTimers == nil {
		return
	}

	tmpTimers := map[int64]*timer.Timer{}
	for _, t := range e.rawTimers {
		if t.IsClose() {
			//remove it
			// logger.Infof("remove closed-timer for uid = %s, timer.id = %d", e.UID, t.ID)
		} else {
			tmpTimers[t.ID] = t
		}
	}
	e.rawTimers = tmpTimers
}

// TODO 为了避免循环引用，后面要考虑怎么弄这个
func curServerID() string {
	// return app.GetServerID()
	return ""
}
