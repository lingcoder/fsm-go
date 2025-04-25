package main

import (
	"fmt"
	"github.com/lingcoder/fsm-go"
	"log"
)

// Define states
type OrderState string

const (
	OrderCreated   OrderState = "CREATED"
	OrderPaid      OrderState = "PAID"
	OrderShipped   OrderState = "SHIPPED"
	OrderDelivered OrderState = "DELIVERED"
	OrderCancelled OrderState = "CANCELLED"
	OrderRefunded  OrderState = "REFUNDED"
)

// Define events
type OrderEvent string

const (
	EventPay     OrderEvent = "PAY"
	EventShip    OrderEvent = "SHIP"
	EventDeliver OrderEvent = "DELIVER"
	EventCancel  OrderEvent = "CANCEL"
	EventRefund  OrderEvent = "REFUND"
)

// Define context
type OrderContext struct {
	OrderID   string
	UserID    string
	Amount    float64
	Timestamp int64
}

// Define action
type OrderAction struct{}

func (a *OrderAction) Execute(from OrderState, to OrderState, event OrderEvent, ctx OrderContext) error {
	fmt.Printf("Order %s transitioning from %s to %s on event %s\n",
		ctx.OrderID, from, to, event)

	// Perform business logic based on the transition
	// For example, update database, send notifications, etc.
	return nil
}

// Define condition
type OrderCondition struct{}

func (c *OrderCondition) IsSatisfied(ctx OrderContext) bool {
	// Add your condition logic here
	// For example, check if order amount is valid, user has permission, etc.
	return true
}

// Define a more specific condition for cancellation
type CancellationCondition struct{}

func (c *CancellationCondition) IsSatisfied(ctx OrderContext) bool {
	// Only allow cancellation for orders with amount less than 1000
	return ctx.Amount < 1000
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
		Perform(&OrderAction{}).
		Register()

	builder.ExternalTransition().
		From(OrderPaid).
		To(OrderShipped).
		On(EventShip).
		When(&OrderCondition{}).
		Perform(&OrderAction{}).
		Register()

	builder.ExternalTransition().
		From(OrderShipped).
		To(OrderDelivered).
		On(EventDeliver).
		When(&OrderCondition{}).
		Perform(&OrderAction{}).
		Register()

	// Multiple source states can transition to cancelled
	builder.ExternalTransitions().
		FromAmong(OrderCreated, OrderPaid).
		To(OrderCancelled).
		On(EventCancel).
		When(&CancellationCondition{}).
		Perform(&OrderAction{}).
		Register()

	// Refund can only happen from cancelled state
	builder.ExternalTransition().
		From(OrderCancelled).
		To(OrderRefunded).
		On(EventRefund).
		When(&OrderCondition{}).
		Perform(&OrderAction{}).
		Register()

	// Build the state machine
	stateMachine, _ := builder.Build("OrderStateMachine")

	// Print the state machine structure
	fmt.Println(stateMachine.ShowStateMachine())

	// Use the state machine
	ctx := OrderContext{
		OrderID:   "ORD-001",
		UserID:    "USR-001",
		Amount:    100.0,
		Timestamp: 1619712000,
	}

	// Transition from CREATED to PAID
	newState, err := stateMachine.FireEvent(OrderCreated, EventPay, ctx)
	if err != nil {
		log.Fatalf("Failed to transition: %v", err)
	}
	fmt.Printf("New state: %v\n", newState)

	// Transition from PAID to SHIPPED
	newState, err = stateMachine.FireEvent(OrderPaid, EventShip, ctx)
	if err != nil {
		log.Fatalf("Failed to transition: %v", err)
	}
	fmt.Printf("New state: %v\n", newState)

	// Try to cancel a shipped order (should fail as we only defined cancellation for CREATED and PAID)
	newState, err = stateMachine.FireEvent(OrderShipped, EventCancel, ctx)
	if err != nil {
		fmt.Printf("Expected error: %v\n", err)
	}

	// Generate PlantUML diagram
	fmt.Println("\nPlantUML Diagram:")
	fmt.Println(stateMachine.GeneratePlantUML())
}
