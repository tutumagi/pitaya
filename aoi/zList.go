package aoi

type zList struct {
	radius Coord
	list   *DSortLinkList
}

func newZAOIList(radius Coord) *zList {
	compare := func(left NodeData, right NodeData) bool {
		if left.(*xzaoi).aoi.z < right.(*xzaoi).aoi.z {
			return true
		}
		return false
	}
	return &zList{
		radius: radius,
		list:   NewDSortLinkList(compare),
	}
}

func (sl *zList) Insert(aoi *xzaoi) {
	node := &Node{
		Data: aoi,
	}
	aoi.znode = node
	sl.list.Insert(node)
}

func (sl *zList) Remove(aoi *xzaoi) {
	if aoi.znode != nil {
		sl.list.Remove(aoi.znode)
		aoi.znode = nil
	}
}

func (sl *zList) Move(aoi *xzaoi, oldCoord Coord) {
	coord := aoi.aoi.z
	if coord > oldCoord {
		sl.list.ReSort(aoi.znode, false)
	} else {
		sl.list.ReSort(aoi.znode, true)
	}
}

func (sl *zList) mark(aoi *xzaoi) {
	// if aoi.aoi.flag&_InterestedFlag > 0 {
	coord := aoi.aoi.z
	prev := aoi.zPrev()

	minCoord := coord - sl.radius
	for prev != nil && prev.aoi.z >= minCoord {
		// if prev.aoi.flag&_InterestedFlag > 0 {
		prev.markVal++
		// }
		prev = prev.zPrev()
	}

	next := aoi.zNext()
	maxCoord := coord + sl.radius
	for next != nil && next.aoi.z <= maxCoord {
		// if next.aoi.flag&_InterestedFlag > 0 {
		next.markVal++
		// }
		next = next.zNext()
	}
	// }
}

func (sl *zList) clearMark(aoi *xzaoi) {
	prev := aoi.zPrev()
	coord := aoi.aoi.z

	minCoord := coord - sl.radius
	for prev != nil && prev.aoi.z >= minCoord {
		prev.markVal = 0
		prev = prev.zPrev()
	}

	next := aoi.zNext()
	maxCoord := coord + sl.radius
	for next != nil && next.aoi.z <= maxCoord {
		next.markVal = 0
		next = next.zNext()
	}
}
