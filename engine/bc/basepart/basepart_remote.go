package basepart

import (
	"context"
	"fmt"

	"github.com/tutumagi/pitaya/agent"
	"github.com/tutumagi/pitaya/cluster"
	"github.com/tutumagi/pitaya/engine/bc/metapart"
	"github.com/tutumagi/pitaya/engine/math32"
	"github.com/tutumagi/pitaya/logger"
	"github.com/tutumagi/pitaya/serialize"
	"gitlab.gamesword.com/nut/entitygen/attr"

	"github.com/tutumagi/pitaya/protos"
	"go.uber.org/zap"
)

// Remote 实体相关操作，只允许在业务 app 注册
type Remote struct {
	Entity

	// 因为创建远端client，需要下面三个字段，先放到这里，
	// 见 ClientConnected
	rpcClient cluster.RPCClient
	// encoder codec.PacketEncoder,
	serializer       serialize.Serializer
	serviceDiscovery cluster.ServiceDiscovery
}

func NewRemote(
	rpcClient cluster.RPCClient,
	serializer serialize.Serializer,
	serviceDiscovery cluster.ServiceDiscovery,
) *Remote {
	return &Remote{
		rpcClient:        rpcClient,
		serializer:       serializer,
		serviceDiscovery: serviceDiscovery,
	}
}

type EntityHandler struct{}

// DoEnterSpace 玩家进入场景，当cellmgrapp 拿到 场景服的地址后 通知过来的
func (r *EntityHandler) DoEnterSpace(ctx context.Context, req *protos.RealEnterSpaceReq) (*protos.Response, error) {
	logger.Debug("entity remote do enterspace", zap.String("req", req.String()))

	if req.EntityID == "" || req.SpaceID == "" {
		logger.Warn("do enter space entityid or spaceid is empty")
		return &protos.Response{}, nil
	}

	e := GetEntity(req.EntityLabel, req.EntityID)
	if e == nil {
		// 可能是玩家下线了，这里就会拿不到该玩家
		logger.Warnf("do enter space entity is not exist %s %s", req.EntityLabel, req.EntityID)
		return &protos.Response{}, nil
	}

	logger.Debug("entity in base server, now request space server do enterspace", zap.String("req", req.String()))
	e.migrateToSpaceFromBase(
		req.SpaceID,
		req.SpaceKind,
		math32.Vector3{X: req.EnterPos.X, Y: req.EnterPos.Y, Z: req.EnterPos.Z},
		req.SpaceServerID,
		req.ViewLayer,
	)

	return &protos.Response{}, nil
}

// EnterSpaceResult 玩家进入场景后的处理，这里主要处理 base 的 实体相关的空间记录
func (r *EntityHandler) EnterSpaceResult(ctx context.Context, req *protos.EnterSpaceResultNotify) (*protos.Response, error) {
	logger.Infof("entity enter space result %s", req.String())

	e := GetEntity(req.EntityLabel, req.EntityID)
	if e == nil {
		// 可能是玩家下线了，这里就会拿不到该玩家
		logger.Warn("do enter space entity is not exist")
		return &protos.Response{}, nil
	}

	e.enterSpaceResult(req)

	return &protos.Response{}, nil
}

// CreateBaseSpaces  在 baseapp 上创建空间
func (r *EntityHandler) CreateBaseSpace(
	ctx context.Context,
	entity interface{},
	req *protos.CreateBaseSpaceReq,
) (*protos.Response, error) {
	err := r.createBaseSpaceFromRemote(entity, req)
	if err != nil {
		return nil, err
	}

	return &protos.Response{}, nil
}

