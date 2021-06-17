package aoi

import (
	"fmt"
	"math"
)

// EntityCoord 实体的坐标节点
type EntityCoord struct {
	*BaseCoord

	// 该实体节点的监听者
	// 比如该实体有矩形的视距，
	// 在安装矩形视距触发器是，会额外生成两个Coorder，
	// 一个是正向范围的，一个负向范围的
	// 这两个Coorder 就会在这个watcher里面
	// 当当前节点更新时，watchers 也进行更新
	watchers map[*BaseCoord]struct{}

	delWatcherNodeNum int
	updating          int
}

func (n *EntityCoord) String() string {
	return fmt.Sprintf("<EntityCoord(%s)> watchersCount:%d x:%.2f y:%.2f z:%.2f oldx:%.2f oldy:%.2f oldz:%.2f",
		n.BaseCoord.entity.AoiID(), len(n.watchers), n.pos.X, n.pos.Y, n.pos.Z, n.old.X, n.old.Y, n.old.Z)
}

// 打印x，方便调试
func (n *EntityCoord) xString() string {
	return fmt.Sprintf("<EntityCoord(%s)> x:%.2f oldx:%.2f ",
		n.BaseCoord.entity.AoiID(), n.pos.X, n.old.X)
}

// 打印y，方便调试
func (n *EntityCoord) yString() string {
	return fmt.Sprintf("<EntityCoord(%s)> y:%.2f oldy:%.2f ",
		n.BaseCoord.entity.AoiID(), n.pos.Y, n.old.Y)
}

// 打印z，方便调试
func (n *EntityCoord) zString() string {
	return fmt.Sprintf("<EntityCoord(%s)> z:%.2f oldz:%.2f ",
		n.BaseCoord.entity.AoiID(), n.pos.Z, n.old.Z)
}

// NewEntityNode ctor
func NewEntityNode(entity Entityer) *EntityCoord {
	e := &EntityCoord{
		BaseCoord:         newBaseCoordWithEntity(entity),
		watchers:          make(map[*BaseCoord]struct{}, 4),
		delWatcherNodeNum: 0,
		updating:          0,
	}
	e.BaseCoord.delegate = e
	e.addFlags(nodeFlagEntity)

	return e
}

// // 实体节点默认没有扩展坐标，扩展坐标都等于实体实际的位置
// func (n *EntityCoord) xx() float32 {
// 	if n.entity == nil || n.hasFlags(nodeFlagRemoved|nodeFlagRemoving) {
// 		return -math.MaxFloat32
// 	}
// 	return n.entity.GetPosition().X
// }
// func (n *EntityCoord) yy() float32 {
// 	if n.entity == nil || n.hasFlags(nodeFlagRemoved|nodeFlagRemoving) {
// 		return -math.MaxFloat32
// 	}
// 	return n.entity.GetPosition().Y
// }
// func (n *EntityCoord) zz() float32 {
// 	if n.entity == nil || n.hasFlags(nodeFlagRemoved|nodeFlagRemoving) {
// 		return -math.MaxFloat32
// 	}
// 	return n.entity.GetPosition().Z
// }

func (n *EntityCoord) addWatcherNode(node *BaseCoord) bool {
	_, ok := n.watchers[node]
	if ok {
		return false
	}
	n.watchers[node] = struct{}{}

	// e.onAddWatcherNode(node)
	return true
}

// func (e *EntityCoord) onAddWatcherNode(node *BaseCoord) {}

func (n *EntityCoord) delWatcherNode(node *BaseCoord) bool {
	delete(n.watchers, node)
	return true
}

// Update the coord
func (n *EntityCoord) Update() {
	// 下面三行 参考 kbe 的注释
	// entity_coordinate_node.cpp line:347
	// 在这里做一下更新的原因是，很可能在CoordinateNode::Update()的过程中导致实体位置被移动
	// 而导致次数update被调用，在某种情况下会出现问题
	// 例如：// A->B, B-A（此时old_*是B）, A->B（此时old_*是B，而xx等目的地就是B）,此时update中会误判为没有移动。
	// https://github.com/kbengine/kbengine/issues/407
	// n.setOldx(n.pos.X)
	// n.setOldy(n.pos.Y)
	// n.setOldz(n.pos.Z)

	n.addFlags(nodeFlagEntityNodeUpdating)
	n.updating++

	if n.system != nil {
		n.system.Update(n.BaseCoord, n.pos.X, n.pos.Z)
	}

	for watcher := range n.watchers {
		watcher.delegate.Update()
	}

	n.updating--

	if n.updating == 0 {
		n.removeFlags(nodeFlagEntityNodeUpdating)
	}

	// e.clearDelWatcherNodes()
}

func (n *EntityCoord) resetOld() {
	n.old.X = n.pos.X
	n.old.Y = n.pos.Y
	n.old.Z = n.pos.Z
}

// ResetFlags 这里 实体节点 移除坐标系统后，再进入坐标系统时，需要重置一下flag
func (n *EntityCoord) ResetFlags() {
	n.removeFlags(nodeFlagRemoved | nodeFlagRemoving)
}

// 被节点管理器移除了
func (n *EntityCoord) onSystemRemoved() {
	n.old.X = -math.MaxFloat32
	n.old.Y = -math.MaxFloat32
	n.old.Z = -math.MaxFloat32

	w := n.entity.Witness()

	for other := range w.InterestIn {
		// 告诉其他实体，我不关注他了
		if other.Witness() != nil {
			other.Witness().DelInterestBy(w.entity)
		}
		w.DelInterestIn(other)

		// 这里考虑是否可以不发回调回去，因为玩家已经离开场景了，前端已经清理了
		// 或者额外加个参数，告诉业务方，这时候的离开视野 是当前玩家离开场景了，还是普通的别人离开视野
		// w.entity.OnLeaveAOI(other)
	}

	for other := range w.InterestedBy {
		// 告诉其他实体，我要离开视野了
		other.Witness().leaveView(n.entity)
	}
}

// func (e *EntityCoord) clearDelWatcherNodes() {
// 	if e.hasFlags(nodeFlagEntityNodeUpdating | nodeFlagRemoved | nodeFlagRemoving) {
// 		return
// 	}

// 	if e.delWatcherNodeNum > 0 {

// 	}
// }
