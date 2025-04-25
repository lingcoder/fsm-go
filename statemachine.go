package fsm

import (
	"fmt"
	"github.com/pkg/errors"
	"sync"
)

// StateMachine is a generic state machine interface
// S: State type, must be comparable (e.g., string, int)
// E: Event type, must be comparable
// C: Context type, can be any type, used to pass data during state transitions
type StateMachine[S comparable, E comparable, C any] interface {
	// FireEvent triggers a state transition based on the current state and event
	// Returns the new state and any error that occurred
	FireEvent(sourceState S, event E, ctx C) (S, error)

	// ShowStateMachine returns a string representation of the state machine
	ShowStateMachine() string

	// GeneratePlantUML returns a PlantUML diagram of the state machine
	GeneratePlantUML() string
}

// Transition represents a state transition
type Transition[S comparable, E comparable, C any] struct {
	Source      S
	Target      S
	Event       E
	Conditions  []Condition[C]
	Actions     []Action[S, E, C]
	TransType   TransitionType
	Description string
}

// TransitionType defines the type of transition
type TransitionType int

const (
	// External transitions change the state from source to target
	External TransitionType = iota
	// Internal transitions don't change the state
	Internal
)

// Condition is an interface for transition conditions
type Condition[C any] interface {
	// IsSatisfied returns true if the condition is met
	IsSatisfied(ctx C) bool
}

// Action is an interface for transition actions
type Action[S comparable, E comparable, C any] interface {
	// Execute runs the action during a state transition
	Execute(from, to S, event E, ctx C) error
}

// StateMachineImpl is the implementation of StateMachine
type StateMachineImpl[S comparable, E comparable, C any] struct {
	id          string
	transitions map[S]map[E][]*Transition[S, E, C]
	mutex       sync.RWMutex
}

// NewStateMachine creates a new state machine
func NewStateMachine[S comparable, E comparable, C any](id string) *StateMachineImpl[S, E, C] {
	return &StateMachineImpl[S, E, C]{
		id:          id,
		transitions: make(map[S]map[E][]*Transition[S, E, C]),
	}
}

// FireEvent triggers a state transition based on the current state and event
func (sm *StateMachineImpl[S, E, C]) FireEvent(sourceState S, event E, ctx C) (S, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	// Check if the source state exists
	eventMap, ok := sm.transitions[sourceState]
	if !ok {
		var zeroState S
		return zeroState, errors.Errorf("state not found: %v", sourceState)
	}

	// Check if there are transitions for the given event
	transitions, ok := eventMap[event]
	if !ok || len(transitions) == 0 {
		var zeroState S
		return zeroState, errors.Errorf("no transition found for event %v from state %v", event, sourceState)
	}

	// Find the first transition with satisfied conditions
	for _, transition := range transitions {
		// Check all conditions
		conditionsMet := true
		for _, condition := range transition.Conditions {
			if !condition.IsSatisfied(ctx) {
				conditionsMet = false
				break
			}
		}

		if conditionsMet {
			// Execute all actions
			for _, action := range transition.Actions {
				if err := action.Execute(sourceState, transition.Target, event, ctx); err != nil {
					var zeroState S
					return zeroState, errors.Wrap(err, "action execution failed")
				}
			}

			// Return the target state
			return transition.Target, nil
		}
	}

	// No transition with satisfied conditions found
	var zeroState S
	return zeroState, errors.Errorf("no transition conditions met for event %v from state %v", event, sourceState)
}

// RegisterTransition registers a transition in the state machine
func (sm *StateMachineImpl[S, E, C]) RegisterTransition(transition *Transition[S, E, C]) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// Initialize maps if they don't exist
	if _, ok := sm.transitions[transition.Source]; !ok {
		sm.transitions[transition.Source] = make(map[E][]*Transition[S, E, C])
	}

	// Add the transition
	sm.transitions[transition.Source][transition.Event] = append(
		sm.transitions[transition.Source][transition.Event],
		transition,
	)

	return nil
}

// ShowStateMachine returns a string representation of the state machine
func (sm *StateMachineImpl[S, E, C]) ShowStateMachine() string {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	result := fmt.Sprintf("StateMachine: %s\n", sm.id)
	for source, eventMap := range sm.transitions {
		result += fmt.Sprintf("  State: %v\n", source)
		for event, transitions := range eventMap {
			for _, transition := range transitions {
				result += fmt.Sprintf("    Event: %v -> State: %v\n", event, transition.Target)
			}
		}
	}
	return result
}

// GeneratePlantUML returns a PlantUML diagram of the state machine
func (sm *StateMachineImpl[S, E, C]) GeneratePlantUML() string {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	result := "@startuml\n"
	result += fmt.Sprintf("title State Machine: %s\n", sm.id)
	result += "\n"

	// Define states
	for source := range sm.transitions {
		result += fmt.Sprintf("state \"%v\" as %v\n", source, source)
	}
	result += "\n"

	// Define transitions
	for source, eventMap := range sm.transitions {
		for event, transitions := range eventMap {
			for _, transition := range transitions {
				result += fmt.Sprintf("%v --> %v : %v\n", source, transition.Target, event)
			}
		}
	}

	result += "@enduml\n"
	return result
}
