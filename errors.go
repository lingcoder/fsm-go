package fsm

import (
	"fmt"
	"github.com/pkg/errors"
)

// Error constants
const (
	// Error messages
	ErrStateMachineNotFound     = "state machine not found"
	ErrStateMachineAlreadyExist = "state machine already exists"
	ErrStateNotFound            = "state not found"
	ErrEventNotAccepted         = "event not accepted in state"
	ErrTransitionNotFound       = "no transition found"
	ErrConditionNotMet          = "transition conditions not met"
	ErrActionExecutionFailed    = "action execution failed"
)

// Error creation helper functions
func StateMachineNotFound(id string) error {
	return fmt.Errorf("%s: %s", ErrStateMachineNotFound, id)
}

func StateMachineAlreadyExist(id string) error {
	return fmt.Errorf("%s: %s", ErrStateMachineAlreadyExist, id)
}

func StateNotFound(state interface{}) error {
	return fmt.Errorf("%s: %v", ErrStateNotFound, state)
}

func EventNotAccepted(state, event interface{}) error {
	return fmt.Errorf("%s: event=%v, state=%v", ErrEventNotAccepted, event, state)
}

func TransitionNotFound(state, event interface{}) error {
	return fmt.Errorf("%s: event=%v, state=%v", ErrTransitionNotFound, event, state)
}

func ConditionNotMet(state, event interface{}) error {
	return fmt.Errorf("%s: event=%v, state=%v", ErrConditionNotMet, event, state)
}

func ActionExecutionError(err error) error {
	return errors.Wrap(err, ErrActionExecutionFailed)
}
