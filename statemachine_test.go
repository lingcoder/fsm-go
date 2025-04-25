package fsm

import (
	"fmt"
	"sync"
	"testing"
)

// Define test states, events and context
type testState string
type testEvent string
type testContext struct {
	Value string
}

const (
	StateA testState = "A"
	StateB testState = "B"
	StateC testState = "C"
	StateD testState = "D"
)

const (
	Event1 testEvent = "Event1"
	Event2 testEvent = "Event2"
	Event3 testEvent = "Event3"
)

// Simple condition that always returns true
type alwaysTrueCondition struct{}

func (c *alwaysTrueCondition) IsSatisfied(ctx testContext) bool {
	return true
}

// Simple action that does nothing
type noopAction struct{}

func (a *noopAction) Execute(from, to testState, event testEvent, ctx testContext) error {
	return nil
}

// Create a test state machine
func createTestStateMachine(tb testing.TB) StateMachine[testState, testEvent, testContext] {
	builder := NewStateMachineBuilder[testState, testEvent, testContext]()

	// Define state transitions
	builder.ExternalTransition().
		From(StateA).
		To(StateB).
		On(Event1).
		When(&alwaysTrueCondition{}).
		Perform(&noopAction{})

	builder.ExternalTransition().
		From(StateB).
		To(StateC).
		On(Event2).
		When(&alwaysTrueCondition{}).
		Perform(&noopAction{})

	builder.ExternalTransition().
		From(StateC).
		To(StateA).
		On(Event3).
		When(&alwaysTrueCondition{}).
		Perform(&noopAction{})

	// Build the state machine
	sm, err := builder.Build("TestStateMachine")
	if err != nil {
		tb.Fatalf("Failed to build state machine: %v", err)
	}

	return sm
}

// Benchmark: Single-thread state transition
func BenchmarkSingleThreadTransition(b *testing.B) {
	sm := createTestStateMachine(b)
	ctx := testContext{Value: "test"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		currentState := StateA

		newState, err := sm.FireEvent(currentState, Event1, ctx)
		if err != nil {
			b.Fatalf("Failed to fire event: %v", err)
		}
		currentState = newState

		newState, err = sm.FireEvent(currentState, Event2, ctx)
		if err != nil {
			b.Fatalf("Failed to fire event: %v", err)
		}
		currentState = newState

		newState, err = sm.FireEvent(currentState, Event3, ctx)
		if err != nil {
			b.Fatalf("Failed to fire event: %v", err)
		}
	}
}

// Benchmark: Multi-thread state transition
func BenchmarkMultiThreadTransition(b *testing.B) {
	sm := createTestStateMachine(b)
	ctx := testContext{Value: "test"}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			currentState := StateA

			newState, err := sm.FireEvent(currentState, Event1, ctx)
			if err != nil {
				b.Fatalf("Failed to fire event: %v", err)
			}
			currentState = newState

			newState, err = sm.FireEvent(currentState, Event2, ctx)
			if err != nil {
				b.Fatalf("Failed to fire event: %v", err)
			}
			currentState = newState

			newState, err = sm.FireEvent(currentState, Event3, ctx)
			if err != nil {
				b.Fatalf("Failed to fire event: %v", err)
			}
		}
	})
}

// Benchmark: State machine creation
func BenchmarkStateMachineCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		builder := NewStateMachineBuilder[testState, testEvent, testContext]()

		builder.ExternalTransition().
			From(StateA).
			To(StateB).
			On(Event1).
			When(&alwaysTrueCondition{}).
			Perform(&noopAction{})

		machineId := fmt.Sprintf("TestMachine-%d", i)
		_, err := builder.Build(machineId)
		if err != nil {
			b.Fatalf("Failed to build state machine: %v", err)
		}

		// Clean up to avoid memory leaks
		RemoveStateMachine(machineId)
	}
}

// Benchmark: Large state machine with many states and transitions
func BenchmarkLargeStateMachine(b *testing.B) {
	// Create a state machine with 100 states and transitions
	builder := NewStateMachineBuilder[testState, testEvent, testContext]()

	const numStates = 100
	states := make([]testState, numStates)
	for i := 0; i < numStates; i++ {
		states[i] = testState(fmt.Sprintf("State%d", i))
	}

	// Create chain transitions: State0 -> State1 -> ... -> State99 -> State0
	for i := 0; i < numStates; i++ {
		from := states[i]
		to := states[(i+1)%numStates]

		builder.ExternalTransition().
			From(from).
			To(to).
			On(Event1).
			When(&alwaysTrueCondition{}).
			Perform(&noopAction{})
	}

	sm, err := builder.Build("LargeStateMachine")
	if err != nil {
		b.Fatalf("Failed to build large state machine: %v", err)
	}

	ctx := testContext{Value: "test"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		currentState := states[i%numStates]
		newState, err := sm.FireEvent(currentState, Event1, ctx)
		if err != nil {
			b.Fatalf("Failed to fire event: %v", err)
		}
		_ = newState
	}
}

// Benchmark: Concurrent state machine access
func BenchmarkConcurrentStateMachineAccess(b *testing.B) {
	sm := createTestStateMachine(b)
	ctx := testContext{Value: "test"}

	var wg sync.WaitGroup
	numGoroutines := 100

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(numGoroutines)

		for j := 0; j < numGoroutines; j++ {
			go func(j int) {
				defer wg.Done()

				// Each goroutine executes a different state transition
				currentState := StateA
				event := Event1

				if j%3 == 1 {
					currentState = StateB
					event = Event2
				} else if j%3 == 2 {
					currentState = StateC
					event = Event3
				}

				_, err := sm.FireEvent(currentState, event, ctx)
				if err != nil {
					// Ignore errors, as some are expected in concurrent testing
					// For example, when state doesn't match the event
				}
			}(j)
		}

		wg.Wait()
	}
}
