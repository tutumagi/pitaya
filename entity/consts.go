package entity

import "time"

// 常量定义
const (
	MigrateTimeOut           = 1 * time.Minute
	EnterSpaceRequestTimeout = MigrateTimeOut + 1*time.Minute
)
