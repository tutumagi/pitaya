package metapart

import (
	"reflect"

	"github.com/tutumagi/pitaya/logger"

	"gitlab.gamesword.com/nut/entitygen/attr"
)

// 每个实体都有自己的实体ID，角色的实体ID就是角色ID
type _Empty struct {
	ID string `bson:"id" json:"id"`
}

// TypeDesc 实体类型信息
type TypeDesc struct {
	name         string
	isPersistent bool

	aoiRadius float32
	useAOI    bool
	// entity type
	baseTyp reflect.Type
	cellTyp reflect.Type
	// model type
	// mTyp reflect.Type

	interestTypNames []string

	meta *attr.Meta

	Routers *Routers
}

// SetUseAOI 设置该实体类型是否要用到AOI
func (desc *TypeDesc) SetUseAOI(use bool, radius float32) *TypeDesc {
	if radius < 0 {
		logger.Warn("aoi distance < 0, fix it to 0")
		radius = 0
	}
	desc.aoiRadius = radius
	desc.useAOI = use

	return desc
}

func (desc *TypeDesc) UseAOI() bool {
	return desc.useAOI
}

func (desc *TypeDesc) AOIDistance() float32 {
	return desc.aoiRadius
}

func (desc *TypeDesc) SetMeta(meta *attr.Meta) {
	if desc.meta != nil {
		logger.Warnf("%s has already have meta", desc.name)
		return
	}
	desc.meta = meta
}

func (desc *TypeDesc) Meta() *attr.Meta {
	return desc.meta
}

func (desc *TypeDesc) DefAttr(key string, typ attr.AttrTyp, flag attr.AttrFlag, storeDB bool) {
	desc.meta.DefAttr(key, typ, flag, storeDB)
}

// func (desc *TypeDesc) DynamicStruct() interface{} {
// 	return desc.builder().New()
// }

func (desc *TypeDesc) DynamicSliceOfStruct() interface{} {
	return desc.meta.DynamicSliceOfStruct()
}

// func (desc *TypeDesc) builder() dynamicstruct.DynamicStruct {
// 	if desc.dynStruct == nil {
// 		builder := dynamicstruct.ExtendStruct(_Empty{})
// 		for k, v := range desc.attrsDef {
// 			tagStr := "-"
// 			if v.storeDB {
// 				tagStr = k
// 			}
// 			// Field的 name 必须是大写开头的，因为go语言 反射必须是外部包可见的field
// 			// 写到db是使用的 bson， json是内存中 marshal unmarshal使用的，所以json不忽略，
// 			// 当不需要存储到db时，bson 忽略，使用 `-` tag
// 			builder.AddField(
// 				strings.Title(k), // 首字母大写
// 				v.typ,
// 				fmt.Sprintf(`json:"%s" bson:"%s"`, k, tagStr),
// 			)
// 		}

// 		desc.dynStruct = builder.Build()
// 	}
// 	return desc.dynStruct
// }

// 通过 dynamicStruct 解析到的struct，转为 map[string]interface{}
// func (desc *TypeDesc) unmarshal(srcStruct interface{}) map[string]interface{} {
// 	return desc.meta.Unmarshal(srcStruct)
// }

// // 通过 dynamicStruct 解析到的struct，转为 map[string]interface{}
// func (desc *TypeDesc) unmarshalSlice(srcStruct interface{}) []map[string]interface{} {
// 	var attrs = []map[string]interface{}{}
// 	readers := dynamicstruct.NewReader(srcStruct).ToSliceOfReaders()
// 	for _, r := range readers {
// 		attrs = append(attrs, desc.readerToMap(r))
// 	}

// 	return attrs
// }

// // 将 dynamicstruct.Reader 转为 map[string]interface{}
// func (desc *TypeDesc) readerToMap(r dynamicstruct.Reader) map[string]interface{} {
// 	var attrs = map[string]interface{}{}
// 	for _, field := range r.GetAllFields() {
// 		name := strings.ToLower(field.Name()) // TODO 这里有性能瓶颈，可以考虑 修改dynamicstruct 的源码，去缓存这个 小写开头的字符串
// 		attrs[name] = field.Interface()
// 	}

// 	return attrs
// }

func (desc *TypeDesc) UnmarshalBson(bytes []byte) (*attr.StrMap, error) {
	return desc.meta.UnmarshalBson(bytes, nil)
}

func (desc *TypeDesc) UnmarshalJSON(bytes []byte) (*attr.StrMap, error) {
	return desc.meta.UnmarshalJson(bytes, nil)
}

func (desc *TypeDesc) TypName() string {
	return desc.name
}

func (desc *TypeDesc) IsPersistent() bool {
	return desc.isPersistent
}

func (desc *TypeDesc) GetDef(k string) *attr.FieldDef {
	return desc.meta.GetDef(k)
}

func (desc *TypeDesc) CellTyp() reflect.Type {
	return desc.cellTyp
}

// func (desc *TypeDesc) NewCellPartInterface() interface{} {
// 	entityInstance := reflect.New(desc.cellTyp)
// 	return reflect.Indirect(entityInstance).FieldByName("CellEntity").Addr().Interface()
// }

func (desc *TypeDesc) BaseTyp() reflect.Type {
	return desc.baseTyp
}

// func (desc *TypeDesc) kind() string {
// 	if strings.Contains(desc.name, consts.ServiceEntityType) {
// 		return route.ServiceKind
// 	} else if strings.Contains(desc.name, consts.SpaceEntityType) {
// 		return route.SpaceKind
// 	} else {
// 		return route.EntityKind
// 	}
// }

// func (desc *TypeDesc) NewBasePartInterface() interface{} {
// 	entityInstance := reflect.New(desc.cellTyp)
// 	return reflect.Indirect(entityInstance).FieldByName("BaseEntity").Addr().Interface()
// }

// // SetInterestTypName 定义该类型 关心哪些类型
// func (desc *TypeDesc) SetInterestTypName(typeNames ...string) {
// 	if len(desc.interestTypNames) > 0 {
// 		logger.Warn("has already define interest type name", zap.String("name", desc.name), zap.Any("typNames", typeNames))
// 		return
// 	}
// 	for _, typName := range typeNames {
// 		f := aoi.FlagFromType(typName)
// 		desc.aoiInterestFlag |= f
// 	}
// 	desc.interestTypNames = typeNames
// }

// // SetInterestFlag 直接设置关心flag
// func (desc *TypeDesc) SetInterestFlag(f aoi.Flag) {
// 	desc.aoiInterestFlag |= f
// }

var registerEntityTypes = map[string]*TypeDesc{}

// var registerEntityTypesRWLock = &sync.RWMutex{}

// GetTypeDesc 获取实体类型信息
func GetTypeDesc(typName string) *TypeDesc {
	return registerEntityTypes[typName]
}

func AddTypDesc(
	typName string,
	baseTyp reflect.Type,
	cellTyp reflect.Type,
	persistent bool,
) *TypeDesc {

	entityVal := reflect.ValueOf(baseTyp)
	entityType := entityVal.Type()

	if entityType.Kind() == reflect.Ptr {
		entityType = entityType.Elem()
	}

	typeDesc := &TypeDesc{
		name:         typName,
		isPersistent: persistent,
		// aoiFlag:         aoi.FlagFromType(typName),
		// aoiInterestFlag: aoi.NoneFlag,
		aoiRadius: 0,
		baseTyp:   baseTyp,
		cellTyp:   cellTyp,

		// meta: &attr.Meta{},
	}

	registerEntityTypes[typName] = typeDesc

	return typeDesc
}
