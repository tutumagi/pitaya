package entity

import (
	"reflect"

	"github.com/tutumagi/pitaya/logger"
	"github.com/tutumagi/pitaya/math32"

	uuid "github.com/satori/go.uuid"
	"github.com/tutumagi/pitaya/common"
	"go.uber.org/zap"
)

var (
	registerEntityTypes = map[string]*TypeDesc{}
	entityManager       = newEntityManager()
)

type _EntityManager struct {
	entitiesByType map[string]Map
}

func newEntityManager() *_EntityManager {
	return &_EntityManager{
		entitiesByType: map[string]Map{},
	}
}

func (em *_EntityManager) put(entity *Entity) {
	if em.entitiesByType[entity.typName] == nil {
		em.entitiesByType[entity.typName] = make(Map, 1000)
	}
	m, ok := em.entitiesByType[entity.typName]
	if ok {
		m.Add(entity)
	} else {
		m = make(Map, 1000)
		m.Add(entity)
		em.entitiesByType[entity.typName] = m
	}
}

func (em *_EntityManager) del(e *Entity) {
	if m, ok := em.entitiesByType[e.typName]; ok {
		m.Del(e.ID)
	}
}

func (em *_EntityManager) get(typName string, id string) *Entity {
	if m, ok := em.entitiesByType[typName]; ok {
		return m.Get(id)
	}
	return nil
}

// RegisterEntity 注册实体
func RegisterEntity(typName string, entity IEntity, model interface{}, persistant bool, useAOI ...bool) *TypeDesc {
	if _, ok := registerEntityTypes[typName]; ok {
		logger.Log.Warnf("entity type %s already register", typName)
		return nil
	}

	entityVal := reflect.ValueOf(entity)
	entityType := entityVal.Type()

	if entityType.Kind() == reflect.Ptr {
		entityType = entityType.Elem()
	}

	var entityModelType reflect.Type
	if model != nil {
		entityModelType = reflect.TypeOf(model)
		if entityModelType.Kind() == reflect.Ptr {
			entityModelType = entityModelType.Elem()
		}
	}

	beUseAOI := false
	if len(useAOI) > 0 {
		beUseAOI = useAOI[0]
	}

	typeDesc := &TypeDesc{
		IsPersistent: persistant,
		useAOI:       beUseAOI,
		aoiDistance:  0,
		eTyp:         entityType,
		mTyp:         entityModelType,
	}

	registerEntityTypes[typName] = typeDesc

	return typeDesc
}

// GetTypeDesc 获取实体类型信息
func GetTypeDesc(typName string) *TypeDesc {
	return registerEntityTypes[typName]
}

// createEntity 创建实体
//	 isCreate 表示是否是新建的实体（DB里面没有的）
func createEntity(typName string, entityID string, data interface{}, space *Space, pos math32.Vector3, isCreate bool) *Entity {
	entity := createEntityOnlyInit(typName, entityID, data, space, isCreate)

	entity.I.OnCreated()

	if space != nil {
		if pos != math32.ZeroVec3 {
			space.enter(entity, pos)
		} else {
			space.enter(entity, entity.I.DefaultPos())
		}
	}

	return entity
}

// createEntityOnlyInit 创建实体
//	 isCreate 表示是否是新建的实体（DB里面没有的）
func createEntityOnlyInit(typName string, entityID string, data interface{}, space *Space, isCreate bool) *Entity {
	typeDesc, ok := registerEntityTypes[typName]
	if !ok {
		logger.Log.Errorf("unknown entity type:%s", typName)
	}

	if entityID == "" {
		entityID = uuid.NewV1().String()
	}

	var entity *Entity
	var entityInstance reflect.Value

	entityInstance = reflect.New(typeDesc.eTyp)
	entity = reflect.Indirect(entityInstance).FieldByName("Entity").Addr().Interface().(*Entity)

	entity.init(typName, entityID, entityInstance)
	entity.Space = nil

	entityManager.put(entity)
	// 如果数据不为空，表示是从db load过来的
	if data != nil {
		entity.Data = data
	} else {
		// 如果初始化数据为空，获取默认的数据
		entity.Data = entity.I.DefaultModel(entityID)
	}

	if isCreate == true {
		entity.Save()
	}

	return entity
}

func loadEntity(typName string, entityID string, space *Space) *Entity {
	typDesc, ok := registerEntityTypes[typName]
	if !ok {
		logger.Log.Warn("entity type not register", zap.String("typName", typName))
		// TODO 通知 加载entity 失败了
		return nil
	}

	model := reflect.New(typDesc.mTyp).Interface()

	// TODO 这里要改为异步的
	// param := dbmgr.QueryPara{
	// 	TblName: define.EntityTableName(typName),
	// 	KeyName: "id",
	// 	KeyStr:  entityID,
	// }

	// code, err := dbmgr.Get(param, model)
	_, err := storage.Load(common.EntityTableName(typName), "id", entityID, model)

	if err != nil {
		// 是空数据的错误，不打印日志
		// if code != dbmgr.DB_DATA_EMPTY {
		// 	logger.Log.Error("load entity from db err", zap.String("typName", typName), zap.String("id", entityID), zap.Int("errCode", code), zap.Error(err))
		// }
		return nil
	}

	ex := entityManager.get(typName, entityID)
	if ex != nil {
		logger.Log.Warn("has already laod the entity", zap.String("typName", typName), zap.String("id", entityID))
		return ex
	}

	if space != nil && space.IsDestroyed() {
		space = nil
	}

	return createEntity(typName, entityID, model, space, math32.ZeroVec3, false)
}

// GetEntity get entity
func GetEntity(typName string, id string) *Entity {
	return entityManager.get(typName, id)
}

// GetEntitiesByType 根据实体类型获取实体map
func GetEntitiesByType(typName string) Map {
	return entityManager.entitiesByType[typName]
}
