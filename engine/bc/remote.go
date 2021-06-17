package bc

// import (
// 	"context"

// 	"github.com/tutumagi/pitaya/logger"
// 	"github.com/tutumagi/pitaya/engine/math32"
// 	"gitlab.gamesword.com/nut/dreamcity/pb/golang/pb"

// 	"github.com/tutumagi/pitaya"
// 	"github.com/tutumagi/pitaya/component"
// 	"github.com/tutumagi/pitaya/protos"
// 	"go.uber.org/zap"
// )

// // EntityRemoteInBase 实体相关操作，只允许在业务 app 注册
// type EntityRemoteInBase struct {
// 	component.Base
// }

// // Init calls when pitaya start
// func (r *EntityRemoteInBase) Init() {
// 	if svrConfig.curServerUseSpace {
// 		panic("EntityRemoteInBase 只允许在业务 app 注册")
// 	}
// }

// // DoEnterSpace 玩家进入场景，当cellmgrapp 拿到 场景服的地址后 通知过来的
// func (r *EntityRemoteInBase) DoEnterSpace(ctx context.Context, req *pb.RealEnterSpaceReq) (*protos.Response, error) {
// 	logger.Debug("entity remote do enterspace", zap.String("req", req.String()))

// 	if req.EntityID == "" || req.SpaceID == "" {
// 		logger.Warn("do enter space entityid or spaceid is empty")
// 		return &protos.Response{}, nil
// 	}

// 	e := GetEntity(req.EntityLabel, req.EntityID)
// 	if e == nil {
// 		// 可能是玩家下线了，这里就会拿不到该玩家
// 		logger.Warnf("do enter space entity is not exist %s %s", req.EntityLabel, req.EntityID)
// 		return &protos.Response{}, nil
// 	}

// 	logger.Debug("entity in base server, now request space server do enterspace", zap.String("req", req.String()))
// 	e.migrateToSpaceFromBase(
// 		req.SpaceID,
// 		req.SpaceKind,
// 		math32.Vector3{X: req.EnterPos.X, Y: req.EnterPos.Y, Z: req.EnterPos.Z},
// 		req.SpaceServerID,
// 		req.ViewLayer,
// 	)

// 	return &protos.Response{}, nil
// }

// // EnterSpaceResult 玩家进入场景后的处理，这里主要处理 base 的 实体相关的空间记录
// func (r *EntityRemoteInBase) EnterSpaceResult(ctx context.Context, req *pb.EnterSpaceResultNotify) (*protos.Response, error) {
// 	logger.Infof("entity enter space result %s", req.String())

// 	e := GetEntity(req.EntityLabel, req.EntityID)
// 	if e == nil {
// 		// 可能是玩家下线了，这里就会拿不到该玩家
// 		logger.Warn("do enter space entity is not exist")
// 		return &protos.Response{}, nil
// 	}

// 	e.enterSpaceResult(req)

// 	return &protos.Response{}, nil
// }

// // CreateBaseSpaces  在 baseapp 上创建空间
// func (r *EntityRemoteInBase) CreateBaseSpace(
// 	ctx context.Context,
// 	req *pb.CreateSpaceReq,
// ) (*protos.Response, error) {
// 	err := createBaseSpaceFromRemote(req)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &protos.Response{}, nil
// }

// // CreateBaseSpaces  在 baseapp 上创建空间(新增这个协议是为了不影响之前的流程)
// // 主要用于解决base、cell 和 cellmgr 之间的启动依赖。
// func (r *EntityRemoteInBase) CreateBaseSpaceIfNeed(
// 	ctx context.Context,
// 	req *pb.CreateSpaceReq,
// ) (*protos.Response, error) {

// 	logger.Infof("CreateBaseSpaceIfNeed, req=%+v", req)

