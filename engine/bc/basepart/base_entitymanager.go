package basepart

import (
	"reflect"
	"sync"
	"unsafe"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/tutumagi/pitaya/engine/bc/metapart"
	"github.com/tutumagi/pitaya/engine/dbmgr"
	"github.com/tutumagi/pitaya/logger"
	"gitlab.gamesword.com/nut/dreamcity/game/define"
	"gitlab.gamesword.com/nut/entitygen/attr"
	"go.uber.org/zap"
)

type _BaseEntityManager struct {
	sync.RWMutex
	entitiesByType map[string]BaseMap

	system *actor.ActorSystem
}

func newBaseEntityManager(system *actor.ActorSystem) *_BaseEntityManager {
	return &_BaseEntityManager{
		entitiesByType: map[string]BaseMap{},

		system: system,
	}
}

func (em *_BaseEntityManager) put(entity *Entity) {
	em.Lock()
	defer em.Unlock()

	typName := entity.TypName()
	if em.entitiesByType[typName] == nil {
		em.entitiesByType[typName] = make(BaseMap, 1000)
	}
	m, ok := em.entitiesByType[typName]
	if ok {
		m.Add(entity)
	} else {
		m = make(BaseMap, 1000)
		m.Add(entity)
		em.entitiesByType[typName] = m
	}
}

func (em *_BaseEntityManager) del(e *Entity) {
	em.Lock()
	defer em.Unlock()

	if m, ok := em.entitiesByType[e.TypName()]; ok {
		m.Del(e.ID)
	}
}

func (em *_BaseEntityManager) get(id string, typName string) *Entity {
	em.RLock()
	defer em.RUnlock()

	if m, ok := em.entitiesByType[typName]; ok {
		return m.Get(id)
	}
	return nil
}

func (em *_BaseEntityManager) getPid(id string, typName string) *actor.PID {
	em.RLock()
	defer em.RUnlock()

	if m, ok := em.entitiesByType[typName]; ok {
		return m.Get(id).pid
	}
	return nil
}

func (em *_BaseEntityManager) getEntitiesByTypName(typName string) BaseMap {
	em.RLock()
	defer em.RUnlock()

	return em.entitiesByType[typName]
}

func (em *_BaseEntityManager) foreach(fn func(*Entity) bool) {
	em.RLock()
	defer em.RUnlock()

	for _, es := range em.entitiesByType {
		for _, e := range es {
			if !fn(e) {
				goto exit
			}
		}
	}

exit:
}

func (em *_BaseEntityManager) CreateEntity(
	label string,
	id string,
	data *attr.StrMap,
) *Entity {
	if id == "" {
		id = metapart.NewUUID()
	}
	entity := em.get(label, id)
	if entity != nil {
		logger.Warnf("id duplicated %s", id)
		return entity
	}
	entity = createBaseEntityOnlyInit(id, label)
	err := entity.I.OnInit()
	// TODO 这里要返回错误码
	if err != nil {
		return nil
	}

	// 创建pid给该实体
	props := actor.PropsFromProducer(func() actor.Actor {
		return entity
	})
	// 这里命名的pid的组成方式，一方面为了出错时的调试，一方面是为了唯一性
	entity.pid, _ = em.system.Root.SpawnNamed(props, label+entity.ID)

	em.put(entity)

	// entity.data = attr.NewStrMap(data)
	if data == nil {
		entity.data = attr.NewStrMap(nil)
	} else {
		entity.data = data
	}

	entity.data.Set("id", entity.ID)
	entity.data.ClearChangeKey()

	if entity.IsPersistent() {
		entity.saveToDB()
	}

	if err != nil {
		logger.Warnf("create entity failed. id:%s label:%s err:%s",
			id, label, data,
		)
		return nil
	}

	entity.I.OnCreated()

	return entity
}

// func (em *_BaseEntityManager) count() int {
// 	em.RLock()
// 	defer em.RUnlock()

// }

// var baseEntManager = newBaseEntityManager()

// GetEntity get entity
func GetEntity(typName string, id string) *Entity {
	return baseEntManager.get(id, typName)
}

// GetEntitiesByType 根据实体类型获取实体map
func GetEntitiesByType(typName string) BaseMap {
	return baseEntManager.getEntitiesByTypName(typName)
}

// 遍历所有实体，如果返回false，则停止遍历
func ForEach(fn func(e *Entity) bool) {
	baseEntManager.foreach(fn)
}

func CreateEntity(
	label string,
	id string,
	// data map[string]interface{},
	data *attr.StrMap,
	shouldSaveDB bool,
) *Entity {
	return baseEntManager.CreateEntity(label, id, data)
	// if id == "" {
	// 	id = metapart.NewUUID()
	// }
	// entity := baseEntManager.get(label, id)
	// if entity != nil {
	// 	logger.Warnf("id duplicated %s", id)
	// 	return entity
	// }
	// entity = createBaseEntityOnlyInit(id, label)
	// err := entity.I.OnInit()
	// // TODO 这里要返回错误码
	// if err != nil {
	// 	return nil
	// }

	// // 创建pid给该实体
	// props := actor.PropsFromProducer(func() actor.Actor {
	// 	return entity
	// })
	// entity.pid, _ = baseEntManager.system.Root.SpawnNamed(props, "entity_"+entity.ID)

	// baseEntManager.put(entity)

	// // entity.data = attr.NewStrMap(data)
	// if data == nil {
	// 	entity.data = attr.NewStrMap(nil)
	// } else {
	// 	entity.data = data
	// }

	// entity.data.Set("id", entity.ID)
	// entity.data.ClearChangeKey()

	// if shouldSaveDB {
	// 	entity.saveToDB()
	// }

	// if err != nil {
	// 	logger.Warnf("create entity failed. id:%s label:%s shouldSaveDB:%s err:%s",
	// 		id, label, data, shouldSaveDB,
	// 	)
	// 	return nil
	// }

	// entity.I.OnCreated()

	// return entity
}

