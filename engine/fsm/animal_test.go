package fsm

import (
	"testing"

	. "github.com/go-playground/assert/v2"
)

const (
	// 默认状态
	Idle StateType = 1
	// 闲着走路状态
	Walk StateType = 2
	// 被喂东西状态
	BeFeed StateType = 3
	// 被捕获状态
	Hunted StateType = 4
)

type IdleAction struct{}

func (a *IdleAction) Execute(ctx ...StateContext) StateType {
	return Default
}

func (a *IdleAction) Tick(dtms int32, ctx StateContext) {
}

type WalkAction struct{}

func (a *WalkAction) Execute(ctx ...StateContext) StateType {
	return Default
}
func (a *WalkAction) Tick(dtms int32, ctx StateContext) {
}

type BeFeedAction struct{}

func (a *BeFeedAction) Execute(ctx ...StateContext) StateType {
	ani := ctx[0].(*Animal)
	ani.hp -= 20
	if ani.hp <= 0 {
		ani.hp = 0
		return Hunted
	}
	if ani.roundPeople == 0 {
		return Idle
	}
	return Walk
}
func (a *BeFeedAction) Tick(dtms int32, ctx StateContext) {
}

type HuntAction struct{}

func (a *HuntAction) Execute(ctx ...StateContext) StateType {
	return Default
}
func (a *HuntAction) Tick(dtms int32, ctx StateContext) {
}

type Animal struct {
	hp  int32
	fsm *StateMachine

	roundPeople int32 // 周围的人
}

func TestHuntFSM(t *testing.T) {
	animal := &Animal{
		hp: 50,
		fsm: &StateMachine{
			States: States{
				Default: NewState(nil, Walk, BeFeed),
				Idle:    NewState(&IdleAction{}, Walk, BeFeed),
				Walk:    NewState(&WalkAction{}, BeFeed, Idle),
				BeFeed:  NewState(&BeFeedAction{}, Walk, Idle, Hunted),
				Hunted:  NewState(&HuntAction{}),
			},
		},
	}

	// 当前是默认状态（Idle），不需要再进入Idle状态
	err := animal.fsm.EnterState(Idle, nil)
	Equal(t, err, ErrStateReject)
	Equal(t, animal.fsm.Prev, Default)
	Equal(t, animal.fsm.Cur, Default)

	// 有人过来了
	animal.roundPeople++

	// 当前是Idle状态，可以进入行走状态
	err = animal.fsm.EnterState(Walk, nil)
	Equal(t, err, nil)
	Equal(t, animal.fsm.Prev, Default)
	Equal(t, animal.fsm.Cur, Walk)

	// 有人过来了
	animal.roundPeople++
	// 再次进入行走状态会拒绝
	err = animal.fsm.EnterState(Walk, nil)
	Equal(t, err, ErrStateReject)
	Equal(t, animal.fsm.Prev, Default)
	Equal(t, animal.fsm.Cur, Walk)

	// 每次投喂减20生命值
	err = animal.fsm.EnterState(BeFeed, animal)
	Equal(t, err, nil)
	Equal(t, animal.fsm.Prev, BeFeed)
	// 有人，所以是Walk状态
	Equal(t, animal.fsm.Cur, Walk)

	animal.roundPeople -= 2
	err = animal.fsm.EnterState(BeFeed, animal)
	Equal(t, err, nil)
	Equal(t, animal.fsm.Prev, BeFeed)
	// 没有人了，是idle状态
	Equal(t, animal.fsm.Cur, Idle)

	err = animal.fsm.EnterState(BeFeed, animal)
	Equal(t, err, nil)
	Equal(t, animal.fsm.Prev, BeFeed)
	Equal(t, animal.fsm.Cur, Hunted)

}
