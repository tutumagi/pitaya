package aoi

import (
	"fmt"
	"strings"
)

// // ComparableData 可比较的数据
// type ComparableData interface {
// 	//	if use ascending. return true when left < right. otherwise return false
// 	//	else use descending. return true when left > right. otherwise return true
// 	Less(left ComparableData, right ComparableData) bool
// }

// NodeData type bind to Node
type NodeData interface{}

// Less function Type
//	left < right return true
type Less func(left NodeData, right NodeData) bool

// DSortLinkList sorted-double-link-list is no-safe-goroutine
//	有序的双向链表
type DSortLinkList struct {
	head  *Node
	tail  *Node
	count int32

	less Less
}

// Node in DSortLinkList
type Node struct {
	Prev *Node
	Next *Node
	Data NodeData
}

// func (n *Node) GetData() NodeData {
// 	return n.data
// }

// NewDSortLinkList new a sorted-double-link-list with ascend or descend
func NewDSortLinkList(less Less) *DSortLinkList {
	return &DSortLinkList{
		less: less,
	}
}

// Insert new Node
//	TODO: insert 的最坏时间复杂度是 O(n)，要考虑是否有瓶颈
func (l *DSortLinkList) Insert(node *Node) {
	l.count++

	head := l.head
	// DSortLinkList is empty
	if head == nil {
		l.head = node
		l.tail = node
	} else {
		// iterate from head -> tail find the index in linklist by `ascending` condition
		cur := head
		for cur != nil && l.less(cur.Data, node.Data) {
			cur = cur.Next
		}
		// 在 cur 前面 处插入 node
		if cur == nil {
			// 需要插入在最后面
			l.tail.Next = node
			node.Prev = l.tail

			l.tail = node
		} else {
			prev := cur.Prev

			node.Next = cur
			cur.Prev = node
			if prev == nil {
				// 如果是插入在 head 前面，则 head 重新指向 最新的 node
				l.head = node
			} else {
				prev.Next = node
				node.Prev = prev
			}
			if cur.Next == nil {
				l.tail = cur
			}
		}
	}
}

// Remove node from linklist
func (l *DSortLinkList) Remove(node *Node) {
	if l.count <= 0 {
		return
	}
	l.count--
	prev := node.Prev
	next := node.Next

	if prev == nil {
		l.head = next
	} else {
		prev.Next = next
	}
	if next == nil {
		l.tail = prev
	} else {
		next.Prev = prev
	}

	node.Prev = nil
	node.Next = nil
}

// Modify node' data to newData. The method will resort the linklist
//	适用于 NodeData 为比较简单的数据结构时，比如内置数据类型(int, double ...)
func (l *DSortLinkList) Modify(node *Node, newData NodeData) {
	backend := l.less(newData, node.Data)
	node.Data = newData
	l.ReSort(node, backend)
}

// ReSort imp.
//	from node position.
// 	if `backword` is true, 则向前查找最合适的位置进行重排序
// 	否则，向后查找最合适的位置进行重排序
func (l *DSortLinkList) ReSort(node *Node, backward bool) {
	// fmt.Print("begin sort ", l.String())
	if backward {

		find := node.Prev
		if find == nil || l.less(find.Data, node.Data) {
			return
		}

		for find != nil && !l.less(find.Data, node.Data) {
			find = find.Prev
		}

		// 处理 node 的 prev的 next 和 next的 prev
		if node.Prev != nil {
			node.Prev.Next = node.Next
		}
		if node.Next != nil {
			node.Next.Prev = node.Prev
		} else {
			// node 是 tail
			if node.Prev != nil {
				l.tail = node.Prev
			}
		}

		// 在 find 后面插入 node

		// 处理 node 的 prev
		if find != nil {

			node.Prev = find
			node.Next = find.Next

			if find.Next != nil {
				find.Next.Prev = node
			}

			find.Next = node
		} else {

			// 在头部插入
			l.head.Prev = node
			node.Next = l.head
			node.Prev = nil

			l.head = node

		}

	} else {
		find := node.Next
		if find == nil || !l.less(find.Data, node.Data) {
			return
		}

		for find != nil && l.less(find.Data, node.Data) {
			find = find.Next
		}

		// 处理 node 的 prev的 next 和 next的 prev
		if node.Prev != nil {
			node.Prev.Next = node.Next
		} else {
			if node.Next != nil {
				l.head = node.Next
			}
		}
		if node.Next != nil {
			node.Next.Prev = node.Prev
		}

		// 在 find 前面插入
		if find != nil {

			node.Prev = find.Prev
			node.Next = find

			if find.Prev != nil {
				find.Prev.Next = node
			}
			find.Prev = node
		} else {

			// 在最后插入
			l.tail.Next = node
			node.Prev = l.tail
			node.Next = nil

			l.tail = node
		}
	}

	// TODO: 调试用的，测试稳定后，删掉
	if l.head.Next == nil {
		panic("fuck the head")
	}
	// TODO: 调试用的，测试稳定后，删掉
	if node.Next == node {
		panic("what the fuck")
	}
}

// IsEmpty DSortLinkList is empty
func (l *DSortLinkList) IsEmpty() bool {
	return l.head == nil
}

// Count of DSortLinkList
func (l *DSortLinkList) Count() int32 {
	return l.count
}

// getDataByIndex in DSortLinkList
//	just use for test
func (l *DSortLinkList) getDataByIndex(index int) NodeData {
	i := 0
	node := l.head
	for i < index {
		node = node.Next
		i++
	}
	return node.Data
}

// getNodeByIndex in DSortLinkList
//	just use for test
func (l *DSortLinkList) getNodeByIndex(index int) *Node {
	i := 0
	node := l.head
	for i < index {
		node = node.Next
		i++
	}
	return node
}

func (l *DSortLinkList) String() string {
	node := l.head
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("DSortLinkList %p:", &l))
	for node != nil {
		sb.WriteString(fmt.Sprintf("[%v]", node.Data))
		if node != l.tail {
			sb.WriteString(" -> ")
		}
		node = node.Next
	}
	sb.WriteString("\n")

	return sb.String()
}