// createBaseEntityOnlyInit 创建实体
//	 isCreate 表示是否是新建的实体（DB里面没有的）
func createBaseEntityOnlyInit(
	id string,
	label string,
) *Entity {
	typeDesc := metapart.GetTypeDesc(label)
	if typeDesc == nil {
		logger.Panic("unknown entity type", zap.String("name", label))
	}

	entityInstance := reflect.New(typeDesc.BaseTyp())
	entity := reflect.Indirect(entityInstance).FieldByName("Entity").Addr().Interface().(*Entity)

	entity.init(id, label, entityInstance)

	return entity
}

// LoadEntity 加载实体
func LoadEntity(typName string, entityID string) (*Entity, error) {
	param := dbmgr.QueryPara{
		TblName: define.EntityTableName(typName),
		KeyName: "id",
		Key:     entityID,
	}

	return LoadFilterEntity(typName, param)
}

// LoadFilterEntities 加载大量实体，不会加入到任何场景中去
func LoadFilterEntities(typName string, queryFilter dbmgr.QueryPara) []*Entity {
	typDesc := metapart.GetTypeDesc(typName)
	if typDesc == nil {
		logger.Warn("entity type not register", zap.String("typName", typName))
		// TODO 通知 加载entity 失败了
		return nil
	}
	logger.Info("load many entity from db",
		zap.String("typName", typName),
		zap.Any("queryFilter", queryFilter))

	// var model map[string]interface{} = map[string]interface{}{}
	// model := typDesc.DynamicSliceOfStruct()
	model := typDesc.Meta().CreateSlice()
	//数据中有很多是双主键设置，但有些是以单主键去查询，redis不支持双主键数据间主键查询，故这里还是直接读取mongodb操作，看后面优化再修改
	if err := dbmgr.GetFromMongo(queryFilter, model); err != nil {
		// 是空数据的错误，不打印日志
		if err != dbmgr.ErrDataEmpty {
			logger.Error("load many entity from db err",
				zap.String("typName", typName),
				zap.Any("queryFilter", queryFilter),
				zap.Error(err))
		}
		return nil
	}

	// 将 model 转为 []map[string]interface{}
	// manyAttrs := typDesc.meta.UnmarshalSlice(model)
	// TODO  强依赖 meta 里面的实现
	manyAttrs := *(*[]*attr.StrMap)(unsafe.Pointer(reflect.ValueOf(model).Elem().UnsafeAddr()))

	var entities = []*Entity{}
	for _, attrs := range manyAttrs {
		// entityID := attrs["id"].(string)
		entityID := attrs.Str("id")
		ex := baseEntManager.get(typName, entityID)
		if ex != nil {
			logger.Warn("has already laod the entity", zap.String("typName", typName), zap.String("id", entityID))
			entities = append(entities, ex)
			continue
		}
		if entityID == "" {
			logger.Errorf("load entity from db. but db id is empty (dbID:%s attrs:%v)", entityID, attrs)
		}
		e := CreateEntity(typName, entityID, attrs, false)

		entities = append(entities, e)
	}

	return entities
}

func LoadFilterEntity(typName string, queryFilter dbmgr.QueryPara) (*Entity, error) {
	typDesc := metapart.GetTypeDesc(typName)
	if typDesc == nil {
		logger.Warn("entity type not register", zap.String("typName", typName))
		// TODO 通知 加载entity 失败了
		return nil, define.ErrEntityNotRegister
	}

	_ = typDesc

	model := typDesc.Meta().Create()

	if err := dbmgr.Get(queryFilter, model); err != nil {
		// 是空数据的错误，不打印日志
		if err != dbmgr.ErrDataEmpty {
			logger.Error("load entity from db err", zap.String("typName", typName), zap.Any("query", queryFilter), zap.Error(err))
			return nil, define.ErrLoadDB
		}
		return nil, nil
	}

	// NOTE: 看下有没有可能不用反射，拿到实际的 interface wrapper 住的数据
	attrs := (*attr.StrMap)(unsafe.Pointer(reflect.ValueOf(model).Elem().UnsafeAddr()))
	// attrs := (*attr.StrMap)(unsafe.Pointer(uintptr(unsafe.Pointer(&model)) - 0x10))
	entityID := attrs.Str("id")
	ex := baseEntManager.get(typName, entityID)
	if ex != nil {
		logger.Warn("has already load the entity", zap.String("typName", typName), zap.String("id", entityID))
		return ex, nil
	}

	return CreateEntity(typName, entityID, attrs, false), nil
}

// 加载BuildModel
func LoadTypeModelFromDB(typName string, param dbmgr.QueryPara) (interface{}, error) {
	typDesc := metapart.GetTypeDesc(typName)
	if typDesc == nil {
		logger.Warn("entity type not register", zap.String("typName", typName))
		// TODO 通知 加载entity 失败了
		return nil, define.ErrEntityNotRegister
	}

	_ = typDesc

	model := typDesc.Meta().Create()

	if err := dbmgr.Get(param, model); err != nil {
		// 是空数据的错误，不打印日志
		if err != dbmgr.ErrDataEmpty {
			logger.Warnf("load entity from db err", zap.String("typName", typName), zap.Any("query", param), zap.Error(err))
			return nil, define.ErrLoadDB
		}
		return nil, nil
	}

	return model, nil
}
