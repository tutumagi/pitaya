package aoi

import (
	"fmt"
	"sort"

	"github.com/tutumagi/pitaya/engine/algo"
)

// node 存储坐标的节点，包含x，z轴两个在双向链表中的节点
type node struct {
	coord *BaseCoord

	xnode *algo.Node // 引用在x方向的双向链表中的节点
	znode *algo.Node // 引用在z方向的双向链表中的节点
}

func (xz *node) String() string {
	if xz.coord.entity != nil {
		return fmt.Sprintf("<xzaoi> %s", xz.coord.entity)
	}
	return fmt.Sprintf("<xzaoi> %s", xz.coord.delegate)
}

// func (xz *node) xPrev() *node {
// 	if xz.xnode != nil && xz.xnode.Prev != nil {
// 		return xz.xnode.Prev.Data.(*node)
// 	}
// 	return nil
// }

// func (xz *node) xNext() *node {
// 	if xz.xnode != nil && xz.xnode.Next != nil {
// 		return xz.xnode.Next.Data.(*node)
// 	}
// 	return nil
// }

// func (xz *node) zPrev() *node {
// 	if xz.znode != nil && xz.znode.Prev != nil {
// 		return xz.znode.Prev.Data.(*node)
// 	}
// 	return nil
// }
// func (xz *node) zNext() *node {
// 	if xz.znode != nil && xz.znode.Next != nil {
// 		return xz.znode.Next.Data.(*node)
// 	}
// 	return nil
// }

// CoordSystem aoi manager
type CoordSystem struct {
	xSweepList *algo.DSortLinkList
	zSweepList *algo.DSortLinkList

	count int32
}

// NewCoordSystem with aoiDist
func NewCoordSystem() Systemer {
	return &CoordSystem{
		xSweepList: algo.NewDSortLinkList(xLessThan, xGreaterThan, xPass),
		zSweepList: algo.NewDSortLinkList(zLessThan, zGreaterThan, zPass),
	}
}

// Insert new aoi with {x, z}
func (mgr *CoordSystem) Insert(coord *BaseCoord) {
	// 节点从系统中洗出后， xnode 和 znode 都为nil，但是 xznode不为nil
	// 这里先判断原来的xznode 有没有，没有则再进行创建
	xzNode := coord.node()
	if xzNode == nil {
		// TODO 这里加一个节点池
		xzNode = &node{
			coord: coord,
		}
		coord.setNode(xzNode)
	}
	xzNode.xnode = &algo.Node{Data: xzNode}
	xzNode.znode = &algo.Node{Data: xzNode}

	coord.setSystem(mgr)

	mgr.xSweepList.Insert(xzNode.xnode)
	mgr.zSweepList.Insert(xzNode.znode)

	mgr.count++

	coord.delegate.resetOld()
}

// InsertWithRef 根据参考的节点，去进行插入，这里不会进行排序
//	当已经有大量实体的时候，比如150000个的时候，为了避免从头开始找位置进行插入，可以找一个参考点
//	这种很适合地图中的实体有类似格子的实体的应用
func (mgr *CoordSystem) InsertWithRef(coord *BaseCoord, ref *BaseCoord) {
	xzNode := coord.node()
	if xzNode == nil {
		// TODO 这里加一个节点池
		xzNode = &node{
			coord: coord,
		}
		coord.setNode(xzNode)
	}
	xzNode.xnode = &algo.Node{Data: xzNode}
	xzNode.znode = &algo.Node{Data: xzNode}

	coord.setSystem(mgr)

	// 先根据参考点插入
	if ref != nil {
		mgr.xSweepList.InsertPrevRef(xzNode.xnode, ref.xzNode.xnode)
		mgr.zSweepList.InsertPrevRef(xzNode.znode, ref.xzNode.znode)
	} else {
		mgr.xSweepList.InsertPrevRef(xzNode.xnode, nil)
		mgr.zSweepList.InsertPrevRef(xzNode.znode, nil)
	}
	// 然后update此节点，进行排序
	coord.delegate.Update()

	mgr.count++

	coord.delegate.resetOld()
}

