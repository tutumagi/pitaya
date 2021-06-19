package basepart

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"
	"unsafe"

	"github.com/AsynkronIT/protoactor-go/actor"

	"github.com/tutumagi/pitaya/agent"
	"github.com/tutumagi/pitaya/engine/bc/metapart"
	"github.com/tutumagi/pitaya/engine/common"
	"github.com/tutumagi/pitaya/engine/components/app"
	"github.com/tutumagi/pitaya/engine/dbmgr"
	"github.com/tutumagi/pitaya/engine/math32"
	err "github.com/tutumagi/pitaya/errors"
	"github.com/tutumagi/pitaya/logger"
	"github.com/tutumagi/pitaya/protos"
	"github.com/tutumagi/pitaya/timer"
	"gitlab.gamesword.com/nut/dreamcity/game/define"
	"gitlab.gamesword.com/nut/entitygen/attr"
	"go.uber.org/zap"
)

// 在 base server 里面的 space 信息
type _BaseSpaceInfo struct {
	ID       string
	Kind     int32
	ServerID string
}

func newBaseSpaceInfo(id string, kind int32, serverID string) *_BaseSpaceInfo {
	return &_BaseSpaceInfo{
		ID:       id,
		Kind:     kind,
		ServerID: serverID,
	}
}

// Entity 实体
type Entity struct {
	UID string // UID 只有在是玩家连接的实体，才有UID，消息收发是靠UID来标识的
	ID  string // 实体ID，如果是玩家实体，则ID也表示角色ID
	// Data interface{}

	data *attr.StrMap

	v        interface{}
	I        IBaseEntity
	typeDesc *metapart.TypeDesc

	destroyed     bool
	destroyReason int32

	baseSpace *_BaseSpaceInfo // 在 base server 里面用此 space info

	coord *math32.Vector3
	yaw   float32

	Viewlayer metapart.ViewLayer // 视野类型

	rawTimers map[int64]*timer.Timer

	enteringSpaceRequest *protos.EnterSpaceReq

	pid *actor.PID

	client *agent.Remote
}

// NewEntity ctor
// func NewEntity(id string, data interface{}, pos math32.Vector3) *BaseEntity {
// 	return &BaseEntity{
// 		ID:           id,
// 		Data:         data,
// 		pos:          pos,
// 		InterestedIn: Set{},
// 		InterestedBy: Set{},
// 	}
// }

func (e *Entity) String() string {
	if e == nil {
		return ""
	}
	// return fmt.Sprintf("<BaseEntity>label:%s id:%s data:%+v", e.typName, e.ID, e.Data)
	name := "no typedesc"
	if e.typeDesc != nil {
		name = e.TypName()
	}
	return fmt.Sprintf("<BaseEntity>label:%s id:%s pos:%s", name, e.ID, e.coord)
}

// GiveClientTo gives Client to other entity
func (e *Entity) GiveClientTo(other *Entity) {
	if e.client == nil {
		logger.Warnf("%s.GiveClientTo(%s): Client is nil", e, other)
		return
	}

	logger.Debugf("%s.GiveClientTo(%s): Client=%s", e, other, e.client)

	client := e.client

	// client.ownerid = other.ID // hack ownerid so that destroy entity messages will be synced with create entity messages
	e.SetClient(nil)
	other.SetClient(client)
}

