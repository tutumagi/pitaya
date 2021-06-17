package bc

import (
	"reflect"

	"github.com/tutumagi/pitaya/engine/bc/basepart"
	"github.com/tutumagi/pitaya/engine/bc/cellpart"
	"github.com/tutumagi/pitaya/engine/bc/internal/consts"
	"github.com/tutumagi/pitaya/engine/bc/metapart"
	"github.com/tutumagi/pitaya/logger"
	"go.uber.org/zap"
)

type emptyBaseEntity struct {
	basepart.Entity
}
type emptyCellEntity struct {
	cellpart.Entity
}

func (e *emptyCellEntity) CellAttrChanged(keys map[string]struct{}) {}

type emptyBaseSpace struct {
	basepart.Space
}
type emptyCellSpace struct {
	cellpart.Space
}

var _ basepart.IBaseEntity = &emptyBaseEntity{}
var _ cellpart.ICellEntity = &emptyCellEntity{}
var _ basepart.ISpace = &emptyBaseSpace{}
var _ cellpart.ISpace = &emptyCellSpace{}

// var defaultBasePartType = reflect.TypeOf(emptyBaseEntity{})
// var defaultCellPartType = reflect.TypeOf(emptyCellEntity{})

// RegisterEntity 注册实体
func RegisterEntity(
	typName string,
	baseEntityPtr basepart.IBaseEntity,
	cellEntityPtr cellpart.ICellEntity,
	persistent bool,
) *metapart.TypeDesc {
	if td := metapart.GetTypeDesc(typName); td != nil {
		// TODO  这里要考虑下怎么规避不要重复注册
		logger.Warn("entity type already register", zap.String("name", typName))
		return td
	}

	if baseEntityPtr == nil {
		baseEntityPtr = &emptyBaseEntity{}
	}

	if cellEntityPtr == nil {
		cellEntityPtr = &emptyCellEntity{}
	}

	baseEntityType := reflect.ValueOf(baseEntityPtr).Type()
	if baseEntityType.Kind() == reflect.Ptr {
		baseEntityType = baseEntityType.Elem()
	}

	cellEntityType := reflect.ValueOf(cellEntityPtr).Type()
	if cellEntityType.Kind() == reflect.Ptr {
		cellEntityType = cellEntityType.Elem()
	}

	return metapart.AddTypDesc(typName, baseEntityType, cellEntityType, persistent)
}

// RegisterSpace register custom space
func RegisterSpace(kind int32, spacePtrBase basepart.ISpace, spacePtrCell cellpart.ISpace) *metapart.TypeDesc {
	// spaceVal := reflect.Indirect(reflect.ValueOf(spacePtr))
	// spaceType = spaceVal.Type()

	if spacePtrBase == nil {
		spacePtrBase = &emptyBaseSpace{}
	}

	if spacePtrCell == nil {
		spacePtrCell = &emptyCellSpace{}
	}

	return RegisterEntity(consts.SpaceTypeName(kind), spacePtrBase, spacePtrCell, false)
}

func RegisterService(typName string, entityPtr basepart.IBaseEntity) *metapart.TypeDesc {
	return RegisterEntity(consts.ServiceTypeName(typName), entityPtr, nil, false)
}

// 生成一个实体ID
func NewEntityID() string {
	return metapart.NewUUID()
}