// InsertZeroRadiusEntities 初始化需要插入大量aoi半径为0的实体的优化插入
//	1. 这里参数不使用[]*BaseCoord的原因，避免多产生一个临时变量来存储
//	2. 先针对x轴，z轴进行排序，然后分别插入x轴和z轴，因为插入的都是没有aoi半径的实体，所以分别插入不会有影响
func (mgr *CoordSystem) InsertZeroRadiusEntities(entities []Entityer) {
	for _, e := range entities {
		coord := e.Coord()
		xzNode := coord.node()
		if xzNode == nil {
			// TODO 这里加一个节点池
			xzNode = &node{
				coord: coord.BaseCoord,
			}
			coord.setNode(xzNode)
		}
		xzNode.xnode = &algo.Node{Data: xzNode}
		xzNode.znode = &algo.Node{Data: xzNode}

		coord.setSystem(mgr)

		mgr.count++
	}
	// 插入x轴
	var coordsX _SortByCoordX = entities
	sort.Sort(coordsX)
	for _, e := range coordsX {
		mgr.xSweepList.Insert(e.Coord().xzNode.xnode)
	}

	var coordsZ _SortByCoordZ = entities
	sort.Sort(coordsZ)
	for _, e := range coordsZ {
		mgr.zSweepList.Insert(e.Coord().xzNode.znode)
	}

	for _, e := range entities {
		e.Coord().resetOld()
	}
}

// Remove aoi coord
func (mgr *CoordSystem) Remove(coord *BaseCoord) {
	// coord.addFlags(nodeFlagRemoving)
	// mgr.Update(coord, -math.MaxFloat32, -math.MaxFloat32)
	// coord.addFlags(nodeFlagRemoved)
	// 时序很重要
	coord.addFlags(nodeFlagRemoving)

	// 是否统一释放 remove 节点 参考 keb coordinate_system.cpp line:315
	xzaoi := coord.node()
	if xzaoi.xnode != nil {
		mgr.xSweepList.Remove(xzaoi.xnode)
		xzaoi.xnode = nil
	}
	if xzaoi.znode != nil {
		mgr.zSweepList.Remove(xzaoi.znode)
		xzaoi.znode = nil
	}

	coord.delegate.onSystemRemoved()
	coord.setSystem(nil)

	coord.addFlags(nodeFlagRemoved)

	mgr.count--
}

// Update 坐标变化后，比如坐标移动了
func (mgr *CoordSystem) Update(coord *BaseCoord, newX float32, newZ float32) {
	// 如果老的坐标 不等于新的坐标
	if coord.old.X != newX {
		coord.pos.X = newX
		xnode := coord.node().xnode
		mgr.xSweepList.ReSort(xnode, false)
		mgr.xSweepList.ReSort(xnode, true)
	}

	if coord.old.Z != newZ {
		coord.pos.Z = newZ
		znode := coord.node().znode
		mgr.zSweepList.ReSort(znode, false)
		mgr.zSweepList.ReSort(znode, true)
	}

	coord.delegate.resetOld()
}

// Dump the nodes
func (mgr *CoordSystem) Dump() string {
	return fmt.Sprintf("%s\n%s", mgr.dumpXList(), mgr.dumpZList())
}

func (mgr *CoordSystem) dumpXList() string {
	return fmt.Sprintf("<xList> %s", mgr.xSweepList.DebugString(func(data interface{}) string {
		return data.(*node).coord.delegate.xString()
	}))
}

func (mgr *CoordSystem) dumpZList() string {
	return fmt.Sprintf("<zList> %s", mgr.zSweepList.DebugString(func(data interface{}) string {
		return data.(*node).coord.delegate.zString()
	}))
}
