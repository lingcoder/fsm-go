package fsm

import (
	"sync"
)

// stateMachineRegistry manages state machine instances
type stateMachineRegistry struct {
	stateMachines map[string]interface{}
	mutex         sync.RWMutex
}

// Global registry instance
var registry = &stateMachineRegistry{
	stateMachines: make(map[string]interface{}),
}

// RegisterStateMachine adds a state machine to the registry
// Parameters:
//
//	machineId: Unique identifier for the state machine
//	stateMachine: State machine instance to register
//
// Returns:
//
//	Error if a state machine with the same ID already exists
func RegisterStateMachine[S comparable, E comparable, C any](machineId string, stateMachine StateMachine[S, E, C]) error {
	registry.mutex.Lock()
	defer registry.mutex.Unlock()

	if _, exists := registry.stateMachines[machineId]; exists {
		return StateMachineAlreadyExist(machineId)
	}

	registry.stateMachines[machineId] = stateMachine
	return nil
}

// GetStateMachine retrieves a state machine by ID
// Parameters:
//
//	machineId: Unique identifier for the state machine to retrieve
//
// Returns:
//
//	The state machine instance and error if not found or type mismatch
func GetStateMachine[S comparable, E comparable, C any](machineId string) (StateMachine[S, E, C], error) {
	registry.mutex.RLock()
	defer registry.mutex.RUnlock()

	if sm, exists := registry.stateMachines[machineId]; exists {
		if typedSM, ok := sm.(StateMachine[S, E, C]); ok {
			return typedSM, nil
		}
		return nil, StateMachineNotFound(machineId)
	}

	return nil, StateMachineNotFound(machineId)
}

// ListStateMachines returns a list of all registered state machine IDs
// Returns:
//
//	Slice of state machine IDs
func ListStateMachines() []string {
	registry.mutex.RLock()
	defer registry.mutex.RUnlock()

	result := make([]string, 0, len(registry.stateMachines))
	for id := range registry.stateMachines {
		result = append(result, id)
	}

	return result
}

// RemoveStateMachine removes a state machine from the registry
// Parameters:
//
//	machineId: Unique identifier for the state machine to remove
//
// Returns:
//
//	True if the state machine was found and removed, false otherwise
func RemoveStateMachine(machineId string) bool {
	registry.mutex.Lock()
	defer registry.mutex.Unlock()

	if _, exists := registry.stateMachines[machineId]; exists {
		delete(registry.stateMachines, machineId)
		return true
	}

	return false
}
