package algo

//根据数组下标随机权重
type NumberRange struct {
	ItemIndex int32
	WeightVal float32
}

func (a *NumberRange) Weight() float32 {
	return a.WeightVal
}
func InitNumberRange(length int) []IWeight {
	rangeList := make([]IWeight, 0, length)
	for i := 0; i < length; i++ {
		rangeList = append(rangeList, &NumberRange{
			ItemIndex: int32(i),
			WeightVal: 1,
		})
	}
	return rangeList
}
