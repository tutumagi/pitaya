package bc

import (
	"encoding/json"
	"reflect"
	"testing"

	. "github.com/go-playground/assert/v2"
	"github.com/tutumagi/pitaya/engine/bc/basepart"
	"github.com/tutumagi/pitaya/engine/bc/cellpart"
	"github.com/tutumagi/pitaya/engine/bc/metapart"
	"gitlab.gamesword.com/nut/entitygen/attr"
)

// import (
// 	"encoding/json"
// 	"reflect"
// 	"testing"

// 	. "github.com/go-playground/assert/v2"
// )

type CustomField1 struct {
	Name string
	Age  int32
}

type CustomField2 struct {
	FF *CustomField1
}

const mockAttrEntityLabel = "mockAttrEntityLabel"

type MockAttrEntity struct {
	basepart.Entity
}

type MockAttrCellEntity struct {
	cellpart.Entity
}

func (m *MockAttrCellEntity) CellAttrChanged(keys map[string]struct{}) {}

func regEntity() {
	if metapart.GetTypeDesc(mockAttrEntityLabel) != nil {
		return
	}

	typDesc := RegisterEntity(mockAttrEntityLabel, &MockAttrEntity{}, &MockAttrCellEntity{}, false)
	typDesc.SetMeta(attr.NewMeta(func() interface{} {
		return &MockAttrEntity{}
	}, func() interface{} {
		return &[]*MockAttrEntity{}
	}))

	typDesc.DefAttr("Int", attr.Int, attr.AfBase, false)
	typDesc.DefAttr("UInt", attr.UInt, attr.AfBase, false)
	typDesc.DefAttr("Int8", attr.Int8, attr.AfBase, false)
	typDesc.DefAttr("Int16", attr.Int16, attr.AfBase, false)
	typDesc.DefAttr("Int32", attr.Int32, attr.AfBase, false)
	typDesc.DefAttr("Int64", attr.Int64, attr.AfBase, false)
	typDesc.DefAttr("UInt8", attr.Uint8, attr.AfBase, false)
	typDesc.DefAttr("UInt16", attr.Uint16, attr.AfBase, false)
	typDesc.DefAttr("UInt32", attr.Uint32, attr.AfBase, false)
	typDesc.DefAttr("UInt64", attr.Uint64, attr.AfBase, false)

	typDesc.DefAttr("Float32", attr.Float32, attr.AfBase, false)
	typDesc.DefAttr("Float64", attr.Float64, attr.AfBase, false)

	typDesc.DefAttr("Bool", attr.Bool, attr.AfBase, false)

	typDesc.DefAttr("String", attr.String, attr.AfBase, false)

	typDesc.DefAttr("MapStrStr", map[string]string{}, attr.AfBase, false)
	typDesc.DefAttr("MapInt32Str", map[int32]string{}, attr.AfBase, false)
	typDesc.DefAttr("MapInt32Int32", map[int32]int32{}, attr.AfBase, false)
	typDesc.DefAttr("MapStrInt32", map[string]int32{}, attr.AfBase, false)

	typDesc.DefAttr("SliceStr", []string{}, attr.AfBase, false)
	typDesc.DefAttr("SliceInt32", []int32{}, attr.AfBase, false)

	typDesc.DefAttr("Custom1", &CustomField1{}, attr.AfBase, false)
	typDesc.DefAttr("Custom2", &CustomField2{}, attr.AfBase, false)
}

func newExpectAttrMap() *attr.StrMap {
	expectData := map[string]interface{}{
		"Int":    attr.Int,
		"UInt":   attr.UInt,
		"Int8":   attr.Int8,
		"Int16":  attr.Int16,
		"Int32":  attr.Int32,
		"Int64":  attr.Int64,
		"UInt8":  attr.Uint8,
		"UInt16": attr.Uint16,
		"UInt32": attr.Uint32,
		"UInt64": attr.Uint64,

		"Float32": attr.Float32,
		"Float64": attr.Float64,

		"Bool": attr.Bool,

		"String": attr.String,

		"MapStrStr":     map[string]string{},
		"MapInt32Str":   map[int32]string{},
		"MapInt32Int32": map[int32]int32{},
		"MapStrInt32":   map[string]int32{},

		"SliceStr":   []string{},
		"SliceInt32": []int32{},

		"Custom1": &CustomField1{
			Name: "gege",
			Age:  66,
		},
		"Custom2": &CustomField2{
			FF: &CustomField1{
				Name: "lili",
				Age:  88,
			},
		},
	}

	attrMap := attr.NewStrMap(nil)
	attrMap.SetData(expectData)

	return attrMap
}

// 只修改基础类型 attrmap
func modifyedPrimaryTypeAttrMap() *attr.StrMap {
	attrMap := attr.NewStrMap(nil)

	attrMap.Set("Int", int(10))
	attrMap.Set("UInt", uint(20))
	attrMap.Set("Int8", int8(30))
	attrMap.Set("Int16", int16(40))
	attrMap.Set("Int32", int32(50))
	attrMap.Set("Int64", int64(60))
	attrMap.Set("UInt8", uint8(70))
	attrMap.Set("UInt16", uint16(80))
	attrMap.Set("UInt32", uint32(90))
	attrMap.Set("UInt64", uint64(100))

	attrMap.Set("Float32", float32(110))
	attrMap.Set("Float64", float64(120))

	attrMap.Set("Bool", true)

	attrMap.Set("String", "hello world")

	return attrMap
}

