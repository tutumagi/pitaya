package entity

import (
	"github.com/tutumagi/pitaya/logger"

	"reflect"
)

// TypeDesc 实体类型信息
type TypeDesc struct {
	IsPersistent bool
	useAOI       bool
	aoiDistance  int32
	// entity type
	eTyp reflect.Type
	// model type
	mTyp reflect.Type
}

// SetPersistent 设置该实体类型描述是否需要持久化
func (desc *TypeDesc) SetPersistent(persistent bool) *TypeDesc {
	desc.IsPersistent = persistent

	return desc
}

// SetUseAOI 设置该实体类型是否要用到AOI
func (desc *TypeDesc) SetUseAOI(use bool, distance int32) *TypeDesc {
	if distance < 0 {
		logger.Log.Warn("aoi distance < 0")
	}

	desc.useAOI = use
	desc.aoiDistance = distance

	return desc
}
