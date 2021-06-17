package aoi

import "math"

// RangeTrigger 进出范围 触发器
type RangeTrigger struct {
	radius float32

	removing bool

	// 原始节点，实际类型 *EntityCoord
	origin *EntityCoord
	// 正向范围的位置节点
	positiveBoundary *RangeTriggerNode
	// 负向范围的位置节点
	negativeBoundary *RangeTriggerNode

	callback ViewCallback
}

func newRangeTrigger(origin *EntityCoord, radius float32, callback ViewCallback) *RangeTrigger {
	r := &RangeTrigger{}
	r.origin = origin
	r.radius = radius
	r.removing = false
	r.callback = callback

	return r
}

func (r *RangeTrigger) reinstall(origin *EntityCoord) bool {
	r.uninstall()
	r.origin = origin

	return r.install()
}

func (r *RangeTrigger) onEnter(node *BaseCoord) {
	r.callback.onEnter(node)
}

func (r *RangeTrigger) onLeave(node *BaseCoord) {
	r.callback.onLeave(node)
}

func (r *RangeTrigger) originCoord() *EntityCoord {
	return r.origin
}

// 更新视距范围
func (r *RangeTrigger) update(radius float32) {
	oldRadius := r.radius

	r.radius = radius

	if r.positiveBoundary != nil {
		r.positiveBoundary.setRange(radius)
		r.positiveBoundary.setOldRange(oldRadius)
		r.positiveBoundary.Update()
	}

	if r.negativeBoundary != nil {
		r.negativeBoundary.setRange(-radius)
		r.negativeBoundary.setOldRange(-oldRadius)
		r.negativeBoundary.Update()
	}
}

func (r *RangeTrigger) install() bool {
	if r.positiveBoundary == nil {
		r.positiveBoundary = newRangeTriggerNode(r, 0, true)
	} else {
		r.positiveBoundary.setRange(0)
	}

	if r.negativeBoundary == nil {
		r.negativeBoundary = newRangeTriggerNode(r, 0, false)
	} else {
		r.negativeBoundary.setRange(0)
	}

	r.positiveBoundary.addFlags(nodeFlagInstalling)
	r.negativeBoundary.addFlags(nodeFlagInstalling)

	/*
		注意：此处必须是先安装negativeBoundary_再安装positiveBoundary_，如果调换顺序则会导致View的BUG，例如：在一个实体enterView触发时销毁了进入View的实体
		此时实体销毁时并未触发离开View事件，而未触发View事件导致其他实体的View列表中引用的该销毁的实体是一个无效指针。

		原因如下：
		由于总是优先安装在positiveBoundary_，而边界在安装过程中导致另一个实体进入View了， 然后他在这个过程中可能销毁了， 而另一个边界negativeBoundary_还没有安装，
		而节点删除时会设置节点的xx为-FLT_MAX，让其向negativeBoundary_方向离开，所以positiveBoundary_不能检查到这个边界也就不会触发View离开事件。
	*/
	r.negativeBoundary.old.X = -math.MaxFloat32
	r.negativeBoundary.old.Y = -math.MaxFloat32
	r.negativeBoundary.old.Z = -math.MaxFloat32
	// 负向视距范围为 -radius
	r.negativeBoundary.setRange(-r.radius)
	r.negativeBoundary.setOldRange(-r.radius)

	// NOTE: 加入插入的参考点优化，如果不需要此优化，则注释此行，打开下面两行
	// r.origin.System().Insert(r.negativeBoundary.BaseCoord)
	// r.negativeBoundary.Update()
	r.origin.system.InsertWithRef(r.negativeBoundary.BaseCoord, r.origin.BaseCoord)

	// update 可能导致实体销毁简介导致自己被重置，此时返回安装失败
	if r.negativeBoundary == nil {
		return false
	}
	r.negativeBoundary.removeFlags(nodeFlagInstalling)

	r.positiveBoundary.old.X = math.MaxFloat32
	r.positiveBoundary.old.Y = math.MaxFloat32
	r.positiveBoundary.old.Z = math.MaxFloat32
	// 正向视距范围为 radius
	r.positiveBoundary.setRange(r.radius)
	r.positiveBoundary.setOldRange(r.radius)

	// NOTE: 加入插入的参考点优化，如果不需要此优化，则注释此行，打开下面两行
	// r.origin.System().Insert(r.positiveBoundary.BaseCoord)
	// r.positiveBoundary.Update()
	r.origin.system.InsertWithRef(r.positiveBoundary.BaseCoord, r.origin.BaseCoord)

	if r.positiveBoundary != nil {
		r.positiveBoundary.removeFlags(nodeFlagInstalling)
		return true
	}

	return false
}

func (r *RangeTrigger) uninstall() bool {
	if r.removing {
		return false
	}
	r.removing = true

	if r.positiveBoundary != nil && r.positiveBoundary.system != nil {
		r.positiveBoundary.system.Remove(r.positiveBoundary.BaseCoord)
		r.positiveBoundary.onTriggerUninstall()
	}

	if r.negativeBoundary != nil && r.negativeBoundary.system != nil {
		r.negativeBoundary.system.Remove(r.negativeBoundary.BaseCoord)
		r.negativeBoundary.onTriggerUninstall()
	}

	r.positiveBoundary = nil
	r.negativeBoundary = nil

	r.removing = false

	return true
}

func (r *RangeTrigger) onNodePassZ(triggerNode *RangeTriggerNode, node *BaseCoord, isFront bool) {
	if node == r.origin.BaseCoord {
		return
	}

	wasInZ := triggerNode.wasInZRange(node)
	isInZ := triggerNode.isInZRange(node)

	if wasInZ == isInZ {
		return
	}

	wasIn := triggerNode.wasInXRange(node) && wasInZ
	isIn := triggerNode.isInXRange(node) && isInZ

	if wasIn == isIn {
		return
	}

	if isIn {
		r.onEnter(node)
	} else {
		r.onLeave(node)
	}
}

func (r *RangeTrigger) onNodePassX(triggerNode *RangeTriggerNode, node *BaseCoord, isFront bool) {
	if node == r.origin.BaseCoord {
		return
	}

	wasInZ := triggerNode.wasInZRange(node)
	isInZ := triggerNode.isInZRange(node)

	// 如果 z 轴情况有变化，则处理 z 轴时在做处理，优先级为zyx，这样才可以保证只有一次enter或者leave
	if wasInZ != isInZ {
		return
	}

	wasIn := triggerNode.wasInXRange(node) && wasInZ
	isIn := triggerNode.isInXRange(node) && isInZ

	// 如果情况没有发生变化则忽略
	if wasIn == isIn {
		return
	}

	if isIn {
		r.onEnter(node)
	} else {
		r.onLeave(node)
	}
}

func (r *RangeTrigger) isInstalled() bool {
	return r.positiveBoundary != nil && r.negativeBoundary != nil
}
