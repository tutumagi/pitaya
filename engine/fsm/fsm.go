package fsm

import (
	"fmt"

	"github.com/tutumagi/pitaya/logger"

	e "github.com/tutumagi/pitaya/errors"
)

// ErrStateReject 状态机错误
var ErrStateReject = e.NewError(fmt.Errorf("状态机切换错误"), "FSM_001")

// ErrStateNoAction 没有action
var ErrStateNoAction = e.NewError(fmt.Errorf("该状态没有Action错误"), "FSM_002")

const (
	// Default 默认状态
	Default StateType = 0
)

// StateType 状态类型
type StateType int32

// StateContext 触发状态操作时的传递给对应Action的参数类型
type StateContext interface{}

// Action 进入某一个状态时触发的操作
type Action interface {
	Execute(ctx ...StateContext) StateType
	Tick(dtms int32, ctx StateContext) // 帧循环
}

// State 这是一个状态，进入该状态时会触发Action操作
type State struct {
	Action Action
	// 可以切换的状态列表
	States map[StateType]struct{}
}

// NewState 新建一个状态
func NewState(act Action, states ...StateType) State {
	s := State{
		Action: act,
		States: map[StateType]struct{}{},
	}
	for _, state := range states {
		s.States[state] = struct{}{}
	}
	return s
}

// States 状态类型对应状态结构
type States map[StateType]State

// StateMachine state machine
type StateMachine struct {
	Prev   StateType // 上一个状态
	Cur    StateType // 当前状态
	States States    // 所有的状态

	StateChange func(curState StateType)
}

// getNextState 获取下一个状态
func (s *StateMachine) getNextState(next StateType) (StateType, error) {
	// 拿到当前的state
	if state, ok := s.States[s.Cur]; ok {
		// 根据当前的state 可以切换的状态列表，获取event对应可切换的状态类型
		if state.States != nil {
			if _, ok := state.States[next]; ok {
				return next, nil
			}
		}
	}
	// 获取失败，返回
	return Default, ErrStateReject
}

// EnterState 给状态机发送事件，触发下一次状态
func (s *StateMachine) EnterState(stateTyp StateType, ctx ...StateContext) error {
	// s.mutex.Lock()
	// defer s.mutex.Unlock()

	for {
		// logger.Debugf("enter state %s", stateTyp)
		// 获取下一次状态类型
		nextStateTyp, err := s.getNextState(stateTyp)
		if err != nil {
			return ErrStateReject
		}

		// 拿到下一次状态类型对应的状态数据
		state, ok := s.States[nextStateTyp]
		if !ok || state.Action == nil {
			logger.Error("must imp action")
			return ErrStateNoAction
		}

		s.Prev = s.Cur
		s.Cur = nextStateTyp

		// 状态变化后回调出去
		if s.Prev != s.Cur && s.StateChange != nil {
			s.StateChange(s.Cur)
		}
		// logger.Debugf("execute state %s", nextStateTyp)
		// 执行切换到此状态后对应的操作
		nextNextStateTyp := state.Action.Execute(ctx...)

		if nextNextStateTyp == s.Cur || nextNextStateTyp == Default {
			return nil
		}

		stateTyp = nextNextStateTyp
	}

}

// Tick 帧循环
func (s *StateMachine) Tick(dtms int32, ctx StateContext) {
	if stat, ok := s.States[s.Cur]; ok {
		stat.Action.Tick(dtms, ctx)
	}
}
