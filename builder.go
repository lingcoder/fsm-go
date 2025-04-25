package fsm

// StateMachineBuilder builds state machines with a fluent API
type StateMachineBuilder[S comparable, E comparable, P any] struct {
	stateMachine *StateMachineImpl[S, E, P]
}

// NewStateMachineBuilder creates a new builder
// Returns:
//
//	A new state machine builder instance
func NewStateMachineBuilder[S comparable, E comparable, P any]() *StateMachineBuilder[S, E, P] {
	return &StateMachineBuilder[S, E, P]{
		stateMachine: NewStateMachine[S, E, P](""),
	}
}

// ExternalTransition starts defining an external transition
// Returns:
//
//	A transition builder for configuring the external transition
func (b *StateMachineBuilder[S, E, P]) ExternalTransition() *TransitionBuilder[S, E, P] {
	return &TransitionBuilder[S, E, P]{
		stateMachine:   b.stateMachine,
		transitionType: External,
	}
}

// InternalTransition starts defining an internal transition
// Returns:
//
//	A transition builder for configuring the internal transition
func (b *StateMachineBuilder[S, E, P]) InternalTransition() *TransitionBuilder[S, E, P] {
	return &TransitionBuilder[S, E, P]{
		stateMachine:   b.stateMachine,
		transitionType: Internal,
	}
}

// ExternalTransitions starts defining multiple external transitions from different source states to the same target state
// Returns:
//
//	A multiple transition builder for configuring the transitions
func (b *StateMachineBuilder[S, E, P]) ExternalTransitions() *MultipleTransitionBuilder[S, E, P] {
	return &MultipleTransitionBuilder[S, E, P]{
		stateMachine:   b.stateMachine,
		transitionType: External,
	}
}

