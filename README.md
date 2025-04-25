# FSM-Go: A Lightweight Finite State Machine for Go

FSM-Go is a lightweight, high-performance, stateless finite state machine implementation in Go, inspired by Alibaba's COLA state machine component.

[中文文档](README-zh.md)

## Features

- Lightweight and stateless design for high performance
- Type-safe implementation using Go generics
- Fluent API for defining state machines
- Support for external, internal, and parallel transitions
- Conditional transitions with custom logic
- Actions that execute during transitions
- Thread-safe for concurrent use
- Visualization support for state machine diagrams

## Installation

```bash
go get github.com/lingcoder/fsm-go
```

## Usage

```go
package main

import (
	"fmt"
	"log"

	"github.com/lingcoder/fsm-go"
)

// Define states
type OrderState string

const (
	OrderCreated   OrderState = "CREATED"
	OrderPaid      OrderState = "PAID"
	OrderShipped   OrderState = "SHIPPED"
	OrderDelivered OrderState = "DELIVERED"
	OrderCancelled OrderState = "CANCELLED"
)

// Define events
type OrderEvent string

const (
	EventPay     OrderEvent = "PAY"
	EventShip    OrderEvent = "SHIP"
	EventDeliver OrderEvent = "DELIVER"
	EventCancel  OrderEvent = "CANCEL"
)

// Define context
type OrderContext struct {
	OrderID   string
	Amount    float64
}

// Define action
type OrderAction struct{}

func (a *OrderAction) Execute(from OrderState, to OrderState, event OrderEvent, ctx OrderContext) error {
	fmt.Printf("Order %s transitioning from %s to %s on event %s\n", 
		ctx.OrderID, from, to, event)
	return nil
}

// Define condition
type OrderCondition struct{}

func (c *OrderCondition) IsSatisfied(ctx OrderContext) bool {
	return true
}

func main() {
	// Create a builder
	builder := fsm.NewStateMachineBuilder[OrderState, OrderEvent, OrderContext]()
	
	// Define the state machine
	builder.ExternalTransition().
		From(OrderCreated).
		To(OrderPaid).
		On(EventPay).
		When(&OrderCondition{}).
		Perform(&OrderAction{})
	
	builder.ExternalTransition().
		From(OrderPaid).
		To(OrderShipped).
		On(EventShip).
		When(&OrderCondition{}).
		Perform(&OrderAction{})
	
	// Build the state machine
	stateMachine, err := builder.Build("OrderStateMachine")
	if err != nil {
		log.Fatalf("Failed to build state machine: %v", err)
	}
	
	// Create context
	ctx := OrderContext{
		OrderID: "ORD-20250425-001",
		Amount:  100.0,
	}
	
	// Transition from CREATED to PAID
	newState, err := stateMachine.FireEvent(OrderCreated, EventPay, ctx)
	if err != nil {
		log.Fatalf("Transition failed: %v", err)
	}
	
	fmt.Printf("New state: %v\n", newState)
}
```

## Core Concepts

- **State**: Represents a specific state in your business process
- **Event**: Triggers state transitions
- **Transition**: Defines how states change in response to events
  - **External Transition**: Transition between different states
  - **Internal Transition**: Actions within the same state
- **Condition**: Logic that determines if a transition should occur
- **Action**: Logic executed when a transition occurs
- **StateMachine**: The core component that manages states and transitions

## Examples

Check the `examples` directory for more detailed examples:

- `examples/order`: Order processing workflow
- `examples/workflow`: Approval workflow
- `examples/game`: Game state management

## Performance

FSM-Go is designed for high performance:

- Stateless design minimizes memory usage
- Efficient transition lookup
- Thread-safe for concurrent use
- Benchmarks included in the test suite

## Visualization

FSM-Go provides a unified way to visualize your state machine with different formats:

```go
// Default format (PlantUML)
plantUML := stateMachine.GenerateDiagram()

// Generate specific format
table := stateMachine.GenerateDiagram(fsm.MarkdownTable)     // Markdown table format
flow := stateMachine.GenerateDiagram(fsm.MarkdownFlow)       // Markdown flow chart format

// Generate multiple formats at once
combined := stateMachine.GenerateDiagram(fsm.PlantUML, fsm.MarkdownTable, fsm.MarkdownFlow)

// For backward compatibility, these methods are still available but deprecated
plantUML = stateMachine.GeneratePlantUML()
table = stateMachine.GenerateMarkdown()
flow = stateMachine.GenerateMarkdownFlowchart()
```

## License

MIT
