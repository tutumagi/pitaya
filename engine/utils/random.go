package utils

import (
	"math"
	"math/rand"
)

// RandomFloat32 rand float32 in [min, max)
func RandomFloat32(min, max float32) float32 {
	return min + rand.Float32()*(max-min)
}

// RandomUInt32 rand Uint32 in [min, max)
func RandomUInt32(min, max uint32) uint32 {
	return min + uint32(math.Floor(float64(rand.Float32()*float32(max-min))))
}

// RandomInt32 rand int32 in [min, max)
func RandomInt32(min, max int32) int32 {
	if min == max {
		return min
	}
	return min + rand.Int31n(max-min)
}

// UnifyRandomUInt32 前后统一的随机数
func UnifyRandomUInt32(seed int32, maxNum int32) int32 {
	return int32(math.Abs(math.Floor(float64((seed*1103515245+12345)/65536)))) % 32768 % maxNum
}
