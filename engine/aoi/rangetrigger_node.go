package aoi

import (
	"fmt"
	"math"

	"github.com/tutumagi/pitaya/engine/math32"
)

// RangeTriggerNode struct
//	扩展坐标，只有在视距半径大于0的情况下，才会有RangeTriggerNode
//	比如此坐标绑定的实体A，实体A有矩形视距，视距为m
//	则在视距范围的触发器内 会额外有两个坐标
//	一个是正向坐标 此时(x, y, z) 对应为（x+m, y+m, z+m）
//	另外一个是负向坐标 此时(x, y, z) 对应为 (x-m, y-m, z-m)
//	那么 在有实体位置变化时，如果经过了此坐标
//	则会在 x,y,z 三个方向分别判定 [x-m, x+m], [y-m, y+m], [z-m, z+m] 的范围进行 aoi的判定处理
type RangeTriggerNode struct {
	*BaseCoord

	trigger   *RangeTrigger
	oldRadius float32
	radius    float32

	absRadius float32
}

func (r *RangeTriggerNode) String() string {
	f := "-"
	if r.hasFlags(nodeFlagPositiveBoundary) {
		f = "+"
	}
	return fmt.Sprintf("<RangeCoord(%s%s)> radius:%.2f x:%.2f y:%.2f z:%.2f oldx:%.2f oldy:%.2f oldz:%.2f",
		f, r.trigger.origin.entity.AoiID(), r.radius, r.pos.X, r.pos.Y, r.pos.Z, r.old.X, r.old.Y, r.old.Z)
}

// 打印x，方便调试
func (r *RangeTriggerNode) xString() string {
	f := "-"
	if r.hasFlags(nodeFlagPositiveBoundary) {
		f = "+"
	}
	return fmt.Sprintf("<RangeCoord(%s%s)> radius:%.2f x:%.2f oldx:%.2f ",
		f, r.trigger.origin.entity.AoiID(), r.radius, r.pos.X, r.old.X)
}

// 打印y，方便调试
func (r *RangeTriggerNode) yString() string {
	f := "-"
	if r.hasFlags(nodeFlagPositiveBoundary) {
		f = "+"
	}
	return fmt.Sprintf("<RangeCoord(%s%s)> radius:%.2f y:%.2f oldy:%.2f ",
		f, r.trigger.origin.entity.AoiID(), r.radius, r.pos.Y, r.old.Y)
}

// 打印z，方便调试
func (r *RangeTriggerNode) zString() string {
	f := "-"
	if r.hasFlags(nodeFlagPositiveBoundary) {
		f = "+"
	}
	return fmt.Sprintf("<RangeCoord(%s%s)> radius:%.2f z:%.2f oldz:%.2f ",
		f, r.trigger.origin.entity.AoiID(), r.radius, r.pos.Z, r.old.Z)
}

func newRangeTriggerNode(rangeTrigger *RangeTrigger, radius float32, positiveBoundary bool) *RangeTriggerNode {

	r := &RangeTriggerNode{
		BaseCoord: newBaseCoord(nil),
		trigger:   rangeTrigger,
		oldRadius: radius,
		radius:    radius,
		absRadius: math32.Abs(radius),
	}

	r.BaseCoord.delegate = r

	// 正向坐标和负向坐标，都属于隐藏的坐标，隐藏的坐标不会触发各种回调
	if positiveBoundary {
		r._flags = nodeFlagHide | nodeFlagPositiveBoundary
		r.weight = 3
	} else {
		r._flags = nodeFlagHide | nodeFlagNegativeBoundary
		r.weight = 2
	}

	r.trigger.originCoord().addWatcherNode(r.BaseCoord)

	return r
}

func (r *RangeTriggerNode) onTriggerUninstall() {
	if r.trigger.originCoord() != nil {
		r.trigger.originCoord().delWatcherNode(r.BaseCoord)
	}

	r.trigger = nil
}

func (r *RangeTriggerNode) onSystemRemoved() {
	r.old.X = -math.MaxFloat32
	r.old.Y = -math.MaxFloat32
	r.old.Z = -math.MaxFloat32
	r.onRemove()
}

