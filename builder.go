package fsm

// StateMachineBuilder builds state machines with a fluent API
type StateMachineBuilder[S comparable, E comparable, C any] struct {
	stateMachine *StateMachineImpl[S, E, C]
}

// NewStateMachineBuilder creates a new builder
// Returns:
//
//	A new state machine builder instance
func NewStateMachineBuilder[S comparable, E comparable, C any]() *StateMachineBuilder[S, E, C] {
	return &StateMachineBuilder[S, E, C]{
		stateMachine: NewStateMachine[S, E, C](""),
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
	}
}

// ExternalParallelTransition starts defining an external parallel transition
// Returns:
//
//	A parallel transition builder for configuring the parallel transition
func (b *StateMachineBuilder[S, E, C]) ExternalParallelTransition() *ParallelTransitionBuilder[S, E, C] {
	return &ParallelTransitionBuilder[S, E, C]{
		stateMachine:   b.stateMachine,
		transitionType: External,
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
	b.stateMachine.SetReady(true)

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
	sourceId       S
	targetId       S
	event          E
	condition      Condition[C]
	action         Action[S, E, C]
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
	b.sourceId = state
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
	b.targetId = state
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
	b.condition = condition
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
	b.action = action

	// Get or create source and target states
	sourceState := b.stateMachine.GetState(b.sourceId)
	targetState := b.stateMachine.GetState(b.targetId)

	// Add the transition to the source state
	transition := sourceState.AddTransition(b.event, targetState, b.transitionType)

	// Set condition and action
	transition.Condition = b.condition
	transition.Action = b.action

	return b
}

// MultipleTransitionBuilder builds transitions from multiple source states
type MultipleTransitionBuilder[S comparable, E comparable, C any] struct {
	stateMachine   *StateMachineImpl[S, E, C]
	transitionType TransitionType
	sourceIds      []S
	targetId       S
	event          E
	condition      Condition[C]
	action         Action[S, E, C]
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
	b.sourceIds = states
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
	b.targetId = state
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
	b.condition = condition
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
	b.action = action

	// Get or create target state
	targetState := b.stateMachine.GetState(b.targetId)

	// For each source state, add a transition
	for _, sourceId := range b.sourceIds {
		sourceState := b.stateMachine.GetState(sourceId)

		// Add the transition to the source state
		transition := sourceState.AddTransition(b.event, targetState, b.transitionType)

		// Set condition and action
		transition.Condition = b.condition
		transition.Action = b.action
	}

	return b
}

// ParallelTransitionBuilder builds transitions to multiple target states
type ParallelTransitionBuilder[S comparable, E comparable, C any] struct {
	stateMachine   *StateMachineImpl[S, E, C]
	transitionType TransitionType
	sourceId       S
	targetIds      []S
	event          E
	condition      Condition[C]
	action         Action[S, E, C]
}

// From specifies the source state
// Parameters:
//
//	state: Source state
//
// Returns:
//
//	The parallel transition builder for method chaining
func (b *ParallelTransitionBuilder[S, E, C]) From(state S) *ParallelTransitionBuilder[S, E, C] {
	b.sourceId = state
	return b
}

// ToAmong specifies multiple target states
// Parameters:
//
//	states: Multiple target states
//
// Returns:
//
//	The parallel transition builder for method chaining
func (b *ParallelTransitionBuilder[S, E, C]) ToAmong(states ...S) *ParallelTransitionBuilder[S, E, C] {
	b.targetIds = states
	return b
}

// On specifies the triggering event
// Parameters:
//
//	event: The event that triggers these transitions
//
// Returns:
//
//	The parallel transition builder for method chaining
func (b *ParallelTransitionBuilder[S, E, C]) On(event E) *ParallelTransitionBuilder[S, E, C] {
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
//	The parallel transition builder for method chaining
func (b *ParallelTransitionBuilder[S, E, C]) When(condition Condition[C]) *ParallelTransitionBuilder[S, E, C] {
	b.condition = condition
	return b
}

// Perform specifies the action to execute during all transitions
// Parameters:
//
//	action: The action to execute when the transitions occur
//
// Returns:
//
//	The parallel transition builder for method chaining
func (b *ParallelTransitionBuilder[S, E, C]) Perform(action Action[S, E, C]) *ParallelTransitionBuilder[S, E, C] {
	b.action = action

	// Get or create source state
	sourceState := b.stateMachine.GetState(b.sourceId)

	// Get or create all target states
	targetStates := make([]*State[S, E, C], 0, len(b.targetIds))
	for _, targetId := range b.targetIds {
		targetState := b.stateMachine.GetState(targetId)
		targetStates = append(targetStates, targetState)
	}

	// Add parallel transitions
	transitions := sourceState.AddParallelTransitions(b.event, targetStates, b.transitionType)

	// Set condition and action for all transitions
	for _, transition := range transitions {
		transition.Condition = b.condition
		transition.Action = b.action
	}

	return b
}
