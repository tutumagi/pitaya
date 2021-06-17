package utils

import "math"

// FloatEqualLow low precision equal
func FloatEqualLow(l float32, r float32) bool {
	if math.Abs(float64(l-r)) <= 0.01 {
		return true
	} else {
		return false
	}
}

// FloatEqualHigh high precision equal
func FloatEqualHigh(l float32, r float32) bool {
	if math.Abs(float64(l-r)) <= 0.0001 {
		return true
	} else {
		return false
	}
}

//MinInt64 return min
func MinInt64(a, b int64) int64 {
	if a > b {
		return b
	}
	return a
}

//MaxInt32 return max
func MaxInt32(a, b int32) int32 {
	if a > b {
		return a
	}
	return b
}

//MaxInt return max
func MaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

//MinInt32
func MinInt32(a, b int32) int32 {
	if a > b {
		return b
	}
	return a
}

func ListFind(list []int32, val int32) bool {
	left, right := 0, len(list)-1
	for left <= right {
		if list[left] == val || list[right] == val {
			return true
		}
		left, right = left+1, right-1
	}
	return false
}

func IntListFind(list []int, val int) bool {
	left, right := 0, len(list)-1
	for left <= right {
		if list[left] == val || list[right] == val {
			return true
		}
		left, right = left+1, right-1
	}
	return false
}

//Int32ListSum 相加
func Int32ListSum(list []int32) int32 {
	res := int32(0)
	for _, x := range list {
		res += x
	}
	return res
}
