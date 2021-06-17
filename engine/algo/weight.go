package algo

import (
	"math/rand"
	"sort"
	"time"

	"github.com/tutumagi/pitaya/engine/utils"
)

// IWeight weight interface
type IWeight interface {
	Weight() float32
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

type _SortByWeightBox []_WeightBox

func (a _SortByWeightBox) Len() int           { return len(a) }
func (a _SortByWeightBox) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a _SortByWeightBox) Less(i, j int) bool { return a[i].weight < a[j].weight }

type _WeightBox struct {
	orignal IWeight
	weight  float32
}

// RandomWeightWithMaxWeightOnce 根据权重随机一次，这里指定 maxWeight，就是说可能会随机不到元素
// 比如有 权重  1，3，4， maxWeight 为 100， 则有很大可能（(100 - (1 + 3 + 4)) / 100 的可能性）随机不到元素
func RandomWeightWithMaxWeightOnce(items []IWeight, maxWeight float32) IWeight {
	res := _RandomWeight(items, maxWeight, 1)
	if len(res) < 1 {
		return nil
	}
	return _RandomWeight(items, maxWeight, 1)[0]
}

// RandomWeightOnce 根据权重随机一次，肯定会随机到元素
func RandomWeightOnce(items []IWeight) IWeight {
	return _RandomWeight(items, -1, 1)[0]
}

// RandomWeightWithMaxWeight 根据权重随机 total次，这里指定 maxWeight，就是说可能会随机不到元素
// 比如有 权重  1，3，4， maxWeight 为 100， 则有很大可能（(100 - (1 + 3 + 4)) / 100 的可能性）随机不到元素
func RandomWeightWithMaxWeight(items []IWeight, maxWeight float32, total int) []IWeight {
	return _RandomWeight(items, maxWeight, total)
}

// RandomWeight 根据权重随机 total 次
func RandomWeight(items []IWeight, total int) []IWeight {
	return _RandomWeight(items, -1, total)
}
func RandomWeightWithSingle(items []IWeight, total int) []IWeight {
	return _RandomWeightSingle(items, total)
}

//RandomWeightSingle 不重复随机 total 次
func RandomWeightSingle(items []IWeight, total int) []IWeight {
	return _RandomWeightSingle(items, total)
}
func _RandomWeightSingle(items []IWeight, total int) []IWeight {
	boxes := make(_SortByWeightBox, 0, len(items))
	curWeight := float32(0)
	for _, item := range items {
		curWeight += item.Weight()
		boxes = append(boxes, _WeightBox{
			orignal: item,
			weight:  curWeight,
		})
	}
	sort.Sort(boxes)
	result := make([]IWeight, 0, total)
	resultPool := make(map[int]struct{})
	indexPool := make([]float64, len(boxes), len(boxes))
	for i := 0; i < total; i++ {
		randomOnce := func() IWeight {
			r := rand.Float64() * float64(curWeight)
			for index, item := range boxes {
				_, exist := resultPool[index]
				if exist {
					continue
				}
				weight := float64(item.weight)
				if index > 0 {
					weight -= indexPool[index-1]
				}
				if r <= weight {
					resultPool[index] = struct{}{}
					curWeight -= item.orignal.Weight()
					indexPool[index] = float64(item.orignal.Weight())
					for preIndex, _ := range indexPool[index+1:] {
						indexPool[index+1+preIndex] += float64(item.orignal.Weight())
					}
					return item.orignal
				}
			}
			return nil
		}
		if weightRes := randomOnce(); weightRes != nil {
			result = append(result, weightRes)
		} else {
			break
		}
	}
	return result
}

// RandomWeight 权重随机 total 次，返回 [total]IWeight
func _RandomWeight(items []IWeight, maxWeight float32, total int) []IWeight {
	boxes := make(_SortByWeightBox, 0, len(items))
	curWeight := float32(0)
	for _, item := range items {
		curWeight += item.Weight()
		boxes = append(boxes, _WeightBox{
			orignal: item,
			weight:  curWeight,
		})
	}

	sort.Sort(boxes)

	if maxWeight != -1 && maxWeight > curWeight {
		curWeight = maxWeight
	}

	result := make([]IWeight, 0, total)
	for i := 0; i < int(total); i++ {
		randomOnce := func() IWeight {
			r := rand.Float32() * curWeight
			for _, item := range boxes {
				if r <= item.weight {
					return item.orignal
				}
			}
			return nil
		}
		if weightRes := randomOnce(); weightRes != nil {
			result = append(result, weightRes)
		}
	}

	// logger.Info("random resource", zap.Int("count", count), zap.Int("total", total))
	return result
}

//MixArray 打乱数组
func MixArray(items []IWeight) []IWeight {
	if len(items) < 1 {
		return items
	}
	for i := 0; i < len(items); i++ {
		idx := int(utils.RandomInt32(int32(i), int32(len(items))))
		items[i], items[idx] = items[idx], items[i]
	}
	return items
}
