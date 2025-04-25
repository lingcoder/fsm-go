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

// State represents a state in the state machine
type State[S comparable, E comparable, C any] struct {
	id               S
	eventTransitions map[E][]*Transition[S, E, C]
}

// NewState creates a new state
func NewState[S comparable, E comparable, C any](id S) *State[S, E, C] {
	return &State[S, E, C]{
		id:               id,
		eventTransitions: make(map[E][]*Transition[S, E, C]),
	}
}

// AddTransition adds a transition to this state
func (s *State[S, E, C]) AddTransition(event E, target *State[S, E, C], transType TransitionType) *Transition[S, E, C] {
	transition := &Transition[S, E, C]{
		Source:    s,
		Target:    target,
		Event:     event,
		TransType: transType,
	}

	if _, ok := s.eventTransitions[event]; !ok {
		s.eventTransitions[event] = make([]*Transition[S, E, C], 0)
	}
	s.eventTransitions[event] = append(s.eventTransitions[event], transition)
	return transition
}

// GetEventTransitions returns all transitions for a given event
func (s *State[S, E, C]) GetEventTransitions(event E) []*Transition[S, E, C] {
	return s.eventTransitions[event]
}

// GetID returns the state ID
func (s *State[S, E, C]) GetID() S {
	return s.id
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

// Transition represents a state transition
type Transition[S comparable, E comparable, C any] struct {
	Source    *State[S, E, C]
	Target    *State[S, E, C]
	Event     E
	Condition Condition[C]
	Action    Action[S, E, C]
	TransType TransitionType
}

// Transit executes the transition
func (t *Transition[S, E, C]) Transit(ctx C, checkCondition bool) (*State[S, E, C], error) {
	// Verify internal transition
	if t.TransType == Internal && t.Source != t.Target {
		return nil, errors.New("internal transition source and target states must be the same")
	}

	// Check condition if required
	if checkCondition && t.Condition != nil && !t.Condition.IsSatisfied(ctx) {
		return t.Source, nil // Stay at source state if condition is not satisfied
	}

	// Execute action
	if t.Action != nil {
		if err := t.Action.Execute(t.Source.GetID(), t.Target.GetID(), t.Event, ctx); err != nil {
			return nil, errors.Wrap(err, "action execution failed")
		}
	}

	return t.Target, nil
}

// StateMachineImpl is the implementation of StateMachine
type StateMachineImpl[S comparable, E comparable, C any] struct {
	id       string
	stateMap map[S]*State[S, E, C]
	ready    bool
	mutex    sync.RWMutex
}

// NewStateMachine creates a new state machine
func NewStateMachine[S comparable, E comparable, C any](id string) *StateMachineImpl[S, E, C] {
	return &StateMachineImpl[S, E, C]{
		id:       id,
		stateMap: make(map[S]*State[S, E, C]),
		ready:    false,
	}
}

// FireEvent triggers a state transition based on the current state and event
func (sm *StateMachineImpl[S, E, C]) FireEvent(sourceStateId S, event E, ctx C) (S, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	if !sm.ready {
		var zeroState S
		return zeroState, errors.New("state machine is not built yet")
	}

	// Get source state
	sourceState, ok := sm.stateMap[sourceStateId]
	if !ok {
		var zeroState S
		return zeroState, errors.Errorf("state not found: %v", sourceStateId)
	}

	// Get transitions for the event
	transitions := sourceState.GetEventTransitions(event)
	if transitions == nil || len(transitions) == 0 {
		var zeroState S
		return zeroState, errors.Errorf("no transition found for event %v from state %v", event, sourceStateId)
	}

	// Find the first transition with satisfied condition
	for _, transition := range transitions {
		targetState, err := transition.Transit(ctx, true)
		if err != nil {
			var zeroState S
			return zeroState, err
		}

		if targetState != sourceState { // Transition occurred
			return targetState.GetID(), nil
		}
	}

	// No transition with satisfied conditions found
	var zeroState S
	return zeroState, errors.Errorf("no transition conditions met for event %v from state %v", event, sourceStateId)
}

// GetState returns a state by ID, creating it if it doesn't exist
func (sm *StateMachineImpl[S, E, C]) GetState(stateId S) *State[S, E, C] {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if state, ok := sm.stateMap[stateId]; ok {
		return state
	}

	state := NewState[S, E, C](stateId)
	sm.stateMap[stateId] = state
	return state
}

// SetReady marks the state machine as ready
func (sm *StateMachineImpl[S, E, C]) SetReady(ready bool) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	sm.ready = ready
}

// ShowStateMachine returns a string representation of the state machine
func (sm *StateMachineImpl[S, E, C]) ShowStateMachine() string {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	result := fmt.Sprintf("StateMachine(id=%s):\n", sm.id)

	for _, state := range sm.stateMap {
		for event, transitions := range state.eventTransitions {
			for _, transition := range transitions {
				transType := "EXTERNAL"
				if transition.TransType == Internal {
					transType = "INTERNAL"
				}
				result += fmt.Sprintf("  %v --%v(%s)--> %v\n",
					transition.Source.GetID(), event, transType, transition.Target.GetID())
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
	result += fmt.Sprintf("title StateMachine: %s\n", sm.id)

	// Define states
	for stateId := range sm.stateMap {
		result += fmt.Sprintf("state \"%v\" as %v\n", stateId, stateId)
	}

	// Define transitions
	for _, state := range sm.stateMap {
		for event, transitions := range state.eventTransitions {
			for _, transition := range transitions {
				result += fmt.Sprintf("%v --> %v : %v\n",
					transition.Source.GetID(), transition.Target.GetID(), event)
			}
		}
	}

	result += "@enduml\n"
	return result
}
