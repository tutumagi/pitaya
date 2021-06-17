package aoi

import (
	"github.com/tutumagi/pitaya/engine/algo"
)

// 往前找
func zGreaterThan(findingNode algo.NodeData, movingNode algo.NodeData) bool {
	findingCoord := findingNode.(*node).coord
	movingCoord := movingNode.(*node).coord

	fz := findingCoord.pos.Z
	mz := movingCoord.pos.Z

	// fz == mz 时， 移动的坐标是 +A，找到的坐标是B，确保链表顺序为 B -> +A
	// mIsPositive := movingCoord.hasFlags(nodeFlagPositiveBoundary)
	// fIsEntity := findingCoord.entity != nil
	// !(mIsPositive && fIsEntity)

	if fz > mz ||
		(fz == mz && !(findingCoord.entity != nil && (movingCoord._flags&nodeFlagPositiveBoundary > 0)) && !(findingCoord._flags&nodeFlagNegativeBoundary > 0)) {
		return true
	}
	return false
}

// 往后找
func zLessThan(findingNode algo.NodeData, movingNode algo.NodeData) bool {
	findingCoord := findingNode.(*node).coord
	movingCoord := movingNode.(*node).coord

	fz := findingCoord.pos.Z
	mz := movingCoord.pos.Z

	// fz == mz 时， 移动的坐标是 -A，找到的坐标是B，确保链表顺序为 -A -> B
	// mIsNegative := movingCoord.hasFlags(nodeFlagNegativeBoundary)
	// fIsEntity := findingCoord.entity != nil
	// !(mIsNegative && fIsEntity)

	if fz < mz ||
		(fz == mz && !(findingCoord.entity != nil && (movingCoord._flags&nodeFlagNegativeBoundary > 0)) && !(findingCoord._flags&nodeFlagPositiveBoundary > 0)) {
		return true
	}
	return false
}

func zPass(n1 algo.NodeData, n2 algo.NodeData, front bool) {
	coord1 := n1.(*node).coord
	coord2 := n2.(*node).coord

	if !coord2.hasFlags(nodeFlagHideOrRemoved) {
		coord1.delegate.onNodePassZ(coord2, front)
	}
	if !coord1.hasFlags(nodeFlagHideOrRemoved) {
		coord2.delegate.onNodePassZ(coord1, !front)
	}
}
