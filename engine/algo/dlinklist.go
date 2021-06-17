package algo

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

type Compare func(left NodeData, right NodeData) bool

// // Less function Type
// //	left < right return true
// type Less func(left NodeData, right NodeData) bool

// // Greater func type
// // left > right return true
// type Greater func(left NodeData, right NodeData) bool

// Pass 节点数据变化时，跨过的节点回调，front 表示 n1 是 n2 前面的节点
// 当 node 数据变化时，node的位置发生了改变，node从当前位置到改变后的位置跨过的节点
// n1 为 node， n2 为跨过的节点，front 表示 n1 是否是 n2 前面的节点（根据 Less 决定升序还是降序）
//  比如 升序的双向链表： 1,2,3,4,5,6,7
//  如果3变成了6.5，则 3 经过了 4,5,6 这三个节点，front 为 false
//  如果3变成了1.2，则 3 经过了 2 这个节点， front 为 true
type Pass func(n1 NodeData, n2 NodeData, front bool)

// DSortLinkList sorted-double-link-list is no-safe-goroutine
//	有序的双向链表
type DSortLinkList struct {
	root  Node
	count int

	less    Compare
	greater Compare

	pass Pass
}

// Node in DSortLinkList
type Node struct {
	prev *Node
	next *Node
	Data NodeData
}

// NewDSortLinkList new a sorted-double-link-list with ascend or descend
// Less function Type
//	left < right return true
// Greater func type
//	left > right return true
func NewDSortLinkList(less Compare, greater Compare, pass Pass) *DSortLinkList {
	if pass == nil {
		pass = func(n1, n2 NodeData, front bool) {}
	}

	l := &DSortLinkList{
		less:    less,
		greater: greater,
		pass:    pass,
	}
	l.root.next = &l.root
	l.root.prev = &l.root
	l.count = 0
	return l
}

// Front returns the first element of list l or nil if the list is empty.
func (l *DSortLinkList) Front() *Node {
	if l.count == 0 {
		return nil
	}
	return l.root.next
}

// Back returns the last element of list l or nil if the list is empty.
func (l *DSortLinkList) Back() *Node {
	if l.count == 0 {
		return nil
	}
	return l.root.prev
}

// Insert new Node
//	TODO: insert 的最坏时间复杂度是 O(n)，要考虑是否有瓶颈
func (l *DSortLinkList) Insert(node *Node) {
	if l.count == 0 {
		l.insert(node, &l.root)
	} else {
		// iterate from head -> tail find the index in linklist by `ascending` condition
		find := l.Front()
		for find != &l.root && l.less(find.Data, node.Data) {
			l.pass(node.Data, find.Data, false)
			find = find.next
		}

		l.insert(node, find.prev)
	}
}

// Remove node
func (l *DSortLinkList) Remove(e *Node) *Node {
	e.prev.next = e.next
	e.next.prev = e.prev

	e.next = nil
	e.prev = nil
	// e.list = nil
	l.count--
	return e
}

// InsertPrevRef 不排序，插入到ref前面
func (l *DSortLinkList) InsertPrevRef(node *Node, ref *Node) {
	if l.count == 0 {
		l.insert(node, &l.root)
	} else {
		l.insert(node, ref.prev)
	}
}

// modify node' data to newData. The method will resort the linklist
//	适用于 NodeData 为比较简单的数据结构时，比如内置数据类型(int, double ...)
func (l *DSortLinkList) modify(node *Node, newData NodeData) {
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
		find := node.prev

		for find != &l.root && l.greater(find.Data, node.Data) {
			l.pass(node.Data, find.Data, true)
			find = find.prev
		}

		l.move(node, find)
	} else {
		find := node.next

		for find != &l.root && l.less(find.Data, node.Data) {
			l.pass(node.Data, find.Data, false)
			find = find.next
		}

		l.move(node, find.prev)
	}

	// TODO: 调试用的，测试稳定后，删掉
	if l.Front().next == nil {
		panic("fuck the head")
	}
	// TODO: 调试用的，测试稳定后，删掉
	if node.next == node {
		panic("what the fuck")
	}
}

// IsEmpty DSortLinkList is empty
func (l *DSortLinkList) IsEmpty() bool {
	return l.count == 0
}

// Count of DSortLinkList
func (l *DSortLinkList) Count() int {
	return l.count
}

// GetDataByIndex in DSortLinkList
//	just use for test
func (l *DSortLinkList) GetDataByIndex(index int) NodeData {
	i := 0
	node := l.Front()
	for i < index {
		node = node.next
		i++
	}
	return node.Data
}

// getNodeByIndex in DSortLinkList
//	just use for test
func (l *DSortLinkList) getNodeByIndex(index int) *Node {
	i := 0
	node := l.Front()
	for i < index {
		node = node.next
		i++
	}
	return node
}

// insert inserts e after at, increments l.len, and returns s.
func (l *DSortLinkList) insert(e, at *Node) *Node {
	e.prev = at
	e.next = at.next
	e.prev.next = e
	e.next.prev = e
	// e.list = l
	l.count++
	return e
}

// move moves e to next to at and return e
func (l *DSortLinkList) move(e, at *Node) *Node {
	if e == at {
		return e
	}
	e.prev.next = e.next
	e.next.prev = e.prev

	e.prev = at
	e.next = at.next
	e.prev.next = e
	e.next.prev = e

	return e
}

func (l *DSortLinkList) String() string {
	return l.DebugString(func(data interface{}) string { return fmt.Sprintf("%s", data) })
}

// DebugString 方便调试
func (l *DSortLinkList) DebugString(printFn func(data interface{}) string) string {
	node := l.Front()
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("DSortLinkList %p:\n", &l))
	for node != nil {
		sb.WriteString(fmt.Sprintf("[%s]", printFn(node.Data)))
		if node != l.Back() {
			sb.WriteString(" -> \n")
		} else {
			break
		}
		node = node.next
	}
	sb.WriteString("\n")

	return sb.String()
}
