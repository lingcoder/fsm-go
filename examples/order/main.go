package main

import (
	"fmt"
	"log"

	"github.com/lingcoder/fsm-go"
)

// Order states
type OrderState string

const (
	Created   OrderState = "CREATED"
	Paid      OrderState = "PAID"
	Delivered OrderState = "DELIVERED"
	Cancelled OrderState = "CANCELLED"
	Finished  OrderState = "FINISHED"
	Notified  OrderState = "NOTIFIED" // New notification state for parallel transition demo
)

// Order events
type OrderEvent string

const (
	Pay     OrderEvent = "PAY"
	Deliver OrderEvent = "DELIVER"
	Cancel  OrderEvent = "CANCEL"
	Confirm OrderEvent = "CONFIRM"
	Process OrderEvent = "PROCESS" // New process event for parallel transition demo
)

// Order context
type OrderContext struct {
	OrderId string
	Amount  float64
	User    string
}

func main() {
	// Create state machine builder
	builder := fsm.NewStateMachineBuilder[OrderState, OrderEvent, OrderContext]()

	// Using function types for conditions and actions
	// From Created to Paid
	builder.ExternalTransition().
		From(Created).
		To(Paid).
		On(Pay).
		WhenFunc(func(ctx OrderContext) bool {
			return ctx.Amount > 0
		}).
		PerformFunc(func(from, to OrderState, event OrderEvent, ctx OrderContext) error {
			fmt.Printf("Order %s paid amount %.2f\n", ctx.OrderId, ctx.Amount)
			return nil
		})

	// From Paid to Delivered
	builder.ExternalTransition().
		From(Paid).
		To(Delivered).
		On(Deliver).
		PerformFunc(func(from, to OrderState, event OrderEvent, ctx OrderContext) error {
			fmt.Printf("Order %s has been delivered\n", ctx.OrderId)
			return nil
		})

	// From Delivered to Finished
	builder.ExternalTransition().
		From(Delivered).
		To(Finished).
		On(Confirm).
		PerformFunc(func(from, to OrderState, event OrderEvent, ctx OrderContext) error {
			fmt.Printf("Order %s has been completed\n", ctx.OrderId)
			return nil
		})

	// Demonstrate parallel transition: from Paid to both Delivered and Notified
	builder.ExternalParallelTransition().
		From(Paid).
		ToAmong(Delivered, Notified).
		On(Process).
		PerformFunc(func(from, to OrderState, event OrderEvent, ctx OrderContext) error {
			fmt.Printf("Processing order %s to %s state\n", ctx.OrderId, to)
			return nil
		})

	// Demonstrate multiple transitions: from multiple states to Cancelled
	builder.ExternalTransitions().
		FromAmong(Created, Paid, Delivered).
		To(Cancelled).
		On(Cancel).
		PerformFunc(func(from, to OrderState, event OrderEvent, ctx OrderContext) error {
			fmt.Printf("Order %s cancelled from %s state\n", ctx.OrderId, from)
			return nil
		})

	// Build the state machine
	stateMachine, err := builder.Build("OrderStateMachine")
	if err != nil {
		log.Fatalf("Failed to build state machine: %v", err)
	}

	// Create order context
	ctx := OrderContext{
		OrderId: "ORD-20250425-001",
		Amount:  199.99,
		User:    "John Doe",
	}

	// Demonstrate verification feature
	fmt.Println("\n=== Verification Demo ===")
	fmt.Printf("Can trigger Pay event from Created state? %v\n", stateMachine.Verify(Created, Pay))
	fmt.Printf("Can trigger Deliver event from Created state? %v\n", stateMachine.Verify(Created, Deliver))

	// Demonstrate normal transition
	fmt.Println("\n=== Normal Transition Demo ===")
	newState, err := stateMachine.FireEvent(Created, Pay, ctx)
	if err != nil {
		log.Fatalf("Transition failed: %v", err)
	}
	fmt.Printf("New state: %v\n", newState)

	// Demonstrate parallel transition
	fmt.Println("\n=== Parallel Transition Demo ===")
	newStates, err := stateMachine.FireParallelEvent(Paid, Process, ctx)
	if err != nil {
		log.Fatalf("Parallel transition failed: %v", err)
	}
	fmt.Printf("States after parallel transition: %v\n", newStates)

	// Demonstrate multiple transition
	fmt.Println("\n=== Multiple Transition Demo ===")
	ctx.OrderId = "ORD-20250425-002" // New order
	// From Created to Cancelled
	newState, err = stateMachine.FireEvent(Created, Cancel, ctx)
	if err != nil {
		log.Fatalf("Multiple transition failed: %v", err)
	}
	fmt.Printf("State after multiple transition: %v\n", newState)

	// Print state machine diagram
	fmt.Println("\n=== State Machine Diagram ===")
	fmt.Println(stateMachine.ShowStateMachine())

	// Generate PlantUML diagram
	fmt.Println("\n=== PlantUML Diagram ===")
	fmt.Println(stateMachine.GeneratePlantUML())
}