// CreateBaseSpaces  在 baseapp 上创建空间(新增这个协议是为了不影响之前的流程)
// 主要用于解决base、cell 和 cellmgr 之间的启动依赖。
func (r *EntityHandler) CreateBaseSpaceIfNeed(
	ctx context.Context,
	entity interface{},
	req *protos.CreateBaseSpaceReq,
) (*protos.Response, error) {
	caller := entity.(*Remote)
	logger.Infof("CreateBaseSpaceIfNeed, req=%+v", req)

	//判断 MasterSpace是否已经存在
	// assert(req.SpaceID == metapart.MasterSpaceID)
	sp := GetSpace(req.SpaceID) //GetSpace(metapart.MasterSpaceID)
	if sp == nil {
		err := r.createBaseSpaceFromRemote(entity, req)
		if err != nil {
			logger.Errorf("createBaseSpaceFromRemote failed. id:%s kind:%d cellServerID:%s err:%s", req.SpaceID, req.SpaceKind, req.CellServerID, err)
			return nil, err
		}

	} else {

		//sp不为空，且能收到这个消息，说明cellMgr 或者 cell 因某种原因被重新启动了。
		logger.Infof("CreateBaseSpaceIfNeed, my=%s, req=%s", sp.initCellServerID, req.CellServerID)
		if sp.initCellServerID != req.CellServerID {
			sp.initCellServerID = req.CellServerID
		}

		// err := app.SendTo(context.TODO(),
		// 	"",
		// 	"",
		// 	req.CellServerID,
		// 	"cellremote.createcellspace",
		// 	&protos.CreateCellSpaceReq{
		// 		SpaceKind:    req.SpaceKind,
		// 		SpaceID:      req.SpaceID,
		// 		BaseServerID: curServerID(),
		// 		Extra:        map[string]string{},
		// 	},
		// )
		err := caller.SendServiceTo(
			context.TODO(),
			req.CellServerID,
			"cellremote",
			"cellremote.createcellspace",
			&protos.CreateCellSpaceReq{
				SpaceKind:    req.SpaceKind,
				SpaceID:      req.SpaceID,
				BaseServerID: curServerID(),
				Extra:        map[string]string{},
			},
		)
		if err != nil {
			logger.Errorf("create cell space failed! id:%s kind:%d cellServerID:%s err:%s", req.SpaceID, req.SpaceKind, req.CellServerID, err)
			return nil, err
		}

		// 通知 cellappmgr 场景已存在
		// err = app.Send(context.TODO(),
		// 	"",
		// 	"",
		// 	"cellmgrapp.spaceservice.docreatespaceifneednotify",
		// 	&protos.CreateSpaceIfNeedNotify{
		// 		SpaceKind:       sp.kind,
		// 		SpaceID:         sp.ID,
		// 		BaseServerID:    curServerID(),
		// 		CellServerID:    sp.initCellServerID,
		// 		MasterSpaceFlag: 1,
		// 	},
		// )
		err = caller.SendService(
			context.TODO(),
			"spaceservice",
			"cellmgrapp.spaceservice.docreatespaceifneednotify",
			&protos.CreateSpaceIfNeedNotify{
				SpaceKind:       sp.kind,
				SpaceID:         sp.ID,
				BaseServerID:    curServerID(),
				CellServerID:    sp.initCellServerID,
				MasterSpaceFlag: 1,
			},
		)
		if err != nil {
			logger.Warn("notify space load failed", zap.Error(err))
		}
	}

	return &protos.Response{}, nil
}

// // CreateSpace 创建实体
// func (r *EntityComponent) CreateEntity(
// 	ctx context.Context,
// 	req *protos.MigrateEntityData,
// ) (*protos.Response, error) {
// 	err := createEntityFromRemote(req)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &protos.Response{}, nil
// }

// DestroyEntity 销毁实体
func (r *EntityHandler) DestroyEntity(ctx context.Context, req *protos.Entity) (*protos.Response, error) {
	// app.GroupRemoveMember(ctx, groupName, req.Uid)

	e := GetEntity(req.Label, req.Id)
	if e != nil {
		e.Destroy(req.Reason)
	} else {
		logger.Warn("has no entity", zap.String("entity", req.String()))
	}

	return &protos.Response{}, nil
}

