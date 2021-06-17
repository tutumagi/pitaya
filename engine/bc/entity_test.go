package bc

// . "github.com/go-playground/assert/v2"

// type TestHouse struct {
// 	Price int32 `json:"price" bson:"price"`
// }

// type TestPeople struct {
// 	Name  string    `json:"name" bson:"name"`
// 	Age   int32     `json:"age" bson:"age"`
// 	House TestHouse `json:"house" bson:"house"`
// }

// const testEntityTypName = "test_people"

// type TestPeopleEntity struct {
// 	Entity
// }

// func (te *TestPeopleEntity) Tick(dt int32)                            {}
// func (te *TestPeopleEntity) CellAttrChanged(keys map[string]struct{}) {}

// func InitEnv() {
// 	v := viper.New()
// 	v.Set("mongo.host", "0.0.0.0")
// 	v.Set("mongo.port", 27017)
// 	v.Set("mongo.auth", false)
// 	v.Set("mongo.dbname", "for_test")
// 	v.Set("logger.level", "debug")

// 	dbmgr.InitDbHandler(v)

// 	time.Sleep(100 * time.Millisecond)
// }

// func Test_EntityLoadSave(t *testing.T) {
// 	InitEnv()
// 	Initialize("Server", true)

// 	typDesc := RegisterEntity(testEntityTypName, &TestPeopleEntity{}, true)
// 	typDesc.SetMeta(attr.NewMeta(func() interface{} {
// 		return &TestPeople{}
// 	}, func() interface{} {
// 		return &[]*TestPeople{}
// 	}))
// 	typDesc.DefAttr("name", attr.String, attr.AfBaseAndCell, true)
// 	typDesc.DefAttr("age", attr.Int32, attr.AfBaseAndCell, true)
// 	typDesc.DefAttr("house", TestHouse{}, attr.AfBaseAndCell, true)

// 	model := &TestPeople{Name: "tufei", Age: 10, House: TestHouse{Price: 99}}
// 	e := CreateEntity(testEntityTypName, "", attr.A{
// 		"name":  model.Name,
// 		"age":   model.Age,
// 		"house": model.House,
// 	}, true)

// 	Equal(t, e.data.Str("name"), model.Name)
// 	Equal(t, e.data.Int32("age"), model.Age)
// 	Equal(t, e.data.Value("house").(TestHouse).Price, model.House.Price)

// 	e.Save()

// 	// 等待db写入完成
// 	time.Sleep(100 * time.Millisecond)

// 	e.Destroy()

// 	dbEntity, err := LoadEntity(testEntityTypName, e.ID)

// 	Equal(t, err, nil)

// 	Equal(t, dbEntity.data.Str("name"), model.Name)
// 	Equal(t, dbEntity.data.Int32("age"), model.Age)
// 	Equal(t, dbEntity.data.Value("house").(TestHouse).Price, model.House.Price)
// }
