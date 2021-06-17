package algo

import (
	"fmt"
	"math"
	"testing"
)

type _WeightItem struct {
	icon   string
	weight float32
}

func (w _WeightItem) Weight() float32 {
	return w.weight
}

var ItemSquare _WeightItem = _WeightItem{
	icon:   "◻︎",
	weight: 0.01,
}
var ItemTriangle _WeightItem = _WeightItem{
	icon:   "△",
	weight: 0.01,
}
var ItemCircle _WeightItem = _WeightItem{
	icon:   "○",
	weight: 0.01,
}
var ItemLove _WeightItem = _WeightItem{
	icon:   "❤︎",
	weight: 1000,
}
var ItemPentagram _WeightItem = _WeightItem{
	icon:   "★",
	weight: 0.01,
}
var ItemSnow _WeightItem = _WeightItem{
	icon:   "❈",
	weight: 0.01,
}

var TestItems []_WeightItem = []_WeightItem{
	ItemPentagram, ItemSnow, ItemLove, ItemCircle, ItemTriangle, ItemSquare,
}

func Test_Weight(t *testing.T) {
	const column = 100
	const row = 100
	var total = column * row

	t.Run("MaxWeightOnce", func(t *testing.T) {
		var weightStat map[_WeightItem]int32 = maxWeightOnce(row, column, true)
		var totalWeight float32
		for _, x := range TestItems {
			totalWeight += x.Weight()
		}
		for weightItem, stat := range weightStat {
			ratio := float32(stat) / float32(total)
			fmt.Printf("weight item %+v stat %v ratio %.4f\n", weightItem, stat, ratio)
			if !floatEqualLow(ratio, weightItem.weight/totalWeight) {
				t.Error("权重计算失败，请开启 -v 查看详细输出")
			}
		}
	})

	t.Run("MaxWeight", func(t *testing.T) {

		var weightStat map[_WeightItem]int32 = maxWeight(row, column, true)
		var totalWeight float32
		for _, x := range TestItems {
			totalWeight += x.Weight()
		}
		for weightItem, stat := range weightStat {
			ratio := float32(stat) / float32(total)
			fmt.Printf("weight item %+v stat %v ratio %.4f\n", weightItem, stat, ratio)
			if !floatEqualLow(ratio, weightItem.weight/totalWeight) {
				t.Error("权重计算失败，请开启 -v 查看详细输出")
			}
		}
	})
	TestRandomWeightSingle(t)
	//TestRandomWeightTwoSingle(t)
}

func Benchmark_Weight(b *testing.B) {
	const column = 1000
	const row = 1000

	b.Run("MaxWeightOnce", func(b *testing.B) {
		maxWeightOnce(row, column, false)
	})

	b.Run("MaxWeight", func(b *testing.B) {
		maxWeight(row, column, false)
	})
}

func TestRandomWeightSingle(t *testing.T) {
	l := make([]IWeight, 0, len(TestItems))
	var totally float32
	for i, _ := range TestItems {
		l = append(l, TestItems[i])
		totally += TestItems[i].Weight()
	}
	resMap := make(map[_WeightItem]int)
	count := 1000000
	for i := 0; i < count; i++ {
		res := RandomWeightSingle(l, 1)
		if len(res) < 1 {
			t.Error("获取随机元素失败")
		}
		for _, x := range res {
			a := x.(_WeightItem)
			resMap[a]++
		}
	}
	fmt.Println(resMap)
	for res, val := range resMap {
		weight := fmt.Sprintf("%.2f", res.Weight()/totally)
		resWeight := fmt.Sprintf("%.2f", float32(val)/float32(count))
		if weight != resWeight {
			fmt.Println(res.icon, weight, resWeight)
			t.Error("权重计算失败")
		}
	}
	res := RandomWeightSingle(l, len(TestItems))
	fmt.Println(res)
	if len(res) < len(TestItems) {
		fmt.Println(len(res))
		t.Error("随机失败")
	}
}
func TestMixArray(t *testing.T) {
	l := make([]IWeight, 0, len(TestItems))
	for i, _ := range TestItems {
		l = append(l, TestItems[i])
	}
	fmt.Println(len(l))
	fmt.Println("mix pre", l)
	MixArray(l)
	fmt.Println("mix after", l)

}
func floatEqualLow(l float32, r float32) bool {
	if math.Abs(float64(l-r)) <= 0.01 {
		return true
	} else {
		return false
	}
}

func floatEqualHigh(l float32, r float32) bool {
	if math.Abs(float64(l-r)) <= 0.0001 {
		return true
	} else {
		return false
	}
}

func maxWeight(row int, column int, print bool) map[_WeightItem]int32 {
	total := row * column
	var testWeights []IWeight = make([]IWeight, 0, len(TestItems))

	var weightStat map[_WeightItem]int32 = make(map[_WeightItem]int32)
	for _, item := range TestItems {
		testWeights = append(testWeights, item)
		weightStat[item] = 0
	}

	results := RandomWeightWithMaxWeight(testWeights, 1.0, total)
	statResult(weightStat, column, row, func(r, c int) IWeight { return results[r*column+c] }, print)

	return weightStat
}

// func maxWeightSingle(row int, column int, print bool) map[_WeightItem]int32 {
// 	total := row * column
// 	var testWeights []IWeight = make([]IWeight, 0, len(TestItems))

// 	var weightStat map[_WeightItem]int32 = make(map[_WeightItem]int32)
// 	for _, item := range TestItems {
// 		testWeights = append(testWeights, item)
// 		weightStat[item] = 0
// 	}
// 	fmt.Println(len(testWeights), total)
// 	results := RandomWeightWithSingle(testWeights, 1.0, total)
// 	fmt.Println(weightStat, "jie go", results)
// 	statResult(weightStat, column, row, func(r, c int) IWeight {
// 		return results[r*column+c]
// 	}, print)

// 	return weightStat
// }

func maxWeightOnce(row int, column int, print bool) map[_WeightItem]int32 {
	var testWeights []IWeight = make([]IWeight, 0, len(TestItems))

	var weightStat map[_WeightItem]int32 = make(map[_WeightItem]int32)
	for _, item := range TestItems {
		testWeights = append(testWeights, item)
		weightStat[item] = 0
	}

	statResult(weightStat, column, row, func(r, c int) IWeight { return RandomWeightWithMaxWeightOnce(testWeights, 1.0) }, print)

	return weightStat
}

func statResult(weightStat map[_WeightItem]int32, column int, row int, itemFunc func(r int, c int) IWeight, print bool) {
	printf := fmt.Printf

	if print == false {
		printf = func(format string, a ...interface{}) (n int, err error) { return 0, nil }
	}

	printf("\n")
	for c := 0; c < column; c++ {
		for r := 0; r < row; r++ {
			item := itemFunc(r, c)
			if item == nil {
				printf(" ")
			} else {
				weightItem := item.(_WeightItem)
				weightStat[weightItem]++

				printf(weightItem.icon)
			}
		}
		printf("\n")
	}
	printf("\n")
}
