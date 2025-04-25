package fsm

import (
	"fmt"
	"github.com/pkg/errors"
	"strings"
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

	// FireParallelEvent triggers parallel state transitions based on the current state and event
	// Returns a slice of new states and any error that occurred
	FireParallelEvent(sourceState S, event E, ctx C) ([]S, error)

	// Verify checks if there is a valid transition for the given state and event
	// Returns true if a transition exists, false otherwise
	Verify(sourceState S, event E) bool

	// ShowStateMachine returns a string representation of the state machine
	ShowStateMachine() string

	// GenerateDiagram returns a diagram of the state machine in the specified formats
	// If formats is nil or empty, defaults to PlantUML
	// If multiple formats are provided, returns all requested formats concatenated
	GenerateDiagram(formats ...DiagramFormat) string
}

// DiagramFormat defines the supported diagram formats
type DiagramFormat int

const (
	// PlantUML format for UML diagrams
	PlantUML DiagramFormat = iota
	// MarkdownTable format for tabular representation
	MarkdownTable
	// MarkdownFlow format for flowcharts
	MarkdownFlow
)

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

// AddParallelTransitions adds multiple transitions for the same event to different target states
func (s *State[S, E, C]) AddParallelTransitions(event E, targets []*State[S, E, C], transType TransitionType) []*Transition[S, E, C] {
	transitions := make([]*Transition[S, E, C], 0, len(targets))

	for _, target := range targets {
		transition := s.AddTransition(event, target, transType)
		transitions = append(transitions, transition)
	}

	return transitions
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

// ConditionFunc is a function type that implements Condition interface
type ConditionFunc[C any] func(ctx C) bool

// IsSatisfied implements Condition interface
func (f ConditionFunc[C]) IsSatisfied(ctx C) bool {
	return f(ctx)
}

// Action is an interface for transition actions
type Action[S comparable, E comparable, C any] interface {
	// Execute runs the action during a state transition
	Execute(from, to S, event E, ctx C) error
}

// ActionFunc is a function type that implements Action interface
type ActionFunc[S comparable, E comparable, C any] func(from, to S, event E, ctx C) error

// Execute implements Action interface
func (f ActionFunc[S, E, C]) Execute(from, to S, event E, ctx C) error {
	return f(from, to, event, ctx)
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

// StateMachineImpl implements the StateMachine interface
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
		if transition.Condition == nil || transition.Condition.IsSatisfied(ctx) {
			targetState, err := transition.Transit(ctx, true)
			if err != nil {
				var zeroState S
				return zeroState, err
			}
			return targetState.GetID(), nil
		}
	}

	var zeroState S
	return zeroState, errors.Errorf("no transition conditions met for event %v from state %v", event, sourceStateId)
}

// FireParallelEvent triggers parallel state transitions based on the current state and event
func (sm *StateMachineImpl[S, E, C]) FireParallelEvent(sourceStateId S, event E, ctx C) ([]S, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	if !sm.ready {
		return nil, errors.New("state machine is not built yet")
	}

	// Get source state
	sourceState, ok := sm.stateMap[sourceStateId]
	if !ok {
		return nil, errors.Errorf("state not found: %v", sourceStateId)
	}

	// Get transitions for the event
	transitions := sourceState.GetEventTransitions(event)
	if transitions == nil || len(transitions) == 0 {
		return nil, errors.Errorf("no transition found for event %v from state %v", event, sourceStateId)
	}

	// Execute all transitions with satisfied conditions
	var results []S
	var validTransitions []*Transition[S, E, C]

	// First, find all valid transitions
	for _, transition := range transitions {
		if transition.Condition == nil || transition.Condition.IsSatisfied(ctx) {
			validTransitions = append(validTransitions, transition)
		}
	}

	if len(validTransitions) == 0 {
		return nil, errors.Errorf("no transition conditions met for event %v from state %v", event, sourceStateId)
	}

	// Then execute all valid transitions
	for _, transition := range validTransitions {
		targetState, err := transition.Transit(ctx, false) // Don't check condition again
		if err != nil {
			return nil, err
		}
		results = append(results, targetState.GetID())
	}

	return results, nil
}

