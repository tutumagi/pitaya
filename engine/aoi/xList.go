package aoi

import (
	"github.com/tutumagi/pitaya/engine/algo"
)

// head -> tail the value of x is increased

func xGreaterThan(findingNode algo.NodeData, movingNode algo.NodeData) bool {
	findingCoord := findingNode.(*node).coord
	movingCoord := movingNode.(*node).coord
	fx := findingCoord.pos.X
	mx := movingCoord.pos.X

	// mIsPositive := movingCoord.hasFlags(nodeFlagPositiveBoundary)
	// fIsEntity := findingCoord.entity != nil
	// !(mIsPositive && fIsEntity)

	if fx > mx ||
		(fx == mx && !(findingCoord.entity != nil && (movingCoord._flags&nodeFlagPositiveBoundary) > 0) && !(findingCoord._flags&nodeFlagNegativeBoundary > 0)) {
		return true
	}
	return false
}

func xLessThan(findingNode algo.NodeData, movingNode algo.NodeData) bool {
	findingCoord := findingNode.(*node).coord
	movingCoord := movingNode.(*node).coord

	fx := findingCoord.pos.X
	mx := movingCoord.pos.X

	// mIsNegative := movingCoord.hasFlags(nodeFlagNegativeBoundary)
	// fIsEntity := findingCoord.entity != nil
	// !(mIsNegative && fIsEntity)

	if fx < mx ||
		// 下面这行代码的意思是
		// 有A节点，位置在0，有5的视距半径，则此时的链表为 -A(-5) -> A(0) -> +A(5)  带有正负符号的表示非实体节点，而是扩展节点
		// 此时插入B节点（位置在5，有5的视距半径），movingCoord 为A(5) 从链表头开始找，一直找到A(+5)，
		// 此时 findingCoord 为+A(5), 有 positiveFlag 所以返回false，插入到+A(+5)的前面
		// 结果为 -A(-5) -> A(0) -> B(5) -> +A(5)，
		// 这样B就在A的视距范围内了.
		// 然后插入B节点的扩展节点-B(0),+B(10). B(+10)很简单，直接会找到最后，-B(0)找到A(0)时，发现A(0)没有 positionFlag
		// 继续找前一个节点，找到-A(-5), 在 -A(-5) 后面插入 -B(0)
		// 结果为 -A(-5) -> -B(0) -> A(0) -> B(5) -> +A(+5) -> +B(+10)
		// 这样 B的视距范围刚好有A时，可以看到A，A也可以看到B
		(fx == mx && !(findingCoord.entity != nil && (movingCoord._flags&nodeFlagNegativeBoundary > 0)) && !(findingCoord._flags&nodeFlagPositiveBoundary > 0)) {
		return true
	}
	return false
}

// 前 -------------------- 后
// 0, 1, 2, 3, 4, 5, 6, 7, 8
func xPass(n1 algo.NodeData, n2 algo.NodeData, front bool) {
	coord1 := n1.(*node).coord
	coord2 := n2.(*node).coord

	// coord1 为移动的节点
	// front 表示 coord1 是 coord2 前面的节点

	// coord2 越过 coord1
	if !coord2.hasFlags(nodeFlagHideOrRemoved) {
		coord1.delegate.onNodePassX(coord2, front)
	}
	if !coord1.hasFlags(nodeFlagHideOrRemoved) {
		coord2.delegate.onNodePassX(coord1, !front)
	}
}
