package aoi

import (
	"math"

	"github.com/tutumagi/pitaya/logger"
	"github.com/tutumagi/pitaya/engine/math32"
)

type nodeFlag int16

const (
	nodeFlagUnknown            nodeFlag = 0x0
	nodeFlagEntity             nodeFlag = 0x1  // 一个enitty节点
	nodeFlagTrigger            nodeFlag = 0x2  // 一个触发器节点
	nodeFlagHide               nodeFlag = 0x4  // 隐藏节点（其他节点不可见）
	nodeFlagRemoving           nodeFlag = 0x8  // 删除中的节点
	nodeFlagRemoved            nodeFlag = 0x10 // 删除节点
	nodeFlagPending            nodeFlag = 0x20 // 这类节点处于update操作中
	nodeFlagEntityNodeUpdating nodeFlag = 0x40 // entity节点正在执行update操作
	nodeFlagInstalling         nodeFlag = 0x80 // 节点正在安装操作

	// 正边界和负边界的节点的含义是，当 aoi 视距是矩形时，aoi 位置在 x,y，视距半径为m，则如 rangeview.drawio 中所示
	nodeFlagPositiveBoundary nodeFlag = 0x100 // 节点是触发器的正边界
	nodeFlagNegativeBoundary nodeFlag = 0x200 // 节点是触发器的负边界

	nodeFlagHideOrRemoved nodeFlag = nodeFlagRemoved | nodeFlagHide

	nodeFlagBoundary nodeFlag = nodeFlagPositiveBoundary | nodeFlagNegativeBoundary
)

type _BaseCoordDelegate interface {
	// 位置移动时，需要调用此方法，更新在System中的位置
	Update()

	// 当 other 在x轴 跨过 当前节点时，处理关心/被关心的信息
	onNodePassX(other *BaseCoord, isFront bool)
	// 当 other 在z轴 跨过 当前节点时，处理关心/被关心的信息
	onNodePassZ(other *BaseCoord, isFront bool)

	onSystemRemoved()

	// 将old设置为新的值
	resetOld()

	// for debug
	xString() string
	yString() string
	zString() string
}

// BaseCoord 基础坐标节点，实现 Coorder 接口
type BaseCoord struct {
	// 旧的节点位置
	old *math32.Vector3
	// 节点的实际位置
	pos *math32.Vector3

	weight int8

	_flags nodeFlag

	// 节点所在的坐标系统
	system Systemer

	// 节点绑定的实体
	entity Entityer

	// 在十字链表里面的节点
	xzNode *node

	// go 没有标准的oop，使用代理来处理不同`子类`的实例方法
	delegate _BaseCoordDelegate
}

func newBaseCoord(system Systemer) *BaseCoord {
	return &BaseCoord{
		old:    &math32.Vector3{X: -math.MaxFloat32, Y: -math.MaxFloat32, Z: -math.MaxFloat32},
		pos:    &math32.Vector3{X: -math.MaxFloat32, Y: -math.MaxFloat32, Z: -math.MaxFloat32},
		system: system,
		entity: nil,
		weight: 0,
		_flags: nodeFlagUnknown,
	}
}

func newBaseCoordWithEntity(entity Entityer) *BaseCoord {
	return &BaseCoord{
		old:    &math32.Vector3{X: -math.MaxFloat32, Y: -math.MaxFloat32, Z: -math.MaxFloat32},
		pos:    &math32.Vector3{X: -math.MaxFloat32, Y: -math.MaxFloat32, Z: -math.MaxFloat32},
		system: nil,
		entity: entity,
		weight: 0,
		_flags: nodeFlagUnknown,
	}
}

// func (n *BaseCoord) String() string {
// 	return fmt.Sprintf("<BaseCoord> x:%.2f y:%.2f z:%.2f xx:%.2f yy:%.2f zz:%.2f",
// 		n.pos.X, n.pos.Y, n.pos.Z, n.xx(), n.yy(), n.zz())
// }

// System 返回所在的坐标系统
func (n *BaseCoord) System() Systemer {
	return n.system
}

func (n *BaseCoord) setSystem(s Systemer) {
	if n.system != nil && s != nil {
		logger.Debugf("%s coord already has a system", n.entity)
		return
	}
	n.system = s
}

// Position 返回位置
func (n *BaseCoord) Position() *math32.Vector3 {
	return n.pos
}

// SetPosition 设置位置
func (n *BaseCoord) SetPosition(x float32, y float32, z float32) {
	n.pos.X = x
	n.pos.Y = y
	n.pos.Z = z
}

// SetVec3 设置位置
func (n *BaseCoord) SetVec3(vec3 *math32.Vector3) {
	n.SetPosition(vec3.X, vec3.Y, vec3.Z)
}

// SetVec2 设置位置
func (n *BaseCoord) SetVec2(vec2 *math32.Vector2) {
	n.SetPosition(vec2.X, 0, vec2.Y)
}

func (n *BaseCoord) hasFlags(flag nodeFlag) bool {
	return n._flags&flag > 0
}

func (n *BaseCoord) addFlags(flag nodeFlag) {
	n._flags |= flag
}

func (n *BaseCoord) removeFlags(flag nodeFlag) {
	n._flags &= ^flag
}

func (n *BaseCoord) flags() nodeFlag {
	return n._flags
}

func (n *BaseCoord) onNodePassX(other *BaseCoord, isFront bool) {}
func (n *BaseCoord) onNodePassZ(other *BaseCoord, isFront bool) {}

func (n *BaseCoord) node() *node {
	return n.xzNode
}

func (n *BaseCoord) setNode(xzNode *node) {
	// CoordSystem里面做了检查，这里不再做检查
	// if n.xzNode != nil {
	// 	return
	// }
	n.xzNode = xzNode
}

// func (n *CoordinateNode) resetOld() {
// 	n.oldpos.X = n.xx()
// 	n.oldpos.Y = n.xx()
// 	n.oldpos.Z = n.zz()
// }

// x轴 降序排列
type _SortByCoordX []Entityer

func (a _SortByCoordX) Len() int           { return len(a) }
func (a _SortByCoordX) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a _SortByCoordX) Less(i, j int) bool { return a[j].Coord().pos.X < a[i].Coord().pos.X }

// z轴 降序排列
type _SortByCoordZ []Entityer

func (a _SortByCoordZ) Len() int           { return len(a) }
func (a _SortByCoordZ) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a _SortByCoordZ) Less(i, j int) bool { return a[j].Coord().pos.Z < a[i].Coord().pos.Z }
