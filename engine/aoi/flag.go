package aoi

import "github.com/tutumagi/pitaya/logger"

// Flag 实体 flag
// 最多支持 (maxBit - 2)种实体，0用来表示任何实体都不关心，比如树木，石头，maxBit 表示关心所有的实体，比如玩家
// uint32 最多只是 (32-2) = 30种实体
type Flag uint32

const maxFlagCount = 30

// aoi 关注/被关注 flag
const (
	NoneFlag Flag = 0

	// 关心所有的
	maxFlag Flag = 0xFFFFFFFF
)

// InterestAllFlag 关心所有的实体的flag
const InterestAllFlag = maxFlag

// Valid aoi flag 是否合法
func (f Flag) Valid() bool {
	return f <= maxFlag
}

var (
	typeFlagMap map[string]Flag = make(map[string]Flag, maxFlagCount)
	curFlagBit                  = 0
)

// FlagFromType 根据实体类型获取该实体的flag
func FlagFromType(typName string) Flag {
	f, ok := typeFlagMap[typName]
	if ok {
		return f
	}
	if len(typeFlagMap) >= maxFlagCount {
		logger.Panic("太多实体类型了，请修改 flag 类型")
	}
	f = 1 << curFlagBit
	typeFlagMap[typName] = f
	curFlagBit++

	return f
}
