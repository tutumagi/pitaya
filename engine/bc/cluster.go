package bc

// func CreateEntitySomewhere(
// 	id string,
// 	label string,
// 	data map[string]interface{},
// 	shouldSaveDB bool,
// ) error {
// 	freeBaseServer, err := getFreeBaseServer()
// 	if err != nil {
// 		return err
// 	}

// 	databytes, _ := json.Marshal(data)
// 	logger.Infof("create entity id:%s label:%s datalen:%d saveDB:%s", id, label, len(databytes), shouldSaveDB)
// 	entityData := &pb.SEntityData{
// 		UserID:   "",
// 		EntityID: id,
// 		// 实体typName
// 		EntityLabel: label,
// 		// 实体序列化后的数据，目前使用json
// 		EntityDatas:  databytes,
// 		FromServerID: pitaya.GetServerID(),
// 	}
// 	err = pitaya.RPCTo(
// 		context.TODO(),
// 		freeBaseServer.ID,
// 		"entity.createentity",
// 		&protos.Response{},
// 		entityData,
// 	)

// 	if err != nil {
// 		logger.Errorf("create entity(id:%s label:%s) somethere err:%s",
// 			id,
// 			label,
// 			err,
// 		)
// 	}
// 	return err
// }

// func CreateEntityLocal(
// 	id string,
// 	label string,
// 	// data map[string]interface{},
// 	data *attr.StrMap,
// 	shouldSaveDB bool,
// ) error {
// 	CreateEntity(label, id, data, shouldSaveDB)

// 	return nil
// }

/******************************** private method *********************************/
// func createEntityFromRemote(entityData *pb.MigrateEntityData) error {
// 	id := entityData.EntityID
// 	label := entityData.EntityLabel
// 	logger.Debug("createEntityFromRemote ",
// 		zap.String("entityID", id),
// 		zap.String("entityLabel", label),
// 		zap.String("leaveSpaceID", entityData.SpaceID),
// 	)

// 	// TODO 这里进入成功后，要publish 一下，可以先做玩家的，因为 cellmgrapp 和 baseapp 都需要直到玩家在哪个场景
// 	e := GetEntity(label, id)
// 	if e != nil {
// 		return fmt.Errorf("已经存在该实体了 id:%s label:%s", id, label)
// 	}

// 	typDesc := GetTypeDesc(label)
// 	if typDesc == nil {
// 		return fmt.Errorf("没有该实体类型 %s", label)
// 	}
// 	entityMapData, err := typDesc.meta.UnmarshalJson(entityData.EntityDatas)
// 	if err != nil {
// 		return fmt.Errorf("数据解析错误")
// 	}

// 	return CreateEntityLocal(
// 		label,
// 		id,
// 		entityMapData,
// 		entityData.ShoudSaveDB,
// 	)
// }

// func getFreeBaseServer() (*cluster.Server, error) {
// 	servers, err := pitaya.GetServersByType(metapart.BaseAppSrv)
// 	if err != nil {
// 		return nil, err
// 	}

// 	for _, s := range servers {
// 		if s.Type == metapart.BaseAppSrv {
// 			return s, nil
// 		}
// 	}
// 	logger.Errorf("找不到负载最小的 base server")
// 	return nil, metapart.ErrInvalidArgs
// }