// Verify checks if there is a valid transition for the given state and event
func (sm *StateMachineImpl[S, E, C]) Verify(sourceStateId S, event E) bool {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	if !sm.ready {
		return false
	}

	// Get source state
	sourceState, ok := sm.stateMap[sourceStateId]
	if !ok {
		return false
	}

	// Get transitions for the event
	transitions := sourceState.GetEventTransitions(event)

	// Return true if there is at least one transition for this event
	return transitions != nil && len(transitions) > 0
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

// GenerateDiagram returns a diagram of the state machine in the specified formats
// If formats is nil or empty, defaults to PlantUML
// If multiple formats are provided, returns all requested formats concatenated
func (sm *StateMachineImpl[S, E, C]) GenerateDiagram(formats ...DiagramFormat) string {
	if len(formats) == 0 {
		return sm.generatePlantUML()
	}

	var result strings.Builder
	for i, format := range formats {
		if i > 0 {
			result.WriteString("\n\n")
		}

		switch format {
		case MarkdownTable:
			result.WriteString(sm.generateMarkdownTable())
		case MarkdownFlow:
			result.WriteString(sm.generateMarkdownFlow())
		case PlantUML:
			result.WriteString(sm.generatePlantUML())
		default:
			result.WriteString(sm.generatePlantUML())
		}
	}

	return result.String()
}

// generatePlantUML returns a PlantUML diagram of the state machine
func (sm *StateMachineImpl[S, E, C]) generatePlantUML() string {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	var sb strings.Builder
	sb.WriteString("@startuml\n")
	sb.WriteString(fmt.Sprintf("title StateMachine: %s\n", sm.id))

	// Define states
	for stateId := range sm.stateMap {
		sb.WriteString(fmt.Sprintf("state \"%v\" as %v\n", stateId, stateId))
	}

	// Define transitions
	for _, state := range sm.stateMap {
		for _, transitions := range state.eventTransitions {
			for _, transition := range transitions {
				sb.WriteString(fmt.Sprintf("%v --> %v : %v\n", transition.Source.id, transition.Target.id, transition.Event))
			}
		}
	}

	sb.WriteString("@enduml\n")
	return sb.String()
}

// generateMarkdownTable returns a Markdown table representation of the state machine
func (sm *StateMachineImpl[S, E, C]) generateMarkdownTable() string {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# State Machine: %s\n\n", sm.id))

	// States section
	sb.WriteString("## States\n\n")
	for stateId := range sm.stateMap {
		sb.WriteString(fmt.Sprintf("- `%v`\n", stateId))
	}
	sb.WriteString("\n")

	// Transitions section
	sb.WriteString("## Transitions\n\n")
	sb.WriteString("| Source State | Event | Target State | Type |\n")
	sb.WriteString("|-------------|-------|--------------|------|\n")

	// Sort states for consistent output
	stateIds := make([]S, 0, len(sm.stateMap))
	for stateId := range sm.stateMap {
		stateIds = append(stateIds, stateId)
	}

	// Sort events for each state
	for _, sourceId := range stateIds {
		sourceState := sm.stateMap[sourceId]

		for event, transitions := range sourceState.eventTransitions {
			for _, transition := range transitions {
				transType := "External"
				if transition.TransType == Internal {
					transType = "Internal"
				}

				sb.WriteString(fmt.Sprintf("| `%v` | `%v` | `%v` | %s |\n",
					sourceId, event, transition.Target.id, transType))
			}
		}
	}

	return sb.String()
}

// generateMarkdownFlow returns a Mermaid flowchart diagram in Markdown format
func (sm *StateMachineImpl[S, E, C]) generateMarkdownFlow() string {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	var sb strings.Builder
	sb.WriteString("```mermaid\nflowchart TD\n")

	// Define node IDs - we need to ensure they are valid Mermaid IDs
	nodeIds := make(map[S]string)
	i := 0
	for stateId := range sm.stateMap {
		// Create a valid Mermaid ID (alphanumeric and underscores only)
		nodeIds[stateId] = fmt.Sprintf("state_%d", i)
		sb.WriteString(fmt.Sprintf("    %s[\"%v\"]\n", nodeIds[stateId], stateId))
		i++
	}

	// Define transitions
	for _, state := range sm.stateMap {
		sourceNodeId := nodeIds[state.id]

		for event, transitions := range state.eventTransitions {
			for _, transition := range transitions {
				targetNodeId := nodeIds[transition.Target.id]
				sb.WriteString(fmt.Sprintf("    %s -->|%v| %s\n",
					sourceNodeId, event, targetNodeId))
			}
		}
	}

	sb.WriteString("```\n")
	return sb.String()
}

// GeneratePlantUML returns a PlantUML diagram of the state machine
// Deprecated: Use GenerateDiagram(PlantUML) or GenerateDiagram() instead
func (sm *StateMachineImpl[S, E, C]) GeneratePlantUML() string {
	return sm.GenerateDiagram(PlantUML)
}

// GenerateMarkdown returns a Markdown representation of the state machine
// Deprecated: Use GenerateDiagram(MarkdownTable) instead
func (sm *StateMachineImpl[S, E, C]) GenerateMarkdown() string {
	return sm.GenerateDiagram(MarkdownTable)
}

// GenerateMarkdownFlowchart returns a Mermaid flowchart diagram in Markdown format
// Deprecated: Use GenerateDiagram(MarkdownFlow) instead
func (sm *StateMachineImpl[S, E, C]) GenerateMarkdownFlowchart() string {
	return sm.GenerateDiagram(MarkdownFlow)
}