func modifyMapKeyAttrMap() *attr.StrMap {
	attrMap := attr.NewStrMap(nil)

	attrMap.Set("MapStrStr", map[string]string{
		"hello": "world",
	})
	attrMap.Set("MapInt32Str", map[int32]string{
		101: "happyness",
	})
	attrMap.Set("MapInt32Int32", map[int32]int32{
		985: 996,
	})
	attrMap.Set("MapStrInt32", map[string]int32{
		"zune": 886,
	})

	return attrMap
}

func modifySliceKeyAttrMap() *attr.StrMap {
	attrMap := attr.NewStrMap(nil)

	attrMap.Set("SliceStr", []string{
		"hello",
		"world",
	})
	attrMap.Set("SliceInt32", []int32{
		101,
		103,
	})

	return attrMap
}

func modifyCustomKeyAttrMap() *attr.StrMap {
	attrMap := attr.NewStrMap(nil)

	attrMap.Set("Custom1", &CustomField1{
		Name: "tufei",
		Age:  100,
	})
	attrMap.Set("Custom2", &CustomField2{
		FF: &CustomField1{
			Name: "zhijun",
			Age:  500,
		},
	})

	return attrMap
}

func TestKeyChange(t *testing.T) {
	t.Run("primary_key", func(t *testing.T) {
		testWithModifyAttrMap(t, modifyedPrimaryTypeAttrMap())
	})
	t.Run("map_key", func(t *testing.T) {
		testWithModifyAttrMap(t, modifyMapKeyAttrMap())
	})
	t.Run("slice_key", func(t *testing.T) {
		testWithModifyAttrMap(t, modifySliceKeyAttrMap())
	})
	t.Run("custom_key", func(t *testing.T) {
		testWithModifyAttrMap(t, modifyCustomKeyAttrMap())
	})
}

func TestAttr(t *testing.T) {
	regEntity()

	expectData := newExpectAttrMap()

	e := basepart.CreateEntity(mockAttrEntityLabel, "", expectData, false)
	NotEqual(t, e, nil)

	mockEntity := e.Val().(*MockAttrEntity)
	NotEqual(t, mockEntity, nil)

	migrateData := mockEntity.GetMigrateData()
	for k, v := range migrateData {
		Equal(t, v, expectData.Value(k))
	}

	t.Run("json", func(t *testing.T) {
		t.Run("total-marshal", func(t *testing.T) {
			migrateBytes, err := json.Marshal(migrateData)
			Equal(t, err, nil)
			expectBytes, err := json.Marshal(expectData.Data())
			Equal(t, err, nil)
			Equal(t, expectBytes, migrateBytes)
		})
	})
}

func testWithModifyAttrMap(t *testing.T, modifydAttrMap *attr.StrMap) {
	regEntity()

	oldAttrMap := newExpectAttrMap()

	oldMockEntity := basepart.CreateEntity(mockAttrEntityLabel, "", oldAttrMap, false).Val().(*MockAttrEntity)
	newMockEntity := basepart.CreateEntity(mockAttrEntityLabel, "", oldAttrMap, false).Val().(*MockAttrEntity)

	NotEqual(t, oldMockEntity, nil)
	NotEqual(t, newMockEntity, nil)

	t.Run("before", func(t *testing.T) {
		// 修改之前检查两个 实体的属性是否一致
		checkEntityAttrEqual(t, &newMockEntity.Entity, &oldMockEntity.Entity)
	})

	// 修改旧的实体的 key
	for k, v := range modifydAttrMap.Data() {
		oldMockEntity.Data().Set(k, v)
	}
	t.Logf("change keys :%v", oldMockEntity.Data().ChangeKey())
	// 打包旧的实体的key
	changeBytes := oldMockEntity.MarshalChangedKey(oldMockEntity.Data().ChangeKey())

	// 解包这些变化的二进制数据到新的实体上
	newMockEntity.UnmarshalChangedKey(changeBytes)

	t.Run("after", func(t *testing.T) {
		// 检查修改后 两个实体是否一致
		checkEntityAttrEqual(t, &newMockEntity.Entity, &oldMockEntity.Entity)
	})
}

func checkEntityAttrEqual(t *testing.T, newEntity *basepart.Entity, oldEntity *basepart.Entity) {
	for k, v := range oldEntity.Data().Data() {
		entityValue := newEntity.Data().Value(k)
		expectValue := v
		t.Logf("check key %s entityValue:%v type:%s ;expectValue:%v type:%s",
			k,
			entityValue,
			reflect.TypeOf(entityValue).String(),
			expectValue,
			reflect.TypeOf(expectValue).String(),
		)
		EqualSkip(t, 2, newEntity.Data().Value(k), v)
	}
}

// // func checkEntityAttr(t *testing.T, ent *Entity, expectAttrMap *AttrMap) {
// // 	for k, v := range expectAttrMap.attrs {
// // 		entityValue := ent.data.attrs[k]
// // 		expectValue := v
// // 		t.Logf("check key %s entityValue:%v type:%s ;expectValue:%v type:%s",
// // 			k,
// // 			entityValue,
// // 			reflect.TypeOf(entityValue).String(),
// // 			expectValue,
// // 			reflect.TypeOf(expectValue).String(),
// // 		)
// // 		EqualSkip(t, 2, ent.data.attrs[k], v)
// // 	}
// // }