func (r *RangeTriggerNode) onRemove() {

	// 既然自己都要删除了，通知 trigger 卸载
	if r.trigger != nil {
		r.trigger.uninstall()
	}
}

func (r *RangeTriggerNode) onParentRemove(parentNode *RangeTriggerNode) {
	// 既然自己都要删除了，通知 trigger 卸载
	if r.trigger != nil {
		r.trigger.uninstall()
	}
}

// func (r *RangeTriggerNode) xx() float32 {
// 	if r.hasFlags(nodeFlagRemoved|nodeFlagRemoving) || r.trigger == nil {
// 		return -math.MaxFloat32
// 	}

// 	return r.trigger.originCoord().pos.X + r.radius
// }

// func (r *RangeTriggerNode) yy() float32 {
// 	if r.hasFlags(nodeFlagRemoved|nodeFlagRemoving) || r.trigger == nil {
// 		return -math.MaxFloat32
// 	}

// 	return r.trigger.originCoord().pos.Y + r.radius
// }

// func (r *RangeTriggerNode) zz() float32 {
// 	if r.hasFlags(nodeFlagRemoved|nodeFlagRemoving) || r.trigger == nil {
// 		return -math.MaxFloat32
// 	}

// 	return r.trigger.originCoord().pos.Z + r.radius
// }

func (r *RangeTriggerNode) onNodePassX(node *BaseCoord, isFront bool) {
	if !r.hasFlags(nodeFlagRemoved) && r.trigger != nil {
		r.trigger.onNodePassX(r, node, isFront)
	}
}

func (r *RangeTriggerNode) onNodePassZ(node *BaseCoord, isFront bool) {
	if !r.hasFlags(nodeFlagRemoved) && r.trigger != nil {
		r.trigger.onNodePassZ(r, node, isFront)
	}
}

func (r *RangeTriggerNode) setRange(radius float32) {
	r.radius = radius
	r.absRadius = math32.Abs(radius)
}

func (r *RangeTriggerNode) setOldRange(radius float32) {
	r.oldRadius = radius
}

// Update the node
func (r *RangeTriggerNode) Update() {
	if r.system != nil {
		newX := float32(-math.MaxFloat32)
		newZ := float32(-math.MaxFloat32)
		if !(r.hasFlags(nodeFlagRemoved|nodeFlagRemoving) || r.trigger == nil) {
			originPos := r.trigger.originCoord().pos
			newX = originPos.X + r.radius
			newZ = originPos.Z + r.radius
		}

		r.system.Update(r.BaseCoord, newX, newZ)
	}
}

func (r *RangeTriggerNode) wasInXRange(node *BaseCoord) bool {
	originX := r.old.X - r.oldRadius
	lowerBound := originX - math32.Abs(r.oldRadius)
	upperBound := originX + math32.Abs(r.oldRadius)

	return node.old.X >= lowerBound && node.old.X <= upperBound
}

func (r *RangeTriggerNode) wasInZRange(node *BaseCoord) bool {
	originZ := r.old.Z - r.oldRadius
	lowerBound := originZ - math32.Abs(r.oldRadius)
	upperBound := originZ + math32.Abs(r.oldRadius)

	return node.old.Z >= lowerBound && node.old.Z <= upperBound
}

func (r *RangeTriggerNode) isInZRange(node *BaseCoord) bool {
	originZ := r.trigger.originCoord().pos.Z
	lowerBound := originZ - r.absRadius
	upperBound := originZ + r.absRadius

	zz := node.pos.Z
	return zz >= lowerBound && zz <= upperBound
}

func (r *RangeTriggerNode) isInXRange(node *BaseCoord) bool {
	originX := r.trigger.originCoord().pos.X
	lowerBound := originX - r.absRadius
	upperBound := originX + r.absRadius

	xx := node.pos.X
	return xx >= lowerBound && xx <= upperBound
}

func (r *RangeTriggerNode) resetOld() {
	r.old.X = r.pos.X
	r.old.Y = r.pos.Y
	r.old.Z = r.pos.Z

	r.oldRadius = r.radius
}
