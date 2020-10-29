package aoi

// XZListManager aoi manager
// this file copy from https://github.com/xiaonanln/go-aoi.git
type XZListManager struct {
	radius     Coord
	xSweepList *xList
	zSweepList *zList
}

// NewXZListAOIManager with aoiDist
func NewXZListAOIManager(radius Coord) *XZListManager {
	return &XZListManager{
		radius:     radius,
		xSweepList: newXAOILIst(radius),
		zSweepList: newZAOIList(radius),
	}
}

// Enter new aoi with {x, z}
func (mgr *XZListManager) Enter(aoi *Item, x, z Coord) {
	aoi.radius = mgr.radius
	xzaoi := &xzaoi{
		aoi:       aoi,
		neighbors: map[*xzaoi]struct{}{},
		xl:        mgr.xSweepList,
		zl:        mgr.zSweepList,
	}
	aoi.x, aoi.z = x, z
	aoi.impData = xzaoi

	// logger.Info("enter aoi", zap.Float32("x", float32(aoi.x)), zap.Float32("z", float32(aoi.z)))
	mgr.xSweepList.Insert(xzaoi)
	mgr.zSweepList.Insert(xzaoi)
	mgr.Ajust(aoi)
}

// Leave aoi
func (mgr *XZListManager) Leave(aoi *Item) {
	xzaoi := aoi.impData
	mgr.xSweepList.Remove(xzaoi)
	mgr.zSweepList.Remove(xzaoi)
	mgr.Ajust(aoi)
}

// Moved aoi with new {x, y}
func (mgr *XZListManager) Moved(aoi *Item, x, z Coord) {
	oldX := aoi.x
	oldZ := aoi.z
	aoi.x, aoi.z = x, z
	xzaoi := aoi.impData
	if oldX != x {
		mgr.xSweepList.Move(xzaoi, oldX)
	}
	if oldZ != z {
		mgr.zSweepList.Move(xzaoi, oldZ)
	}

	mgr.Ajust(aoi)
}

// Ajust item 有变化
func (mgr *XZListManager) Ajust(item *Item) {

	aoi := item.impData

	mgr.xSweepList.mark(aoi)
	mgr.zSweepList.mark(aoi)

	// AOI marked twice are neighors
	for neighbor := range aoi.neighbors {
		if neighbor.markVal == 2 {
			neighbor.markVal = -2
		} else {
			// if aoi.aoi.flag&_InterestinFlag > 0 {
			if _, ok := aoi.neighbors[neighbor]; ok {
				delete(aoi.neighbors, neighbor)
				if aoi.aoi.callback != nil {
					aoi.aoi.callback.OnLeaveAOI(aoi.aoi, neighbor.aoi)
				}
			}
			// }

			// if neighbor.aoi.flag&_InterestinFlag > 0 {
			if _, ok := neighbor.neighbors[aoi]; ok {
				delete(neighbor.neighbors, aoi)
				if neighbor.aoi.callback != nil {
					neighbor.aoi.callback.OnLeaveAOI(neighbor.aoi, aoi.aoi)
				}
			}
			// }
		}
	}

	// travel in X list again to find all new neighbors, whose markVal == 2
	mgr.xSweepList.getClearMarkedNeighbors(aoi)
	// travel in Z list again to unmark all
	mgr.zSweepList.clearMark(aoi)

}
