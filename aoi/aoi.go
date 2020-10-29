package aoi

type _Flag uint8

const (
	// 需要被周围的关心
	_InterestedFlag _Flag = 1 << iota
	// 需要关心周围的
	_InterestinFlag
	// 即需要被关心也需要关心周围的
	_InterestAllFlag = _InterestedFlag | _InterestinFlag
)

// Coord 坐标单位
type Coord float32

// Item item
type Item struct {
	x      Coord
	z      Coord
	radius Coord
	Data   interface{}

	callback ICallback
	impData  *xzaoi

	flag _Flag
}

// NewItem item with aoi radius, custom data, callback
func NewItem(radius Coord, data interface{}, callback ICallback) *Item {
	return &Item{
		radius:   radius,
		Data:     data,
		flag:     _InterestAllFlag,
		callback: callback,
	}
}

// // NewInterestedItem 创建被关心的aoi item，但不关心周围的
// func NewInterestedItem(data interface{}) *Item {
// 	return &Item{
// 		radius: 0,
// 		Data:   data,
// 		flag:   _InterestedFlag,
// 	}
// }

// // NewInterestinItem 创建关心周围 的aoi item，但不被周围实体关心
// func NewInterestinItem(radius Coord, data interface{}) *Item {
// 	return &Item{
// 		radius: radius,
// 		Data:   data,
// 		flag:   _InterestinFlag,
// 	}
// }

// ICallback interface
type ICallback interface {
	// OnEnterAOI other enter my view
	OnEnterAOI(self *Item, other *Item)
	// OnLeaveAOI other leave my view
	OnLeaveAOI(self *Item, other *Item)
}

// IManager interface
type IManager interface {
	// Enter aoi
	Enter(aoi *Item, x, y Coord)
	Leave(aoi *Item)
	Moved(aoi *Item, x, y Coord)
}
