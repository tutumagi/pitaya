package aoi

import (
	"fmt"
	"math/rand"
	"sort"
	"testing"
)

var less Less = func(left NodeData, right NodeData) bool {
	if left.(int) < right.(int) {
		return true
	}
	return false
}

var moreThan Less = func(left NodeData, right NodeData) bool {
	return less(right, left)
}

func TestDSortLinkList_Insert(t *testing.T) {

	t.Run("ascending", func(t *testing.T) {
		list, sortedDatas := newList(true, []int{10, 3, 100, 6, 5, 1000, 11}, less)
		checkListContent(t, list, sortedDatas)
	})

	t.Run("descending", func(t *testing.T) {
		list, sortedDatas := newList(false, []int{10, 3, 100, 6, 5, 1000, 11}, moreThan)
		checkListContent(t, list, sortedDatas)
	})
}

func TestDSortLinkList_Remove(t *testing.T) {
	remove := func(t *testing.T, ascending bool) {
		orderStr := ""
		var compare Less
		if ascending {
			orderStr = "ascending"
			compare = less
		} else {
			orderStr = "descending"
			compare = moreThan
		}
		t.Run(fmt.Sprintf("remove-head-%s", orderStr), func(t *testing.T) {
			list, sortedDatas := newList(ascending, []int{10, 3, 100, 6, 5, 1000, 11}, compare)
			checkListContent(t, list, sortedDatas)

			removeHeadDatas := sortedDatas[1:]
			list.Remove(list.getNodeByIndex(0))
			checkListContent(t, list, removeHeadDatas)
		})

		t.Run(fmt.Sprintf("remove-tail-%s", orderStr), func(t *testing.T) {
			list, sortedDatas := newList(ascending, []int{10, 3, 100, 6, 5, 1000, 11}, compare)
			checkListContent(t, list, sortedDatas)

			removeHeadDatas := sortedDatas[:len(sortedDatas)-1]
			list.Remove(list.getNodeByIndex(int(list.count) - 1))
			checkListContent(t, list, removeHeadDatas)
		})

		t.Run(fmt.Sprintf("remove-middle-%s", orderStr), func(t *testing.T) {
			list, sortedDatas := newList(ascending, []int{10, 3, 100, 6, 5, 1000, 11}, compare)
			checkListContent(t, list, sortedDatas)

			i := rand.Intn(len(sortedDatas) - 1)
			i++
			var removeMiddleDatas []int = make([]int, 0, len(sortedDatas)-1)
			for index, v := range sortedDatas {
				if index != i {
					removeMiddleDatas = append(removeMiddleDatas, v)
				}
			}
			t.Log(list.String())
			list.Remove(list.getNodeByIndex(i))
			checkListContent(t, list, removeMiddleDatas)
		})
	}

	remove(t, true)
	remove(t, false)
}

func TestDSortLinkList_ReSort(t *testing.T) {
	modifyDataFunc := func(t *testing.T, ascending bool, index int, value int) {
		compare := less
		if ascending == false {
			compare = moreThan
		}

		list, sortedDatas := newList(ascending, []int{10, 18}, compare)
		checkListContent(t, list, sortedDatas)
		node := list.getNodeByIndex(index)

		sortedDatas[index] = value

		list.Modify(node, value)
		if ascending {
			sort.Ints(sortedDatas)
		} else {
			sort.Sort(sort.Reverse(sort.IntSlice(sortedDatas)))
		}

		checkListContent(t, list, sortedDatas)
	}

	t.Run("backend", func(t *testing.T) {
		t.Run("static", func(t *testing.T) {
			// get list[3] 18，modify to 20
			modifyDataFunc(t, true, 1, 20)
			modifyDataFunc(t, true, 1, 30)
			modifyDataFunc(t, true, 1, 40)
			modifyDataFunc(t, true, 1, 1)

		})
		// t.Run("middle", func(t *testing.T) {
		// 	// get list[3] 18，modify to 10
		// 	modifyDataFunc(t, true, 1, 30)
		// })
		// t.Run("middle", func(t *testing.T) {
		// 	// get list[3] 18，modify to 10
		// 	modifyDataFunc(t, true, 1, 40)
		// })
		// t.Run("head", func(t *testing.T) {
		// 	// get list[3] 18，modify to 1
		// 	modifyDataFunc(t, true, 1, 1)
		// })

	})
	// t.Run("forward", func(t *testing.T) {
	// 	t.Run("static", func(t *testing.T) {
	// 		// get list[3] 18，modify to 15
	// 		modifyDataFunc(t, true, 1, 15)
	// 		modifyDataFunc(t, true, 1, rand.Intn(1000))

	// 		modifyDataFunc(t, true, 1, rand.Intn(1000))

	// 		modifyDataFunc(t, true, 1, rand.Intn(1000))

	// 	})
	// 	t.Run("middle", func(t *testing.T) {
	// 		// get list[3] 18，modify to 20
	// 		modifyDataFunc(t, true, 1, 20)

	// 		for i := 0; i < 100; i++ {
	// 			modifyDataFunc(t, true, 1, rand.Intn(1000))
	// 		}

	// 	})
	// 	t.Run("tail", func(t *testing.T) {
	// 		// get list[3] 18，modify to 30
	// 		modifyDataFunc(t, true, 1, 30)

	// 		for i := 0; i < 100; i++ {
	// 			modifyDataFunc(t, true, 1, rand.Intn(1000))
	// 		}
	// 	})
	// })
}

func newList(ascending bool, datas []int, compare Less) (list *DSortLinkList, sortedDatas []int) {
	list = NewDSortLinkList(compare)
	for _, data := range datas {
		node := &Node{
			Data: data,
		}
		list.Insert(node)
	}

	if ascending {
		sort.Ints(datas)
	} else {
		// sort.Sort(sort.Reverse(sort.IntSlice(datas))) 降序排序
		sort.Sort(sort.Reverse(sort.IntSlice(datas)))
	}
	return list, datas

}

func checkListContent(t *testing.T, list *DSortLinkList, datas []int) {
	for index, num := range datas {
		if list.getDataByIndex(index).(int) != num {
			t.Errorf("list[%d] is %d, should is %d", index, list.getDataByIndex(index).(int), num)
		}
	}
	ok, errString := checkHeadAndTail(list, datas[0], datas[len(datas)-1])
	if !ok {
		t.Errorf(errString)
	}
}

func checkHeadAndTail(list *DSortLinkList, headVal int, tailVal int) (bool, string) {
	if list.head.Data.(int) != headVal {
		return false, fmt.Sprintf("head(%d) should be %d", list.head.Data.(int), headVal)
	}
	if list.tail.Data.(int) != tailVal {
		return false, fmt.Sprintf("tail(%d) should be %d", list.tail.Data.(int), tailVal)
	}
	return true, ""
}
