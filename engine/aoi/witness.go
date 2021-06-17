package aoi

import "github.com/tutumagi/pitaya/logger"

// Witness 实体A的观察者，用来处理 其他实体进入/离开A的视距范围
//	以及维护了 被哪些实体关心的列表和关心哪些实体的列表
type Witness struct {
	entity Entityer

	// 视距半径
	radius  float32
	trigger *ViewTrigger

	// 一个滞后范围
	viewHysteresisArea    float32
	hysteresisAreaTrigger *ViewTrigger

	// 关心的实体
	InterestIn map[Entityer]struct{}
	// 被哪些实体关心
	InterestedBy map[Entityer]struct{}
}

// NewWitness ctor
func NewWitness() *Witness {
	w := &Witness{}
	w.entity = nil

	w.radius = 0
	w.trigger = nil

	w.viewHysteresisArea = 0
	w.hysteresisAreaTrigger = nil

	w.InterestedBy = make(map[Entityer]struct{})
	w.InterestIn = make(map[Entityer]struct{})

	return w
}

func (w *Witness) onEnterView(trigger *ViewTrigger, other Entityer) {
	// 如果进入的是 hysteresis 区域，则不产生作用
	if w.trigger == w.hysteresisAreaTrigger {
		return
	}

	// 有实体进入视野
	w.enterView(other)
}

func (w *Witness) enterView(other Entityer) {
	// 告诉其他实体被关注了
	other.Witness().AddInterestBy(w.entity)
	w.AddInterestIn(other)

	w.entity.OnEnterAOI(other)
}

func (w *Witness) onLeaveView(trigger *ViewTrigger, other Entityer) {
	// 如果设置过 滞后 区域，则离开滞后区域 才算离开了view
	if w.hysteresisAreaTrigger != nil && w.hysteresisAreaTrigger != w.trigger {
		return
	}

	w.leaveView(other)
}

// other 离开当前视野
func (w *Witness) leaveView(other Entityer) {
	// 告诉其他实体，当前实体不关注他了
	if other.Witness() != nil {
		other.Witness().DelInterestBy(w.entity)
	}

	w.DelInterestIn(other)

	w.entity.OnLeaveAOI(other)
}

// SetViewRadius 设置视距半径
func (w *Witness) SetViewRadius(radius float32, hyst float32) {
	w.radius = radius
	w.viewHysteresisArea = hyst

	// 由于位置同步使用了相对位置压缩传输，可用范围为-512~512之间，因此超过范围将出现同步错误
	// 这里做一个限制，如果需要过大的数值客户端应该调整坐标单位比例，将其放大使用。
	// 参考: MemoryStream::appendPackXZ
	if w.radius+w.viewHysteresisArea > 512 {
		// 参考KBE witness.cpp  line:290
	}

	if w.radius > 0 && w.entity != nil {
		if w.trigger == nil {
			w.trigger = newViewTrigger(w.entity.Coord(), w.radius)

			// 如果实体已经在场景中，那么需要安装
			if w.entity.Coord().System() != nil {
				w.trigger.install()
			}
		} else {
			w.trigger.update(radius)

			// 如果实体已经在场景中，那么需要安装
			if !w.trigger.isInstalled() && w.entity.Coord().System() != nil {
				w.trigger.reinstall(w.entity.Coord())
			}
		}

		if w.viewHysteresisArea > 0.01 && w.entity != nil {
			if w.hysteresisAreaTrigger == nil {
				w.hysteresisAreaTrigger = newViewTrigger(w.entity.Coord(), w.viewHysteresisArea+w.radius)

				// 如果实体已经在场景中，那么需要安装
				if w.entity.Coord().System() != nil {
					w.hysteresisAreaTrigger.install()
				}
			} else {
				w.hysteresisAreaTrigger.update(w.viewHysteresisArea + w.radius)

				// 如果实体已经在场景中，那么需要安装
				if !w.hysteresisAreaTrigger.isInstalled() && w.entity.Coord().System() != nil {
					w.hysteresisAreaTrigger.reinstall(w.entity.Coord())
				}
			}
		} else {
			// 注意：此处如果不销毁pViewHysteresisAreaTrigger_则必须是update
			// 因为离开View的判断如果pViewHysteresisAreaTrigger_存在，那么必须出了pViewHysteresisAreaTrigger_才算出View
			if w.hysteresisAreaTrigger != nil {
				w.hysteresisAreaTrigger.update(w.viewHysteresisArea + w.radius)
			}
		}
	} else {
		w.UninstallViewTrigger()
	}
}

// InstallViewTrigger 安装视距触发器
func (w *Witness) InstallViewTrigger() {
	if w.trigger != nil {
		// 在设置视距半径为0后，掉线重登录会出现这种情况
		if w.radius <= 0 {
			return
		}

		if w.hysteresisAreaTrigger != nil && w.entity != nil {
			w.hysteresisAreaTrigger.reinstall(w.entity.Coord())
		}

		if w.entity != nil {
			w.trigger.reinstall(w.entity.Coord())
		}
	} else {
		if w.hysteresisAreaTrigger != nil {
			logger.Warnf("trigger is nil but hysteresisAreaTrigger is not nil")
		}
	}
}

// UninstallViewTrigger 卸载数据触发器
func (w *Witness) UninstallViewTrigger() {
	if w.trigger != nil {
		w.trigger.uninstall()
	}

	if w.hysteresisAreaTrigger != nil {
		w.hysteresisAreaTrigger.uninstall()
	}

	// 所有关心的实体 离开当前视距
	for other := range w.InterestIn {
		w.leaveView(other)
	}
}

/************************* Attach / Detach *******************************/

// Attach entity
func (w *Witness) Attach(entity Entityer) {
	w.entity = entity

	// TODO 设置默认的视距半径
	// w.setViewRadius(5, 0)

	w.onAttach(entity)
}

func (w *Witness) onAttach(entity Entityer) {
	// kbe 中是 通知客户端有玩家进入当前世界了
}

// Detach entity
func (w *Witness) Detach(entity Entityer) {
	// kbe 中是 通知客户端有玩家离开当前世界了

	w.clear(entity)
}

func (w *Witness) clear(entity Entityer) {
	w.UninstallViewTrigger()

	for other := range w.InterestedBy {
		// 告诉其他实体，当前实体不关注他了
		if other.Witness() != nil {
			other.Witness().DelInterestBy(w.entity)
		}
	}

	// kbe 中说，这里如果销毁了 trigger，一方面会影响复用，另一方面可能会crash
	// 不需要销毁，后面还可以重用
	// 此处销毁可能会产生错误，因为enterview过程中可能导致实体销毁
	// 见 kbe witness.cpp line:213
	w.reset()
}

func (w *Witness) reset() {
	w.entity = nil

	w.radius = 0
	w.trigger = nil

	w.viewHysteresisArea = 0
	w.hysteresisAreaTrigger = nil

	w.InterestedBy = make(map[Entityer]struct{})
	w.InterestIn = make(map[Entityer]struct{})
}

/************************* interest in/by *********************************/

// AddInterestIn 添加要关心的实体
func (w *Witness) AddInterestIn(other Entityer) {
	w.InterestIn[other] = struct{}{}
}

// DelInterestIn 移除要关心的实体
func (w *Witness) DelInterestIn(other Entityer) {
	delete(w.InterestIn, other)
}

// AddInterestBy 添加被谁关心
func (w *Witness) AddInterestBy(other Entityer) {
	w.InterestedBy[other] = struct{}{}
}

// DelInterestBy 移除被谁关心
func (w *Witness) DelInterestBy(other Entityer) {
	delete(w.InterestedBy, other)
}
