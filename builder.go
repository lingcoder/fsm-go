package fsm

// StateMachineBuilder builds state machines with a fluent API
type StateMachineBuilder[S comparable, E comparable, C any] struct {
	stateMachine       *StateMachineImpl[S, E, C]
	pendingTransitions []*Transition[S, E, C]
}

// NewStateMachineBuilder creates a new builder
// Returns:
//
//	A new state machine builder instance
func NewStateMachineBuilder[S comparable, E comparable, C any]() *StateMachineBuilder[S, E, C] {
	return &StateMachineBuilder[S, E, C]{
		stateMachine:       NewStateMachine[S, E, C](""),
		pendingTransitions: make([]*Transition[S, E, C], 0),
	}
}

// ExternalTransition starts defining an external transition
// Returns:
//
//	A transition builder for configuring the external transition
func (b *StateMachineBuilder[S, E, C]) ExternalTransition() *TransitionBuilder[S, E, C] {
	return &TransitionBuilder[S, E, C]{
		stateMachine:   b.stateMachine,
		transitionType: External,
		builder:        b,
	}
}

// InternalTransition starts defining an internal transition
// Returns:
//
//	A transition builder for configuring the internal transition
func (b *StateMachineBuilder[S, E, C]) InternalTransition() *TransitionBuilder[S, E, C] {
	return &TransitionBuilder[S, E, C]{
		stateMachine:   b.stateMachine,
		transitionType: Internal,
		builder:        b,
	}
}

// ExternalTransitions starts defining multiple external transitions
// Returns:
//
//	A multiple transition builder for configuring transitions from multiple source states
func (b *StateMachineBuilder[S, E, C]) ExternalTransitions() *MultipleTransitionBuilder[S, E, C] {
	return &MultipleTransitionBuilder[S, E, C]{
		stateMachine:   b.stateMachine,
		transitionType: External,
		builder:        b,
	}
}

// Build finalizes the state machine with the given ID
// Parameters:
//
//	machineId: Unique identifier for the state machine
//
// Returns:
//
//	The built state machine and possible error
func (b *StateMachineBuilder[S, E, C]) Build(machineId string) (StateMachine[S, E, C], error) {
	b.stateMachine.id = machineId

	// Register all pending transitions
	for _, transition := range b.pendingTransitions {
		err := b.stateMachine.RegisterTransition(transition)
		if err != nil {
			return nil, err
		}
	}

	// Register the state machine in a factory
	err := RegisterStateMachine[S, E, C](machineId, b.stateMachine)
	if err != nil {
		return nil, err
	}
	return b.stateMachine, nil
}

// TransitionBuilder builds individual transitions
type TransitionBuilder[S comparable, E comparable, C any] struct {
	stateMachine   *StateMachineImpl[S, E, C]
	transitionType TransitionType
	source         S
	target         S
	event          E
	conditions     []Condition[C]
	actions        []Action[S, E, C]
	builder        *StateMachineBuilder[S, E, C]
}

// From specifies the source state
// Parameters:
//
//	state: Source state
//
// Returns:
//
//	The transition builder for method chaining
func (b *TransitionBuilder[S, E, C]) From(state S) *TransitionBuilder[S, E, C] {
	b.source = state
	return b
}

// To specifies the target state
// Parameters:
//
//	state: Target state
//
// Returns:
//
//	The transition builder for method chaining
func (b *TransitionBuilder[S, E, C]) To(state S) *TransitionBuilder[S, E, C] {
	b.target = state
	return b
}

// On specifies the triggering event
// Parameters:
//
//	event: The event that triggers this transition
//
// Returns:
//
//	The transition builder for method chaining
func (b *TransitionBuilder[S, E, C]) On(event E) *TransitionBuilder[S, E, C] {
	b.event = event
	return b
}

// When specifies the condition for the transition
// Parameters:
//
//	condition: The condition that must be satisfied for the transition to occur
//
// Returns:
//
//	The transition builder for method chaining
func (b *TransitionBuilder[S, E, C]) When(condition Condition[C]) *TransitionBuilder[S, E, C] {
	b.conditions = append(b.conditions, condition)
	return b
}

// Perform specifies the action to execute during the transition
// Parameters:
//
//	action: The action to execute when the transition occurs
//
// Returns:
//
//	The transition builder for method chaining
func (b *TransitionBuilder[S, E, C]) Perform(action Action[S, E, C]) *TransitionBuilder[S, E, C] {
	b.actions = append(b.actions, action)

	// Create the transition and add to pending transitions
	transition := &Transition[S, E, C]{
		Source:     b.source,
		Target:     b.target,
		Event:      b.event,
		Conditions: b.conditions,
		Actions:    b.actions,
		TransType:  b.transitionType,
	}

	// Add to pending transitions
	b.builder.pendingTransitions = append(b.builder.pendingTransitions, transition)
	return b
}

// MultipleTransitionBuilder builds transitions from multiple source states
type MultipleTransitionBuilder[S comparable, E comparable, C any] struct {
	stateMachine   *StateMachineImpl[S, E, C]
	transitionType TransitionType
	sources        []S
	target         S
	event          E
	conditions     []Condition[C]
	actions        []Action[S, E, C]
	builder        *StateMachineBuilder[S, E, C]
}

// FromAmong specifies multiple source states
// Parameters:
//
//	states: Multiple source states
//
// Returns:
//
//	The multiple transition builder for method chaining
func (b *MultipleTransitionBuilder[S, E, C]) FromAmong(states ...S) *MultipleTransitionBuilder[S, E, C] {
	b.sources = states
	return b
}

// To specifies the target state
// Parameters:
//
//	state: Target state
//
// Returns:
//
//	The multiple transition builder for method chaining
func (b *MultipleTransitionBuilder[S, E, C]) To(state S) *MultipleTransitionBuilder[S, E, C] {
	b.target = state
	return b
}

// On specifies the triggering event
// Parameters:
//
//	event: The event that triggers these transitions
//
// Returns:
//
//	The multiple transition builder for method chaining
func (b *MultipleTransitionBuilder[S, E, C]) On(event E) *MultipleTransitionBuilder[S, E, C] {
	b.event = event
	return b
}

// When specifies the condition for all transitions
// Parameters:
//
//	condition: The condition that must be satisfied for the transitions to occur
//
// Returns:
//
//	The multiple transition builder for method chaining
func (b *MultipleTransitionBuilder[S, E, C]) When(condition Condition[C]) *MultipleTransitionBuilder[S, E, C] {
	b.conditions = append(b.conditions, condition)
	return b
}

// Perform specifies the action to execute during all transitions
// Parameters:
//
//	action: The action to execute when the transitions occur
//
// Returns:
//
//	The multiple transition builder for method chaining
func (b *MultipleTransitionBuilder[S, E, C]) Perform(action Action[S, E, C]) *MultipleTransitionBuilder[S, E, C] {
	b.actions = append(b.actions, action)

	// Create transitions for each source state
	for _, source := range b.sources {
		transition := &Transition[S, E, C]{
			Source:     source,
			Target:     b.target,
			Event:      b.event,
			Conditions: b.conditions,
			Actions:    b.actions,
			TransType:  b.transitionType,
		}

		// Add to pending transitions
		b.builder.pendingTransitions = append(b.builder.pendingTransitions, transition)
	}
	return b
}