// abouc client
func (e *Entity) SetClient(client *agent.Remote) {
	oldClient := e.client
	if oldClient == client {
		return
	}
	if oldClient != nil {
		// send destroy entity to Client
		// dispatchercluster.SelectByEntityID(e.ID).SendClearClientFilterProp(oldClient.gateid, oldClient.clientid)

		// for neighbor := range e.InterestedBy {
		// 	oldClient.sendDestroyEntity(neighbor)
		// }

		// if !e.Space.IsNil() {
		// 	oldClient.sendDestroyEntity(&e.Space.Entity)
		// }

		// oldClient.sendDestroyEntity(e)
	}
	e.assignClient(client)
	if client != nil {
		// send create entity to new client
		// dispatchercluster.SelectByEntityID(e.ID).SendClearClientFilterProp(client.gateid, client.clientid)
		// client.sendCreateEntity(e, true)

		// if !e.Space.IsNil() {
		// 	client.sendCreateEntity(&e.Space.Entity, false)
		// }

		// for neighbor := range e.InterestedBy {
		// 	client.sendCreateEntity(neighbor, false)
		// }
	}

	// 如果是从有client到没有client，则调用连接断开
	if oldClient != nil && client == nil {
		baseEntManager.system.Root.Send(e.pid, clientDisconnectMsg)
		// e.I.OnClientDisconnected()
	} else if client != nil { // 如果是连接连上
		baseEntManager.system.Root.Send(e.pid, clientConnectMsg)
		// e.I.OnClientConnected()
	}
}

func (e *Entity) assignClient(client *agent.Remote) {
	if e.client != nil {
		// e.client.ownerid = ""
	}
	e.client = client
	if client != nil {
		// e.client.ownerid = e.ID
	}
}

func (e *Entity) init(id string, typName string, entityInstance reflect.Value) {
	e.ID = id
	e.v = entityInstance.Interface()
	e.I = entityInstance.Interface().(IBaseEntity)

	e.typeDesc = metapart.GetTypeDesc(typName)

	e.rawTimers = make(map[int64]*timer.Timer)

	e.enteringSpaceRequest = &protos.EnterSpaceReq{}
	e.resetEnterSpaceRequest()

	// if e.typeDesc.aoiFlag.Valid() && svrConfig.curServerUseSpace {
	// if svrConfig.curServerUseSpace {
	e.coord = math32.NewVector3(0, 0, 0)
	// }

}

func (e *Entity) cancelEnterSpace() {
	requestIsSend := e.enteringSpaceRequest.IsSend
	e.resetEnterSpaceRequest()

	if requestIsSend {
		// TODO rpc通知取消迁移
	}
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

// AttrChangedFromBytes 只会在 cellapp 调用
func (e *Entity) AttrChangedFromBytes(attrBytes map[string][]byte) {
	if len(attrBytes) == 0 {
		return
	}
	e.UnmarshalChangedKey(attrBytes)
	e.attrChanged()
}

// IsSpaceEntity 此实体是否是空间
func (e *Entity) IsSpaceEntity() bool {
	return strings.Contains(e.typeDesc.TypName(), common.SpaceEntityType)
}

// AsSpace 类型转换为space
func (e *Entity) AsSpace() *Space {
	if !e.IsSpaceEntity() {
		logger.Panicf("%s is not a space", e)
	}

	return (*Space)(unsafe.Pointer(e))
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
		if !e.SpaceCreated() {
			return
		}

		changeKeyBytes := e.MarshalChangedKey(cellkeys)
		req := &protos.EntityUpdate{
			Id:    e.ID,
			Label: e.TypName(),
			Attrs: changeKeyBytes,
		}

		// erro := app.SendTo(context.TODO(), e.ID,
		// 	e.TypName(), e.SpaceServerID(), "cellremote.updateattr", req)
		erro := e.SendServiceTo(context.TODO(), e.SpaceServerID(), "cellremote", "cellremote.updateattr", req)
		if erro != nil {
			logger.Warn("send to cellapp update attr err", zap.String("entity", e.String()), zap.Error(erro))
		}
	}

	// 处理完 数据变化key回调，和db操作后，清理掉变化key
	e.data.ClearChangeKey()
}

