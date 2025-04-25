package fsm

import (
	"errors"
)

// Error constants - standard error definitions
var (
	ErrStateMachineNotFound     = errors.New("state machine not found")
	ErrStateMachineAlreadyExist = errors.New("state machine already exists")
	ErrStateNotFound            = errors.New("state not found")
	ErrEventNotAccepted         = errors.New("event not accepted in state")
	ErrTransitionNotFound       = errors.New("no transition found")
	ErrConditionNotMet          = errors.New("transition conditions not met")
	ErrActionExecutionFailed    = errors.New("action execution failed")
	ErrStateMachineNotBuilt     = errors.New("state machine is not built yet")
	ErrInternalTransition       = errors.New("internal transition source and target states must be the same")
)