// ExternalParallelTransition starts defining an external parallel transition
// Returns:
//
//	A parallel transition builder for configuring the parallel transition
func (b *StateMachineBuilder[S, E, P]) ExternalParallelTransition() *ParallelTransitionBuilder[S, E, P] {
	return &ParallelTransitionBuilder[S, E, P]{
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
func (b *StateMachineBuilder[S, E, P]) Build(machineId string) (StateMachine[S, E, P], error) {
	b.stateMachine.id = machineId
	b.stateMachine.SetReady(true)

	// Register the state machine in a factory
	err := RegisterStateMachine[S, E, P](machineId, b.stateMachine)
	if err != nil {
		return nil, err
	}
	return b.stateMachine, nil
}

// TransitionBuilder builds individual transitions
type TransitionBuilder[S comparable, E comparable, P any] struct {
	stateMachine   *StateMachineImpl[S, E, P]
	transitionType TransitionType
	sourceId       S
	targetId       S
	event          E
	condition      Condition[P]
	action         Action[S, E, P]
}

// From specifies the source state
// Parameters:
//
//	state: Source state
//
// Returns:
//
//	The transition builder for method chaining
func (b *TransitionBuilder[S, E, P]) From(state S) *TransitionBuilder[S, E, P] {
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
func (b *TransitionBuilder[S, E, P]) To(state S) *TransitionBuilder[S, E, P] {
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
func (b *TransitionBuilder[S, E, P]) On(event E) *TransitionBuilder[S, E, P] {
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
func (b *TransitionBuilder[S, E, P]) When(condition Condition[P]) *TransitionBuilder[S, E, P] {
	b.condition = condition
	return b
}

// WhenFunc specifies a function as the condition for the transition
// Parameters:
//
//	conditionFunc: The function that must return true for the transition to occur
//
// Returns:
//
//	The transition builder for method chaining
func (b *TransitionBuilder[S, E, P]) WhenFunc(conditionFunc func(payload P) bool) *TransitionBuilder[S, E, P] {
	b.condition = ConditionFunc[P](conditionFunc)
	return b
}

// Perform specifies the action to execute during the transition
// Parameters:
//
//	action: The action to execute when the transition occurs
func (b *TransitionBuilder[S, E, P]) Perform(action Action[S, E, P]) {
	b.action = action

	// Get or create source and target states
	sourceState := b.stateMachine.GetState(b.sourceId)
	targetState := b.stateMachine.GetState(b.targetId)

	// Add the transition to the source state
	transition := sourceState.AddTransition(b.event, targetState, b.transitionType)

	// Set condition and action
	transition.Condition = b.condition
	transition.Action = b.action
}

// PerformFunc specifies a function as the action to execute during the transition
// Parameters:
//
//	actionFunc: The function to execute when the transition occurs
func (b *TransitionBuilder[S, E, P]) PerformFunc(actionFunc func(from, to S, event E, payload P) error) {
	b.action = ActionFunc[S, E, P](actionFunc)

	// Get or create source and target states
	sourceState := b.stateMachine.GetState(b.sourceId)
	targetState := b.stateMachine.GetState(b.targetId)

	// Add the transition to the source state
	transition := sourceState.AddTransition(b.event, targetState, b.transitionType)

	// Set condition and action
	transition.Condition = b.condition
	transition.Action = b.action
}

// MultipleTransitionBuilder builds transitions from multiple source states to a single target state
type MultipleTransitionBuilder[S comparable, E comparable, P any] struct {
	stateMachine   *StateMachineImpl[S, E, P]
	transitionType TransitionType
	sourceIds      []S
	targetId       S
	event          E
	condition      Condition[P]
	action         Action[S, E, P]
}

// FromAmong specifies multiple source states
// Parameters:
//
//	states: Multiple source states
//
// Returns:
//
//	The multiple transition builder for method chaining
func (b *MultipleTransitionBuilder[S, E, P]) FromAmong(states ...S) *MultipleTransitionBuilder[S, E, P] {
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
func (b *MultipleTransitionBuilder[S, E, P]) To(state S) *MultipleTransitionBuilder[S, E, P] {
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
func (b *MultipleTransitionBuilder[S, E, P]) On(event E) *MultipleTransitionBuilder[S, E, P] {
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
func (b *MultipleTransitionBuilder[S, E, P]) When(condition Condition[P]) *MultipleTransitionBuilder[S, E, P] {
	b.condition = condition
	return b
}

// WhenFunc specifies a function as the condition for all transitions
// Parameters:
//
//	conditionFunc: The function that must return true for the transitions to occur
//
// Returns:
//
//	The multiple transition builder for method chaining
func (b *MultipleTransitionBuilder[S, E, P]) WhenFunc(conditionFunc func(payload P) bool) *MultipleTransitionBuilder[S, E, P] {
	b.condition = ConditionFunc[P](conditionFunc)
	return b
}

// Perform specifies the action to execute during all transitions
// Parameters:
//
//	action: The action to execute when the transitions occur
func (b *MultipleTransitionBuilder[S, E, P]) Perform(action Action[S, E, P]) {
	b.action = action

	// Create transitions for each source state
	for _, sourceId := range b.sourceIds {
		sourceState := b.stateMachine.GetState(sourceId)
		targetState := b.stateMachine.GetState(b.targetId)

		// Add the transition to the source state
		transition := sourceState.AddTransition(b.event, targetState, b.transitionType)

		// Set condition and action
		transition.Condition = b.condition
		transition.Action = b.action
	}
}

// PerformFunc specifies a function as the action to execute during all transitions
// Parameters:
//
//	actionFunc: The function to execute when the transitions occur
func (b *MultipleTransitionBuilder[S, E, P]) PerformFunc(actionFunc func(from, to S, event E, payload P) error) {
	b.action = ActionFunc[S, E, P](actionFunc)

	// Create transitions for each source state
	for _, sourceId := range b.sourceIds {
		sourceState := b.stateMachine.GetState(sourceId)
		targetState := b.stateMachine.GetState(b.targetId)

		// Add the transition to the source state
		transition := sourceState.AddTransition(b.event, targetState, b.transitionType)

		// Set condition and action
		transition.Condition = b.condition
		transition.Action = b.action
	}
}

// ParallelTransitionBuilder builds transitions to multiple target states
type ParallelTransitionBuilder[S comparable, E comparable, P any] struct {
	stateMachine   *StateMachineImpl[S, E, P]
	transitionType TransitionType
	sourceId       S
	targetIds      []S
	event          E
	condition      Condition[P]
	action         Action[S, E, P]
}

// From specifies the source state
// Parameters:
//
//	state: Source state
//
// Returns:
//
//	The parallel transition builder for method chaining
func (b *ParallelTransitionBuilder[S, E, P]) From(state S) *ParallelTransitionBuilder[S, E, P] {
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
func (b *ParallelTransitionBuilder[S, E, P]) ToAmong(states ...S) *ParallelTransitionBuilder[S, E, P] {
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
func (b *ParallelTransitionBuilder[S, E, P]) On(event E) *ParallelTransitionBuilder[S, E, P] {
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
func (b *ParallelTransitionBuilder[S, E, P]) When(condition Condition[P]) *ParallelTransitionBuilder[S, E, P] {
	b.condition = condition
	return b
}

// WhenFunc specifies a function as the condition for all transitions
// Parameters:
//
//	conditionFunc: The function that must return true for the transitions to occur
//
// Returns:
//
//	The parallel transition builder for method chaining
func (b *ParallelTransitionBuilder[S, E, P]) WhenFunc(conditionFunc func(payload P) bool) *ParallelTransitionBuilder[S, E, P] {
	b.condition = ConditionFunc[P](conditionFunc)
	return b
}

// Perform specifies the action to execute during all transitions
// Parameters:
//
//	action: The action to execute when the transitions occur
func (b *ParallelTransitionBuilder[S, E, P]) Perform(action Action[S, E, P]) {
	b.action = action

	// Get or create source state
	sourceState := b.stateMachine.GetState(b.sourceId)

	// Create transitions to all target states
	for _, targetId := range b.targetIds {
		targetState := b.stateMachine.GetState(targetId)
		transition := sourceState.AddTransition(b.event, targetState, b.transitionType)
		transition.Condition = b.condition
		transition.Action = b.action
	}
}

// PerformFunc specifies a function as the action to execute during all transitions
// Parameters:
//
//	actionFunc: The function to execute when the transitions occur
func (b *ParallelTransitionBuilder[S, E, P]) PerformFunc(actionFunc func(from, to S, event E, payload P) error) {
	b.action = ActionFunc[S, E, P](actionFunc)

	// Get or create source state
	sourceState := b.stateMachine.GetState(b.sourceId)

	// Create transitions to all target states
	for _, targetId := range b.targetIds {
		targetState := b.stateMachine.GetState(targetId)
		transition := sourceState.AddTransition(b.event, targetState, b.transitionType)
		transition.Condition = b.condition
		transition.Action = b.action
	}
}
