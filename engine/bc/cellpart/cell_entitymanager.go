package cellpart

import (
	"reflect"

	"github.com/tutumagi/pitaya/engine/bc/metapart"
	"github.com/tutumagi/pitaya/logger"
	"gitlab.gamesword.com/nut/entitygen/attr"
	"go.uber.org/zap"
)

type _CellEntityManager struct {
	entitiesByType map[string]CellMap
}

func newCellEntityManager() *_CellEntityManager {
	return &_CellEntityManager{
		entitiesByType: map[string]CellMap{},
	}
}

func (em *_CellEntityManager) put(entity *Entity) {
	typName := entity.typeDesc.TypName()
	if em.entitiesByType[typName] == nil {
		em.entitiesByType[typName] = make(CellMap, 1000)
	}
	m, ok := em.entitiesByType[typName]
	if ok {
		m.Add(entity)
	} else {
		m = make(CellMap, 1000)
		m.Add(entity)
		em.entitiesByType[typName] = m
	}
}

func (em *_CellEntityManager) del(e *Entity) {
	if m, ok := em.entitiesByType[e.typeDesc.TypName()]; ok {
		m.Del(e.ID)
	}
}

func (em *_CellEntityManager) get(typName string, id string) *Entity {
	if m, ok := em.entitiesByType[typName]; ok {
		return m.Get(id)
	}
	return nil
}

func (em *_CellEntityManager) getEntitiesByTypName(typName string) CellMap {
	return em.entitiesByType[typName]
}

func (em *_CellEntityManager) foreach(fn func(*Entity) bool) {
	for _, es := range em.entitiesByType {
		for _, e := range es {
			if !fn(e) {
				goto exit
			}
		}
	}

exit:
}

var cellEntManager = newCellEntityManager()

// GetEntity get entity
func GetEntity(typName string, id string) *Entity {
	return cellEntManager.get(typName, id)
}

// GetEntitiesByType 根据实体类型获取实体map
func GetEntitiesByType(typName string) CellMap {
	return cellEntManager.getEntitiesByTypName(typName)
}

// 遍历所有实体，如果返回false，则停止遍历
func ForEach(fn func(e *Entity) bool) {
	cellEntManager.foreach(fn)
}

func CreateEntity(
	label string,
	id string,
	// data map[string]interface{},
	data *attr.StrMap,
	// shouldSaveDB bool,
) *Entity {
	entity := createCellEntityOnlyInit(id, label)
	err := entity.I.OnInit()
	// TODO 这里要返回错误码
	if err != nil {
		return nil
	}

	cellEntManager.put(entity)

	// entity.data = attr.NewStrMap(data)
	if data == nil {
		entity.data = attr.NewStrMap(nil)
	} else {
		entity.data = data
	}

	entity.data.Set("id", entity.ID)
	entity.data.ClearChangeKey()

	// if shouldSaveDB {
	// 	entity.saveToDB()
	// }

	if err != nil {
		logger.Warnf("create entity failed. id:%s label:%s shouldSaveDB:%s err:%s",
			id, label, data, false,
		)
		return nil
	}

	entity.I.OnCreated()

	return entity
}

// createBaseEntityOnlyInit 创建实体
//	 isCreate 表示是否是新建的实体（DB里面没有的）
func createCellEntityOnlyInit(
	id string,
	label string,
) *Entity {
	typeDesc := metapart.GetTypeDesc(label)
	if typeDesc == nil {
		logger.Panic("unknown entity type", zap.String("name", label))
	}

	if id == "" {
		id = metapart.NewUUID()
	}

	var entity *Entity
	var entityInstance reflect.Value

	entityInstance = reflect.New(typeDesc.CellTyp())
	entity = reflect.Indirect(entityInstance).FieldByName("Entity").Addr().Interface().(*Entity)

	entity.init(id, label, entityInstance)

	return entity
}