// TODO 为了bc package的测试代码，后面要改为私有的
func (e *Entity) MarshalChangedKey(keys map[string]struct{}) map[string][]byte {
	ret := map[string][]byte{}
	for k := range keys {
		val := e.data.Value(k)
		bb, erro := json.Marshal(val)
		if erro != nil {
			logger.Warn("cell attr change marshal key error",
				zap.String("entity", e.String()),
				zap.String("key", k),
				zap.Any("val", val),
				zap.Error(erro))
			continue
		}
		logger.Debug("cell attr change marshal key",
			zap.String("entity", e.String()),
			zap.String("key", k),
			zap.Any("val", val),
			zap.Any("bytes", bb))

		ret[k] = bb
	}

	return ret
}

// UnmarshalChangedKey 为了 bc package 的测试代码，后面要改为私有
func (e *Entity) UnmarshalChangedKey(changeAttrBytes map[string][]byte) {
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

/************************ Getter/Setter *************************/
func (e *Entity) Data() *attr.StrMap {
	return e.data
}

// TODO 理论上不应该暴露这个方法
func (e *Entity) SetBaseSpaceInfo(id string, kind int32, serverID string) {
	e.baseSpace = newBaseSpaceInfo(id, kind, serverID)
}

// SpaceServerID 返回实体所在space的serverID
func (e *Entity) SpaceServerID() string {
	return e.baseSpace.ServerID
}

// SpaceID 返回实体所在的spaceID
func (e *Entity) SpaceID() string {
	if !e.SpaceCreated() {
		return ""
	}

	return e.baseSpace.ID

}

// SpaceKind 返回实体所在的spaceKind
func (e *Entity) SpaceKind() int32 {
	if !e.SpaceCreated() {
		return define.NilSpaceKind
	}

	return e.baseSpace.Kind
}

// SpaceCreated 实体是否已经在space创建了，只有这里返回true，spaceID和spaceKind 才有合法的值
func (e *Entity) SpaceCreated() bool {
	return e.baseSpace != nil
}

// GetPosition get pos
func (e *Entity) GetPosition() *math32.Vector3 {
	return e.coord
}

// GetYaw get yaw
func (e *Entity) GetYaw() float32 {
	return e.yaw
}

func (e *Entity) NotifyCellMove(pos *math32.Vector3, yaw float32) {
	if !e.SpaceCreated() {
		return
	}
	spaceServerID := e.SpaceServerID()

	msg := &protos.SyncPos{
		Id:    e.ID,
		Label: e.TypName(),
		Cur: &protos.Vec3{
			X: pos.X,
			Y: pos.Y,
			Z: pos.Z,
		},
		Yaw: yaw,
	}
	if err := e.SendServiceTo(
		context.TODO(),
		spaceServerID,
		"spaceremote",
		"cellapp.spaceremote.syncpos",
		msg,
	); err != nil {
		logger.Errorf("e:%s syncpos rpc error %s", e.String(), err)
	}
	// if err := app.SendTo(
	// 	context.TODO(),
	// 	e.ID,
	// 	e.TypName(),
	// 	spaceServerID,
	// 	"cellapp.spaceremote.syncpos",
	// 	msg,
	// ); err != nil {
	// 	logger.Errorf("e:%s syncpos rpc error %s", e.String(), err)
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

// // IsSpaceEntity 此实体是否是空间
// func (e *BaseEntity) IsSpaceEntity() bool {
// 	return strings.Contains(e.typeDesc.name, _SpaceEntityType)
// }

// // AsSpace 类型转换为space
// func (e *BaseEntity) AsSpace() *Space {
// 	if !e.IsSpaceEntity() {
// 		logger.Panicf("%s is not a space", e)
// 	}

// 	return (*Space)(unsafe.Pointer(e))
// }

// Val 返回实际用户定义的 实体interface
//	 使用方式：landEntity := e.Val().(*LandEntity)
func (e *Entity) Val() interface{} {
	return e.v
}

func (e *Entity) DestroyReason() int32 {
	return e.destroyReason
}

/************************ IEntity interface default imp ***********************/

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
// func (e *BaseEntity) OnPositionChanged(newPos *math32.Vector3) {}

// // OnYawChanged 实体的朝向发生了变化
// func (e *BaseEntity) OnYawChanged(yaw float32) {}

// OnStopMove 实体停止移动了
func (e *Entity) OnStopMove(newPos *math32.Vector3, yaw float32) {}

// 客户端断开连接
func (e *Entity) OnClientDisconnected() {}

// 客户端连接上
func (e *Entity) OnClientConnected() {}

/************************** AOI ***************************/

// AoiID string return aoi id
func (e *Entity) AoiID() string {
	return e.ID
}

/************************ database **************************/

// Save to db
// Save 表示有数据变化，
func (e *Entity) Save() {
	// // TODO 考虑多并发的情况
	// if !e.data.HasChange() {
	// 	return
	// }
	// 先进行数据变化key 的处理
	e.attrChanged()

	// 在其他app 需要做存储db操作
	e.saveToDB()
}

func (e *Entity) saveToDB() {
	// 如果实体不用持久化，或者是在 aoi server 则return
	if !e.IsPersistent() {
		return
	}

	req := dbmgr.QueryPara{
		TblName: define.EntityTableName(e.TypName()),
		KeyName: "id",
		Key:     e.ID,
	}
	dbmgr.Set(req, e.GetPersistentData())
}

// IsPersistent 实体是否需要持久化
func (e *Entity) IsPersistent() bool {
	return e.typeDesc.IsPersistent()
}

/****************************** Destroy ******************************/

// Destroy destroy entity
func (e *Entity) Destroy(reason ...int32) {
	if e.destroyed {
		return
	}

	if e.typeDesc.TypName() == define.TypNamePlayer {
		logger.Debugf("%s destroy...%v", e, reason)
	}

	if len(reason) > 0 {
		e.destroyReason = reason[0]
	}
	e.destroyEntity(false)

	// 如果是在 base server destroy，通知对应的space 玩家离开了
	if e.SpaceCreated() {
		if erre := e.SendServiceTo(
			context.Background(),
			e.SpaceServerID(),
			"cellremote",
			"cellremote.destroyentity",
			&protos.Entity{
				Id:     e.ID,
				Label:  e.TypName(),
				Reason: e.destroyReason,
			},
		); erre != nil {
			logger.Error("leave rpc error", zap.Error(erre))
		}
		// if erre := app.SendTo(
		// 	context.Background(),
		// 	e.ID,
		// 	e.TypName(),
		// 	e.SpaceServerID(),
		// 	"cellremote.destroyentity",
		// 	&protos.Entity{
		// 		Id:     e.ID,
		// 		Label:  e.TypName(),
		// 		Reason: e.destroyReason,
		// 	},
		// ); erre != nil {
		// 	logger.Error("leave rpc error", zap.Error(erre))
		// }
	}

}

// TODO 如果涉及到迁移实体的操作，这里还没做
func (e *Entity) destroyEntity(isMigrate bool) {
	if isMigrate {
		e.I.OnMigrateOut()
	} else {
		e.I.OnDestroy()
	}

	e.clearRawTimers()
	e.rawTimers = nil

	if isMigrate {
		// TODO 这里是否需要做一些处理？
	} else {
		// 如果不是迁移时的销毁，则保存一下数据
		// TODO 这里销毁时保存数据到服务器，有问题再改
		e.saveToDB()
	}

	e.destroyed = true

	baseEntManager.del(e)
}

func (e *Entity) clearRawTimers() {
	for _, t := range e.rawTimers {
		t.Stop()
	}
	e.rawTimers = map[int64]*timer.Timer{}
}

// IsDestroyed 实体是否已销毁
func (e *Entity) IsDestroyed() bool {
	return e.destroyed
}

func (e *Entity) GetPersistentData() map[string]interface{} {
	return e.data.FilterMap(func(k string) bool {
		if def := e.typeDesc.GetDef(k); def != nil {
			return def.StoreDB()
		}
		return false
	})
}

func (e *Entity) GetMigrateData() map[string]interface{} {
	// return e.data.ToMap(func(k string) bool {
	// 	return true
	// })
	return e.data.ToMap()
}

func (e *Entity) getCellData() map[string]interface{} {
	var cellData = map[string]interface{}{}
	e.data.ForEach(func(k string, v interface{}) bool {
		// 应该是只发送 cell 数据给 aoi，这里先自测完成后再打开
		// def := e.typeDesc.meta.GetDef(k)
		// if def.HasCell() {
		cellData[k] = v
		// }
		return true
	})
	return cellData
}

/*************************** 切换/进入场景 *****************************/

// // GetMigratePBData 迁移逻辑
// func (e *BaseEntity) GetMigratePBData() *protos.MigrateEntityData {
// 	databytes, _ := json.Marshal(e.getMigrateData())
// 	// pos := e.coord.Position()
// 	return &protos.MigrateEntityData{
// 		UserID:   e.UID,
// 		EntityID: e.ID,
// 		// 实体typName
// 		EntityLabel: e.typeDesc.name,
// 		// 实体序列化后的数据，目前使用json
// 		EntityDatas:  databytes,
// 		FromServerID: app.GetServerID(),
// 		// 实体位置
// 		// Pos: &protos.Vec3{X: pos.X, Y: pos.Y, Z: pos.Z},
// 		// // 实体朝向
// 		// Yaw: e.yaw,
// 		// 实体所在space
// 		SpaceID:   e.SpaceID(),
// 		SpaceKind: e.SpaceKind(),
// 	}
// }

func (e *Entity) GetCellData() *protos.SEntityData {
	databytes, _ := json.Marshal(e.getCellData())
	// pos := e.coord.Position()
	return &protos.SEntityData{
		UserID:   e.UID,
		EntityID: e.ID,
		// 实体typName
		EntityLabel: e.TypName(),
		// 实体序列化后的数据，目前使用json
		EntityDatas:  databytes,
		FromServerID: curServerID(),
		// 实体位置
		// Pos: &protos.Vec3{X: pos.X, Y: pos.Y, Z: pos.Z},
		// // 实体朝向
		// Yaw: e.yaw,
		// 实体所在space
		SpaceID:   e.SpaceID(),
		SpaceKind: e.SpaceKind(),
	}
}

// GetMigratePBData 迁移逻辑
// func (e *BaseEntity) GetSimpleMigratePBData() *protos.MigrateEntityData {
// 	// pos := e.coord.Position()
// 	return &protos.MigrateEntityData{
// 		UserID:   e.UID,
// 		EntityID: e.ID,
// 		// 实体typName
// 		EntityLabel:  e.typeDesc.name,
// 		FromServerID: app.GetServerID(),
// 		// 实体位置
// 		// Pos: &protos.Vec3{X: pos.X, Y: pos.Y, Z: pos.Z},
// 		// // 实体朝向
// 		// Yaw: e.yaw,
// 		// 实体所在space
// 		SpaceID:   e.SpaceID(),
// 		SpaceKind: e.SpaceKind(),
// 	}
// }

// PrepareEnterSpace 实体准备进入场景
func (e *Entity) PrepareEnterSpace(spaceID string, spaceKind int32, pos math32.Vector3, viewLayer int32) {
	if e.isEnteringSpace() {
		logger.Warnf("%s is entering space %s, cannot enter space %s", e, e.enteringSpaceRequest.SpaceID, spaceID)
		// e.I.OnEnterSpace()
		e.PushEnterSceneErrorIfNeed(metapart.ErrEnteringSpace)
		return
	}
	// 这个不会触发 positionChanged的回调
	e.coord.Set(pos.X, pos.Y, pos.Z)

	// 如果当前server 不是 空间相关的server
	e.requestMigrateTo(spaceID, spaceKind, pos, viewLayer)
}

func (e *Entity) LeaveSpace() error {
	if e.isEnteringSpace() {
		logger.Warnf("%s is entering space %s, prepare cancel enter %s", e, e.enteringSpaceRequest.SpaceID, e.SpaceID())

		return metapart.ErrEnteringSpace
	}
	// 如果玩家不在任何一个场景里面，则返回 nil
	if !e.SpaceCreated() {
		return nil
	}

	// 如果当前server 不是 空间相关的server
	// 请求离开场景
	// err := app.RPCTo(
	// 	context.Background(),
	// 	e.ID,
	// 	e.TypName(),
	// 	e.SpaceServerID(),
	// 	"cellapp.cellremote.leavespace",
	// 	&protos.Response{},
	// 	&protos.Entity{
	// 		Id:    e.ID,
	// 		Label: e.TypName(),
	// 	})
	err := e.CallServiceTo(
		context.TODO(),
		e.SpaceServerID(),
		"cellremote",
		"cellapp.cellremote.leavespace",
		&protos.Response{},
		&protos.Entity{
			Id:    e.ID,
			Label: e.TypName(),
		})
	if err != nil {
		logger.Errorf("%s leave space(%s) err(%s)", e, e.SpaceID(), err)
		return err
	}
	e.baseSpace = nil
	return nil
}

func (e *Entity) isEnteringSpace() bool {
	now := time.Now().UnixNano()
	return now < (e.enteringSpaceRequest.Time + int64(common.EnterSpaceRequestTimeout))
}

func (e *Entity) requestMigrateTo(spaceID string, spaceKind int32, pos math32.Vector3, viewLayer int32) {
	e.enteringSpaceRequest.EntityID = e.ID
	e.enteringSpaceRequest.EntityLabel = e.TypName()
	e.enteringSpaceRequest.SpaceID = spaceID
	e.enteringSpaceRequest.SpaceKind = spaceKind
	e.enteringSpaceRequest.Pos.X = pos.X
	e.enteringSpaceRequest.Pos.Y = pos.Y
	e.enteringSpaceRequest.Pos.Z = pos.Z
	e.enteringSpaceRequest.Time = time.Now().UnixNano()
	e.enteringSpaceRequest.ViewLayer = viewLayer

	// 请求进入场景
	// err := app.Send(context.Background(), e.ID,
	// 	e.TypName(), "cellmgrapp.spaceservice.enterspace", e.enteringSpaceRequest)
	err := e.SendService(
		context.Background(),
		"spaceservice",
		"cellmgrapp.spaceservice.enterspace",
		e.enteringSpaceRequest,
	)
	if err != nil {
		logger.Errorf("%s enter space(%s) err(%s)", e, spaceID, err)
		e.PushEnterSceneErrorIfNeed(metapart.ErrSpaceRequestFailed)
	}
}

func (e *Entity) PushEnterSceneErrorIfNeed(eee error) {
	if e.UID != "" && eee != nil {
		logger.Errorf("push enter space err. entity:%s err:%s", e, eee)
		retErr, ok := eee.(*err.Error)
		if !ok {
			retErr = metapart.ErrSpaceUnknown(eee)
		}
		e.PushOwnClient(define.RouterAvatarEnterScene, &protos.RspEnter{
			Code: retErr.Code,
			Msg:  retErr.Message,
		})

		if e.enteringSpaceRequest != nil {
			e.enteringSpaceRequest.Time = 0
		}
	}
}

// migrateToSpaceFromBase 迁移到另一个space，这个只会在拿到 space server id 后调用到
func (e *Entity) migrateToSpaceFromBase(spaceID string, spaceKind int32, pos math32.Vector3, spaceServerID string, viewLayer int32) {
	req := &protos.MigrateToCellReq{
		EnterSpaceID:   spaceID,
		EnterSpaceKind: spaceKind,
		EntityInfo: &protos.EntityCellInfo{
			Data:     e.GetCellData(),
			EnterPos: &protos.Vec3{X: pos.X, Y: pos.Y, Z: pos.Z},
			EnterRot: &protos.Vec3{},
		},
		ViewLayer: viewLayer,
	}
	// 通知指定的的cellapp 进行进入场景操作
	// err := app.SendTo(context.Background(), e.ID,
	// 	e.TypName(),
	// 	spaceServerID,
	// 	"cellremote.enterspacefrombase",
	// 	req,
	// )
	err := e.SendServiceTo(
		context.TODO(),
		e.SpaceServerID(),
		"cellremote",
		"cellremote.enterspacefrombase",
		req,
	)
	if err != nil {
		logger.Warn("call enterspace err", zap.String("req", req.String()), zap.Error(err))
		e.PushEnterSceneErrorIfNeed(metapart.ErrSpaceRequestFailed)
	}

	e.resetEnterSpaceRequest()

	// 这个后面还是应该在玩家真正进入场景，就是上面那个 rpc 结束之后再赋值，才是最准确的
	e.baseSpace = &_BaseSpaceInfo{ID: spaceID, Kind: spaceKind, ServerID: spaceServerID}

}

// 只会在业务 app 触发
func (e *Entity) enterSpaceResult(req *protos.EnterSpaceResultNotify) {
	e.resetEnterSpaceRequest()
	if req.Ok {
		logger.Info("enter space success", zap.String("msg", req.String()))

		e.baseSpace = newBaseSpaceInfo(req.SpaceID, req.SpaceKind, req.SpaceServerID)
		e.I.EnterSpaceSuccess(req.SpaceID, req.SpaceKind)
	} else {
		// TODO  目前不会跑到这个分支
		logger.Warn("enter space failed", zap.String("msg", req.String()))
		e.I.EnterSpaceFailed(err.NewError(fmt.Errorf(req.ErrMsg), req.ErrCode))
	}
}

/***************************** timer *****************************/

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

/**************************** push to client *************************/

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

// type basePart interface {
// 	ID() string
// 	UID() string
// 	Label() string
// 	PushEnterSceneErrorIfNeed(eee error)
// }

// type requestEnterSpacePart struct {
// 	b                    basePart
// 	enteringSpaceRequest *protos.EnterSpaceReq
// }

// func (e *requestEnterSpacePart) resetEnterSpaceRequest() {
// 	e.enteringSpaceRequest.Reset()
// 	e.enteringSpaceRequest.FromServerID = app.GetServerID()
// 	if e.enteringSpaceRequest.Pos == nil {
// 		e.enteringSpaceRequest.Pos = &protos.Vec3{X: 0, Y: 0, Z: 0}
// 	} else {
// 		e.enteringSpaceRequest.Pos.X = 0
// 		e.enteringSpaceRequest.Pos.Y = 0
// 		e.enteringSpaceRequest.Pos.Z = 0
// 	}
// }

// func (e *requestEnterSpacePart) requestMigrateTo(spaceID string, spaceKind int32, pos math32.Vector3, viewLayer int32) {
// 	e.enteringSpaceRequest.EntityID = e.b.ID()
// 	e.enteringSpaceRequest.EntityLabel = e.b.Label()
// 	e.enteringSpaceRequest.SpaceID = spaceID
// 	e.enteringSpaceRequest.SpaceKind = spaceKind
// 	e.enteringSpaceRequest.Pos.X = pos.X
// 	e.enteringSpaceRequest.Pos.Y = pos.Y
// 	e.enteringSpaceRequest.Pos.Z = pos.Z
// 	e.enteringSpaceRequest.Time = time.Now().UnixNano()
// 	e.enteringSpaceRequest.ViewLayer = viewLayer

// 	// 请求进入场景
// 	err := app.Send(context.Background(), "cellmgrapp.spaceservice.enterspace", e.enteringSpaceRequest)
// 	if err != nil {
// 		logger.Errorf("%s enter space(%s) err(%s)", e, spaceID, err)
// 		e.b.PushEnterSceneErrorIfNeed(metapart.ErrSpaceRequestFailed)
// 	}
// }
