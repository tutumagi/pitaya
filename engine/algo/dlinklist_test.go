package algo

import (
	"fmt"
	"math/rand"
	"sort"
	"testing"
)

var less Compare = func(left NodeData, right NodeData) bool {
	if left.(int) < right.(int) {
		return true
	}
	return false
}

var moreThan Compare = func(left NodeData, right NodeData) bool {
	return less(right, left)
}

func TestDSortLinkList_Insert(t *testing.T) {

	t.Run("ascending", func(t *testing.T) {
		list, sortedDatas := newList(true, []int{10, 3, 100, 6, 5, 1000, 11}, less, moreThan)
		checkListContent(t, list, sortedDatas)
	})

	t.Run("descending", func(t *testing.T) {
		list, sortedDatas := newList(false, []int{10, 3, 100, 6, 5, 1000, 11}, moreThan, less)
		checkListContent(t, list, sortedDatas)
	})
}

func TestDSortLinkList_Remove(t *testing.T) {
	remove := func(t *testing.T, ascending bool) {
		orderStr := ""
		var lessThan Compare
		var greaterThan Compare
		if ascending {
			orderStr = "ascending"
			lessThan = less
			greaterThan = moreThan
		} else {
			orderStr = "descending"
			lessThan = moreThan
			greaterThan = less
		}
		t.Run(fmt.Sprintf("remove-head-%s", orderStr), func(t *testing.T) {
			list, sortedDatas := newList(ascending, []int{10, 3, 100, 6, 5, 1000, 11}, lessThan, greaterThan)
			checkListContent(t, list, sortedDatas)

			removeHeadDatas := sortedDatas[1:]
			list.Remove(list.getNodeByIndex(0))
			checkListContent(t, list, removeHeadDatas)
		})

		t.Run(fmt.Sprintf("remove-tail-%s", orderStr), func(t *testing.T) {
			list, sortedDatas := newList(ascending, []int{10, 3, 100, 6, 5, 1000, 11}, lessThan, greaterThan)
			checkListContent(t, list, sortedDatas)

			removeHeadDatas := sortedDatas[:len(sortedDatas)-1]
			list.Remove(list.getNodeByIndex(int(list.count) - 1))
			checkListContent(t, list, removeHeadDatas)
		})

		t.Run(fmt.Sprintf("remove-middle-%s", orderStr), func(t *testing.T) {
			list, sortedDatas := newList(ascending, []int{10, 3, 100, 6, 5, 1000, 11}, lessThan, greaterThan)
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
		less := less
		greater := moreThan
		if ascending == false {
			less = moreThan
			greater = less
		}

		list, sortedDatas := newList(ascending, []int{10, 18, 100}, less, greater)
		checkListContent(t, list, sortedDatas)
		node := list.getNodeByIndex(index)

		sortedDatas[index] = value

		list.modify(node, value)
		if ascending {
			sort.Ints(sortedDatas)
		} else {
			sort.Sort(sort.Reverse(sort.IntSlice(sortedDatas)))
		}

		checkListContent(t, list, sortedDatas)
	}

	t.Run("backend", func(t *testing.T) {
		t.Run("static", func(t *testing.T) {
			modifyDataFunc(t, true, 1, 20)
			modifyDataFunc(t, true, 1, 30)
			modifyDataFunc(t, true, 1, 40)
			modifyDataFunc(t, true, 1, 1)

		})
		t.Run("middle", func(t *testing.T) {
			// get list[3] 18，modify to 10
			modifyDataFunc(t, true, 1, 30)
		})
		t.Run("middle", func(t *testing.T) {
			// get list[3] 18，modify to 10
			modifyDataFunc(t, true, 1, 40)
		})
		t.Run("head", func(t *testing.T) {
			// get list[3] 18，modify to 1
			modifyDataFunc(t, true, 1, 1)
		})

	})
	t.Run("forward", func(t *testing.T) {
		t.Run("static", func(t *testing.T) {
			modifyDataFunc(t, true, 1, 15)

			modifyDataFunc(t, true, 1, 20)
		})
		t.Run("middle", func(t *testing.T) {
			modifyDataFunc(t, true, 1, 20)

			for i := 0; i < 100; i++ {
				modifyDataFunc(t, true, 1, rand.Intn(1000))
			}

		})
		t.Run("tail", func(t *testing.T) {
			modifyDataFunc(t, true, 2, 30)

			for i := 0; i < 100; i++ {
				modifyDataFunc(t, true, 2, rand.Intn(1000))
			}
		})

		t.Run("head", func(t *testing.T) {
			modifyDataFunc(t, true, 0, 30)

			for i := 0; i < 100; i++ {
				modifyDataFunc(t, true, 0, rand.Intn(1000))
			}
		})
	})
}

func newList(ascending bool, datas []int, less Compare, greater Compare) (list *DSortLinkList, sortedDatas []int) {
	list = NewDSortLinkList(less, greater, nil)
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
		if list.GetDataByIndex(index).(int) != num {
			t.Errorf("list[%d] is %d, should is %d", index, list.GetDataByIndex(index).(int), num)
		}
	}
	ok, errString := checkHeadAndTail(list, datas[0], datas[len(datas)-1])
	if !ok {
		t.Errorf(errString)
	}
}

func checkHeadAndTail(list *DSortLinkList, headVal int, tailVal int) (bool, string) {
	if list.Front().Data.(int) != headVal {
		return false, fmt.Sprintf("head(%d) should be %d", list.Front().Data.(int), headVal)
	}
	if list.Back().Data.(int) != tailVal {
		return false, fmt.Sprintf("tail(%d) should be %d", list.Back().Data.(int), tailVal)
	}
	return true, ""
}