func (r *EntityHandler) CreateEntity(ctx context.Context, req *protos.SEntityData) (*protos.Response, error) {
	ent := GetEntity(req.EntityLabel, req.EntityID)

	// 如果该Server 没有该实体， 则拿到迁移数据进行 重建实体，这里重建实体不包括timer，只包括数据
	typDesc := metapart.GetTypeDesc(req.EntityLabel)
	if typDesc == nil {
		return nil, fmt.Errorf("没有该实体类型 %s", req.EntityLabel)
	}
	var attrs *attr.StrMap
	var err error
	if req.EntityDatas != nil {
		attrs, err = typDesc.UnmarshalJSON(req.EntityDatas)
		if err != nil {
			return nil, fmt.Errorf("实体数据解析错误")
		}
	}

	ent = baseEntManager.CreateEntity(
		req.EntityLabel,
		req.EntityID,
		attrs,
	)
	logger.Debugf("app.handler Space::CreateEntity e = %+v", ent)
	if ent == nil {
		logger.Warn("create entity in cell return nil", zap.String("req", req.String()))
	} else {
		// ent.SetBaseServerID(req.FromServerID)
		// // 这里把第一次进场景的视野参数给保存住
		// ent.Viewlayer = req.ViewLayer

		// s.EnterSingleEntity(ent, math32.Vector3{
		// 	X: req.EntityInfo.EnterPos.X,
		// 	Y: req.EntityInfo.EnterPos.Y,
		// 	Z: req.EntityInfo.EnterPos.Z,
		// })
	}
	return &protos.Response{}, nil
}

func (r *EntityHandler) ClientConnected(ctx context.Context, entity interface{}, req *protos.ClientConnect) (*protos.Response, error) {
	entityService := entity.(*Remote)
	client, _ := agent.NewRemote(
		req.Sess.Id,
		req.Sess.ServerID,
		req.Sess.Uid,
		req.Sess.RoleID,
		entityService.rpcClient,
		entityService.serializer,
		entityService.serviceDiscovery,
	)
	// TODO 用配置表
	// bootEntityType := app.config.GetString("pitaya.bootentity")
	bootEntityType := "acount"
	e := CreateEntity(bootEntityType, req.BootEntityID, nil, false)
	e.SetClient(client)

	return &protos.Response{}, nil
}

func (r *EntityHandler) createBaseSpaceFromRemote(entity interface{}, req *protos.CreateBaseSpaceReq) error {
	caller := entity.(*Entity)
	logger.Infof("create base space. id:%s kind:%d cellServerID:%s", req.SpaceID, req.SpaceKind, req.CellServerID)
	space := CreateSpace(req.SpaceKind, req.SpaceID, req.CellServerID)
	logger.Infof("create base and cell space success. id:%s kind:%d cellServerID:%s", req.SpaceID, req.SpaceKind, req.CellServerID)

	spaceExtra := space.I.PrepareCellData()
	if req.Extra == nil {
		req.Extra = spaceExtra
	} else {
		for k, v := range spaceExtra {
			req.Extra[k] = v
		}
	}

	logger.Infof("create cell space. id:%s kind:%d cellServerID:%s", req.SpaceID, req.SpaceKind, req.CellServerID)
	// err := app.SendTo(context.TODO(),
	// 	"",
	// 	"",
	// 	req.CellServerID,
	// 	"cellremote.createcellspace",
	// 	&protos.CreateCellSpaceReq{
	// 		SpaceKind:    req.SpaceKind,
	// 		SpaceID:      req.SpaceID,
	// 		BaseServerID: curServerID(),
	// 		Extra:        req.Extra,
	// 	},
	// )
	err := caller.SendServiceTo(
		context.TODO(),
		req.CellServerID,
		"cellremote",
		"cellremote.createcellspace",
		&protos.CreateCellSpaceReq{
			SpaceKind:    req.SpaceKind,
			SpaceID:      req.SpaceID,
			BaseServerID: curServerID(),
			Extra:        req.Extra,
		},
	)
	if err != nil {
		logger.Errorf("create cell space failed. id:%s kind:%d cellServerID:%s err:%s", req.SpaceID, req.SpaceKind, req.CellServerID, err)
		return err
	}
	logger.Infof("create cell space success. id:%s kind:%d cellServerID:%s", req.SpaceID, req.SpaceKind, req.CellServerID)

	return space.cellPartCreated()
}

// TODO 为了避免循环引用，后面要考虑怎么弄这个
func curServerID() string {
	// return app.GetServerID()
	return ""
}
