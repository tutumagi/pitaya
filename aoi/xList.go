package aoi

// head -> tail the value of x is increased
type xList struct {
	radius Coord
	list   *DSortLinkList
}

func newXAOILIst(radius Coord) *xList {
	compare := func(left NodeData, right NodeData) bool {
		if left.(*xzaoi).aoi.x < right.(*xzaoi).aoi.x {
			return true
		}
		return false
	}
	return &xList{
		radius: radius,
		list:   NewDSortLinkList(compare),
	}
}

// Insert new xzaoi item.
func (sl *xList) Insert(aoi *xzaoi) {
	node := &Node{
		Data: aoi,
	}
	aoi.xnode = node

	sl.list.Insert(node)
}

func (sl *xList) Remove(aoi *xzaoi) {
	if aoi.xnode != nil {
		sl.list.Remove(aoi.xnode)
		aoi.xnode = nil
	}
}

func (sl *xList) Move(aoi *xzaoi, oldCoord Coord) {
	// if aoi.node != nil {
	coord := aoi.aoi.x
	if coord > oldCoord {
		sl.list.ReSort(aoi.xnode, false)
	} else {
		sl.list.ReSort(aoi.xnode, true)
	}
	// }
}

// Mark 标记
// 此aoi 周围的x轴方向 在 radius 半径内,将mark+1
func (sl *xList) mark(aoi *xzaoi) {
	// 如果该 aoi 需要关心周围的

	prev := aoi.xPrev()
	coord := aoi.aoi.x

	// 找到x轴上所有 在radius范围之类的item，并做标记
	minCoord := coord - sl.radius
	for prev != nil && prev.aoi.x >= minCoord {
		// if prev.aoi.flag&_InterestedFlag > 0 {
		prev.markVal++
		// }
		prev = prev.xPrev()
	}

	next := aoi.xNext()
	maxCoord := coord + sl.radius
	for next != nil && next.aoi.x <= maxCoord {
		// if next.aoi.flag&_InterestedFlag > 0 {
		next.markVal++
		// }
		next = next.xNext()
	}
}

func (sl *xList) getClearMarkedNeighbors(aoi *xzaoi) {

	dealNewNeighbour := func(newNeighbour *xzaoi) {
		// if aoi.aoi.flag&_InterestinFlag > 0 {
		if _, ok := aoi.neighbors[newNeighbour]; !ok {
			aoi.neighbors[newNeighbour] = struct{}{}
			aoi.aoi.callback.OnEnterAOI(aoi.aoi, newNeighbour.aoi)
		}
		// }

		// if newNeighbour.aoi.flag&_InterestinFlag > 0 {
		if _, ok := newNeighbour.neighbors[aoi]; !ok {
			newNeighbour.neighbors[aoi] = struct{}{}
			newNeighbour.aoi.callback.OnEnterAOI(newNeighbour.aoi, aoi.aoi)
		}
		// }
	}

	prev := aoi.xPrev()
	coord := aoi.aoi.x
	minCoord := coord - sl.radius
	for prev != nil && prev.aoi.x >= minCoord {
		if prev.markVal == 2 { // 表示新邻居
			dealNewNeighbour(prev)
		}
		prev.markVal = 0
		prev = prev.xPrev()
	}

	next := aoi.xNext()
	maxCoord := coord + sl.radius
	for next != nil && next.aoi.x <= maxCoord {
		if next.markVal == 2 {
			dealNewNeighbour(next)
		}
		next.markVal = 0
		next = next.xNext()
	}
}
