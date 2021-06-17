package consts

import (
	"fmt"
	"time"
)

// 常量定义
const (
	MigrateTimeOut           = 1 * time.Minute
	EnterSpaceRequestTimeout = MigrateTimeOut + 1*time.Minute

	SpaceEntityType   = "__space__"
	ServiceEntityType = "__service__"
)

func SpaceTypeName(kind int32) string {
	return fmt.Sprintf("%s%d", SpaceEntityType, kind)
}

func ServiceTypeName(typName string) string {
	return ServiceEntityType + typName
}
