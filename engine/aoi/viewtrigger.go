package aoi

// ViewTrigger 视野触发器
type ViewTrigger struct {
	*RangeTrigger

	witness *Witness
}

func newViewTrigger(origin *EntityCoord, radius float32) *ViewTrigger {
	r := &ViewTrigger{}
	r.RangeTrigger = newRangeTrigger(origin, radius, r)
	r.witness = origin.entity.Witness()

	return r
}

func (r *ViewTrigger) onEnter(node *BaseCoord) {
	if node.flags()&nodeFlagEntity <= 0 {
		return
	}

	// TODO 判断 entity 有没有client？
	r.witness.onEnterView(r, node.entity)
}

func (r *ViewTrigger) onLeave(node *BaseCoord) {
	if node.flags()&nodeFlagEntity <= 0 {
		return
	}

	// TODO 判断 entity 有没有client？
	r.witness.onLeaveView(r, node.entity)
}

// Witness 获取witness
func (r *ViewTrigger) Witness() *Witness {
	return r.witness
}
