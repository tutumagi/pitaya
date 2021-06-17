package fsm

import (
	"testing"

	. "github.com/go-playground/assert/v2"
)

const (
	Off StateType = 1
	On  StateType = 2
)

type OffAction struct{}

func (a *OffAction) Execute(ctx ...StateContext) StateType {
	return Default
}

func (a *OffAction) Tick(dtms int32, ctx StateContext) {
}

type OnAction struct{}

func (a *OnAction) Execute(ctx ...StateContext) StateType {
	return Default
}

func (a *OnAction) Tick(dtms int32, ctx StateContext) {
}

func TestLightFSM(t *testing.T) {
	fsm := &StateMachine{
		States: States{
			Default: NewState(nil, Off),
			Off:     NewState(&OffAction{}, On),
			On:      NewState(&OnAction{}, Off),
		},
	}

	err := fsm.EnterState(Off, nil)
	Equal(t, err, nil)

	err = fsm.EnterState(Off, nil)
	Equal(t, err, ErrStateReject)

	err = fsm.EnterState(On, nil)
	Equal(t, err, nil)

	err = fsm.EnterState(On, nil)
	Equal(t, err, ErrStateReject)

	err = fsm.EnterState(Off, nil)
	Equal(t, err, nil)
}

func TestNoActionFSM(t *testing.T) {
	fsm := &StateMachine{
		States: States{
			Default: NewState(nil, On),
			On:      NewState(nil, Off),
			Off:     NewState(nil, On),
		},
	}

	err := fsm.EnterState(Off, nil)
	Equal(t, err, ErrStateReject)

	err = fsm.EnterState(On, nil)
	Equal(t, err, ErrStateNoAction)
}
