package entity

import "github.com/tutumagi/pitaya/math32"

// MigrateData 实体的迁移数据
type MigrateData struct {
	// 实体ID
	ID string `json:"i"`
	// 实体typName
	TypName string `json:"t"`
	// 实体序列化后的数据，目前使用json
	DataBytes []byte `json:"b"`
	// 实体位置
	Pos math32.Vector3 `json:"p"`
	// 实体朝向
	Yaw float32 `json:"y"`
	// 实体所在space
	SpaceID string `json:"s"`
	// TimerData 定时器数据？
}
