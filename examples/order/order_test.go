package order

import (
	"testing"

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
	Notified  OrderState = "NOTIFIED" // Notification state for parallel transition
)

// Order events
type OrderEvent string

const (
	Pay     OrderEvent = "PAY"
	Deliver OrderEvent = "DELIVER"
	Cancel  OrderEvent = "CANCEL"
	Confirm OrderEvent = "CONFIRM"
	Process OrderEvent = "PROCESS" // Process event for parallel transition
)

// Order context
type OrderContext struct {
	OrderId string
	Amount  float64
	User    string
}

// TestOrderStateMachine tests the basic functionality of a state machine for order processing
func TestOrderStateMachine(t *testing.T) {
	// Create state machine builder
	builder := fsm.NewStateMachineBuilder[OrderState, OrderEvent, OrderContext]()

	// From Created to Paid
	builder.ExternalTransition().
		From(Created).
		To(Paid).
		On(Pay).
		WhenFunc(func(ctx OrderContext) bool {
			return ctx.Amount > 0
		}).
		PerformFunc(func(from, to OrderState, event OrderEvent, ctx OrderContext) error {
			t.Logf("Order %s paid amount %.2f", ctx.OrderId, ctx.Amount)
			return nil
		})

	// From Paid to Delivered
	builder.ExternalTransition().
		From(Paid).
		To(Delivered).
		On(Deliver).
		PerformFunc(func(from, to OrderState, event OrderEvent, ctx OrderContext) error {
			t.Logf("Order %s delivered", ctx.OrderId)
			return nil
		})

	// From Delivered to Finished
	builder.ExternalTransition().
		From(Delivered).
		To(Finished).
		On(Confirm).
		PerformFunc(func(from, to OrderState, event OrderEvent, ctx OrderContext) error {
			t.Logf("Order %s confirmed and finished", ctx.OrderId)
			return nil
		})

	// Cancel transitions from multiple states
	builder.ExternalTransitions().
		FromAmong(Created, Paid, Delivered).
		To(Cancelled).
		On(Cancel).
		PerformFunc(func(from, to OrderState, event OrderEvent, ctx OrderContext) error {
			t.Logf("Order %s cancelled from %s state", ctx.OrderId, from)
			return nil
		})

	// Build the state machine
	stateMachine, err := builder.Build("OrderStateMachine")
	if err != nil {
		t.Fatalf("Failed to build state machine: %v", err)
	}

	// Test verification
	t.Run("Verification", func(t *testing.T) {
		if !stateMachine.Verify(Created, Pay) {
			t.Error("Should be able to trigger Pay event from Created state")
		}
		if stateMachine.Verify(Created, Deliver) {
			t.Error("Should not be able to trigger Deliver event from Created state")
		}
	})

	// Test normal transition
	t.Run("NormalTransition", func(t *testing.T) {
		ctx := OrderContext{
			OrderId: "ORD-20250425-001",
			Amount:  199.99,
			User:    "user1",
		}

		newState, err := stateMachine.FireEvent(Created, Pay, ctx)
		if err != nil {
			t.Fatalf("Failed to transition: %v", err)
		}
		if newState != Paid {
			t.Errorf("Expected state to be %s, got %s", Paid, newState)
		}

		newState, err = stateMachine.FireEvent(newState, Deliver, ctx)
		if err != nil {
			t.Fatalf("Failed to transition: %v", err)
		}
		if newState != Delivered {
			t.Errorf("Expected state to be %s, got %s", Delivered, newState)
		}
	})

	// Test cancel transition
	t.Run("CancelTransition", func(t *testing.T) {
		ctx := OrderContext{
			OrderId: "ORD-20250425-002",
			Amount:  299.99,
			User:    "user2",
		}

		newState, err := stateMachine.FireEvent(Created, Cancel, ctx)
		if err != nil {
			t.Fatalf("Failed to transition: %v", err)
		}
		if newState != Cancelled {
			t.Errorf("Expected state to be %s, got %s", Cancelled, newState)
		}
	})
}

// TestParallelTransition tests the parallel transition functionality
func TestParallelTransition(t *testing.T) {
	// Create state machine builder
	builder := fsm.NewStateMachineBuilder[OrderState, OrderEvent, OrderContext]()

	// Setup parallel transition from Paid to both Delivered and Notified
	builder.ExternalParallelTransition().
		From(Paid).
		ToAmong(Delivered, Notified).
		On(Process).
		PerformFunc(func(from, to OrderState, event OrderEvent, ctx OrderContext) error {
			t.Logf("Processing order %s to %s state", ctx.OrderId, to)
			return nil
		})

	// Build the state machine
	stateMachine, err := builder.Build("ParallelOrderStateMachine")
	if err != nil {
		t.Fatalf("Failed to build state machine: %v", err)
	}

	// Test parallel transition
	t.Run("ParallelTransition", func(t *testing.T) {
		ctx := OrderContext{
			OrderId: "ORD-20250425-001",
			Amount:  199.99,
			User:    "user1",
		}

		newStates, err := stateMachine.FireParallelEvent(Paid, Process, ctx)
		if err != nil {
			t.Fatalf("Failed to transition: %v", err)
		}

		if len(newStates) != 2 {
			t.Errorf("Expected 2 new states, got %d", len(newStates))
		}

		// Check that both target states are in the result
		hasDelivered := false
		hasNotified := false
		for _, state := range newStates {
			if state == Delivered {
				hasDelivered = true
			}
			if state == Notified {
				hasNotified = true
			}
		}

		if !hasDelivered {
			t.Error("Expected Delivered state in result")
		}
		if !hasNotified {
			t.Error("Expected Notified state in result")
		}
	})
}

// TestVisualization tests the visualization functionality
func TestVisualization(t *testing.T) {
	// Create state machine builder
	builder := fsm.NewStateMachineBuilder[OrderState, OrderEvent, OrderContext]()

	// Define basic transitions
	builder.ExternalTransition().
		From(Created).
		To(Paid).
		On(Pay)

	builder.ExternalTransition().
		From(Paid).
		To(Delivered).
		On(Deliver)

	// Build the state machine
	stateMachine, err := builder.Build("VisualizationStateMachine")
	if err != nil {
		t.Fatalf("Failed to build state machine: %v", err)
	}

	// Test default diagram (PlantUML)
	t.Run("DefaultDiagram", func(t *testing.T) {
		diagram := stateMachine.GenerateDiagram()
		if diagram == "" {
			t.Error("Expected non-empty diagram")
		}
		t.Logf("PlantUML diagram: %s", diagram)
	})

	// Test Markdown table
	t.Run("MarkdownTable", func(t *testing.T) {
		diagram := stateMachine.GenerateDiagram(fsm.MarkdownTable)
		if diagram == "" {
			t.Error("Expected non-empty diagram")
		}
		t.Logf("Markdown table: %s", diagram)
	})

	// Test Markdown flow
	t.Run("MarkdownFlowchart", func(t *testing.T) {
		diagram := stateMachine.GenerateDiagram(fsm.MarkdownFlowchart)
		if diagram == "" {
			t.Error("Expected non-empty diagram")
		}
		t.Logf("Markdown flow: %s", diagram)
	})

	// Test multiple formats
	t.Run("MultipleFormats", func(t *testing.T) {
		diagram := stateMachine.GenerateDiagram(fsm.PlantUML, fsm.MarkdownTable)
		if diagram == "" {
			t.Error("Expected non-empty diagram")
		}
		t.Logf("Combined diagram: %s", diagram)
	})
}