// 	//判断 MasterSpace是否已经存在
// 	// assert(req.SpaceID == define.MasterSpaceID)
// 	sp := GetSpace(req.SpaceID) //GetSpace(define.MasterSpaceID)
// 	if sp == nil {
// 		err := createBaseSpaceFromRemote(req)
// 		if err != nil {
// 			logger.Errorf("createBaseSpaceFromRemote failed. id:%s kind:%d cellServerID:%s err:%s", req.SpaceID, req.SpaceKind, req.FreeCellServerID, err)
// 			return nil, err
// 		}

// 	} else {

// 		//sp不为空，且能收到这个消息，说明cellMgr 或者 cell 因某种原因被重新启动了。
// 		logger.Infof("CreateBaseSpaceIfNeed, my=%s, req=%s", sp.initCellServerID, req.FreeCellServerID)
// 		if sp.initCellServerID != req.FreeCellServerID {
// 			sp.initCellServerID = req.FreeCellServerID
// 		}

// 		err := pitaya.SendTo(context.TODO(),
// 			req.FreeCellServerID,
// 			"cellremote.createcellspace",
// 			&pb.CreateSpaceReq{
// 				SpaceKind:        req.SpaceKind,
// 				SpaceID:          req.SpaceID,
// 				FreeCellServerID: req.FreeCellServerID,
// 				Extra:            map[string]string{"BaseServerID": pitaya.GetServerID()},
// 			},
// 		)
// 		if err != nil {
// 			logger.Errorf("create cell space failed! id:%s kind:%d cellServerID:%s err:%s", req.SpaceID, req.SpaceKind, req.FreeCellServerID, err)
// 			return nil, err
// 		}

// 		// 通知 cellappmgr 场景已存在
// 		err = pitaya.Send(context.TODO(),
// 			"cellmgrapp.spaceservice.docreatespaceifneednotify",
// 			&pb.CreateSpaceIfNeedNotify{
// 				SpaceKind:       sp.kind,
// 				SpaceID:         sp.ID,
// 				BaseServerID:    pitaya.GetServerID(),
// 				CellServerID:    sp.initCellServerID,
// 				MasterSpaceFlag: 1,
// 			},
// 		)
// 		if err != nil {
// 			logger.Warn("notify space load failed", zap.Error(err))
// 		}
// 	}

// 	return &protos.Response{}, nil
// }

// // // CreateSpace 创建实体
// // func (r *EntityComponent) CreateEntity(
// // 	ctx context.Context,
// // 	req *pb.MigrateEntityData,
// // ) (*protos.Response, error) {
// // 	err := createEntityFromRemote(req)
// // 	if err != nil {
// // 		return nil, err
// // 	}

// // 	return &protos.Response{}, nil
// // }

// // DestroyEntity 销毁实体
// func (r *EntityRemoteInBase) DestroyEntity(ctx context.Context, req *pb.Entity) (*protos.Response, error) {
// 	// pitaya.GroupRemoveMember(ctx, groupName, req.Uid)

// 	e := GetEntity(req.Label, req.Id)
// 	if e != nil {
// 		e.Destroy(req.Reason)
// 	} else {
// 		logger.Warn("has no entity", zap.String("entity", req.String()))
// 	}

// 	return &protos.Response{}, nil
// }

// // func (r *EntityRemoteInBase) MoveRot(ctx context.Context, req *pb.EntityMoveRot) (*protos.Response, error) {
// // 	// logger.Debugf("sync pos %s", req.String())
// // 	e := GetEntity(req.Label, req.Id)
// // 	if e != nil {
// // 		if req.Pos != nil && req.Rot != nil {
// // 			e.MoveAndRot(math32.Vector3{X: req.Pos.X, Y: req.Pos.Y, Z: req.Pos.Z}, req.Rot.Y)
// // 		} else if req.Pos != nil {
// // 			e.Move(math32.Vector3{X: req.Pos.X, Y: req.Pos.Y, Z: req.Pos.Z})
// // 		} else if req.Rot != nil {
// // 			e.Rot(req.Rot.Y)
// // 		}
// // 	} else {
// // 		logger.Warn("has no entity", zap.String("entityID", req.Id))
// // 	}

// // 	return &protos.Response{}, nil
// // }
