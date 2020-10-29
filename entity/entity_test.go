package entity

import (
	"testing"
	"time"

	"github.com/tutumagi/pitaya/math32"

	. "github.com/go-playground/assert/v2"
)

type TestHouse struct {
	Price int32 `json:"price"`
}

type TestPeople struct {
	Name  string    `json:"name"`
	Age   int32     `bson:"age"`
	House TestHouse `bson:"house"`
}

const testEntityTypName = "test_people"

type TestPeopleEntity struct {
	Entity
}

func (te *TestPeopleEntity) Model() *TestPeople {
	return te.Data.(*TestPeople)
}

func (te *TestPeopleEntity) DefaultModel(id string) interface{} {
	return &TestPeople{}
}

func InitEnv() {
	// v := viper.New()
	// v.Set("mongo.host", "0.0.0.0")
	// v.Set("mongo.port", 27017)
	// v.Set("mongo.useauth", false)
	// v.Set("env", "dev")
	// v.Set("mongo.dbname", "for_test")

	// cfg := config.NewConfig(v)

	// logger.Init(cfg)

	// dbmgr.InitDbHandler(v)

	// time.Sleep(100 * time.Millisecond)
}

func Test_EntityLoadSave(t *testing.T) {
	InitEnv()

	RegisterEntity(testEntityTypName, &TestPeopleEntity{}, &TestPeople{}, true)

	model := &TestPeople{Name: "tufei", Age: 10, House: TestHouse{Price: 99}}
	e := createEntity(testEntityTypName, "", model, nil, math32.Vector3{}, true)

	Equal(t, e.Data, model)

	e.Save()

	// 等待db写入完成
	time.Sleep(100 * time.Millisecond)

	e.Destroy()

	dbEntity := loadEntity(testEntityTypName, e.ID, nil)

	Equal(t, dbEntity.Data, model)
}
