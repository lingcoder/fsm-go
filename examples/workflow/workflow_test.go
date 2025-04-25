package workflow

import (
	"testing"
	"time"

	"github.com/lingcoder/fsm-go"
)

// Define approval workflow states
type ApprovalState string

const (
	Draft     ApprovalState = "DRAFT"
	Submitted ApprovalState = "SUBMITTED"
	InReview  ApprovalState = "IN_REVIEW"
	Approved  ApprovalState = "APPROVED"
	Rejected  ApprovalState = "REJECTED"
	Cancelled ApprovalState = "CANCELLED"
)

// Define approval workflow events
type ApprovalEvent string

const (
	Submit   ApprovalEvent = "SUBMIT"
	Review   ApprovalEvent = "REVIEW"
	Approve  ApprovalEvent = "APPROVE"
	Reject   ApprovalEvent = "REJECT"
	Cancel   ApprovalEvent = "CANCEL"
	Resubmit ApprovalEvent = "RESUBMIT"
)

// Approval workflow context
type ApprovalContext struct {
	DocumentID    string
	Requester     string
	Reviewer      string
	Comments      string
	SubmittedAt   time.Time
	LastUpdatedAt time.Time
}

// Submission condition
type SubmitCondition struct{}

func (c *SubmitCondition) IsSatisfied(ctx ApprovalContext) bool {
	// Check if document is complete
	return ctx.DocumentID != "" && ctx.Requester != ""
}

// Reviewer check condition
type ReviewerCondition struct{}

func (c *ReviewerCondition) IsSatisfied(ctx ApprovalContext) bool {
	// Check if reviewer is assigned
	return ctx.Reviewer != ""
}

// Approval action
type ApprovalAction struct{}

func (a *ApprovalAction) Execute(from, to ApprovalState, event ApprovalEvent, ctx ApprovalContext) error {
	// In a real application, this would perform the actual state transition logic
	return nil
}

// TestApprovalWorkflow tests the basic functionality of an approval workflow state machine
func TestApprovalWorkflow(t *testing.T) {
	// Create state machine builder
	builder := fsm.NewStateMachineBuilder[ApprovalState, ApprovalEvent, ApprovalContext]()

	// Create conditions and actions
	submitCondition := &SubmitCondition{}
	reviewerCondition := &ReviewerCondition{}
	approvalAction := &ApprovalAction{}

	// Draft to Submitted
	builder.ExternalTransition().
		From(Draft).
		To(Submitted).
		On(Submit).
		When(submitCondition).
		Perform(approvalAction)

	// Submitted to InReview
	builder.ExternalTransition().
		From(Submitted).
		To(InReview).
		On(Review).
		When(reviewerCondition).
		Perform(approvalAction)

	// InReview to Approved
	builder.ExternalTransition().
		From(InReview).
		To(Approved).
		On(Approve).
		Perform(approvalAction)

	// InReview to Rejected
	builder.ExternalTransition().
		From(InReview).
		To(Rejected).
		On(Reject).
		Perform(approvalAction)

	// Rejected to Submitted
	builder.ExternalTransition().
		From(Rejected).
		To(Submitted).
		On(Resubmit).
		When(submitCondition).
		Perform(approvalAction)

	// Cancel from multiple states
	builder.ExternalTransitions().
		FromAmong(Draft, Submitted, InReview).
		To(Cancelled).
		On(Cancel).
		Perform(approvalAction)

	// Build the state machine
	stateMachine, err := builder.Build("ApprovalWorkflow")
	if err != nil {
		t.Fatalf("Failed to build state machine: %v", err)
	}

	// Test happy path workflow
	t.Run("HappyPath", func(t *testing.T) {
		ctx := ApprovalContext{
			DocumentID:    "DOC-20250425-001",
			Requester:     "John Doe",
			Reviewer:      "Jane Smith",
			SubmittedAt:   time.Now(),
			LastUpdatedAt: time.Now(),
		}

		// Submit document
		state, err := stateMachine.FireEvent(Draft, Submit, ctx)
		if err != nil {
			t.Fatalf("Failed to submit document: %v", err)
		}
		if state != Submitted {
			t.Errorf("Expected state to be %s, got %s", Submitted, state)
		}

		// Start review
		state, err = stateMachine.FireEvent(state, Review, ctx)
		if err != nil {
			t.Fatalf("Failed to start review: %v", err)
		}
		if state != InReview {
			t.Errorf("Expected state to be %s, got %s", InReview, state)
		}

		// Approve document
		state, err = stateMachine.FireEvent(state, Approve, ctx)
		if err != nil {
			t.Fatalf("Failed to approve document: %v", err)
		}
		if state != Approved {
			t.Errorf("Expected state to be %s, got %s", Approved, state)
		}
	})

	// Test rejection workflow
	t.Run("RejectionPath", func(t *testing.T) {
		ctx := ApprovalContext{
			DocumentID:    "DOC-20250425-002",
			Requester:     "John Doe",
			Reviewer:      "Jane Smith",
			SubmittedAt:   time.Now(),
			LastUpdatedAt: time.Now(),
		}

		// Submit document
		state, err := stateMachine.FireEvent(Draft, Submit, ctx)
		if err != nil {
			t.Fatalf("Failed to submit document: %v", err)
		}

		// Start review
		state, err = stateMachine.FireEvent(state, Review, ctx)
		if err != nil {
			t.Fatalf("Failed to start review: %v", err)
		}

		// Reject document
		state, err = stateMachine.FireEvent(state, Reject, ctx)
		if err != nil {
			t.Fatalf("Failed to reject document: %v", err)
		}
		if state != Rejected {
			t.Errorf("Expected state to be %s, got %s", Rejected, state)
		}

		// Resubmit document
		state, err = stateMachine.FireEvent(state, Resubmit, ctx)
		if err != nil {
			t.Fatalf("Failed to resubmit document: %v", err)
		}
		if state != Submitted {
			t.Errorf("Expected state to be %s, got %s", Submitted, state)
		}
	})

	// Test cancellation
	t.Run("CancellationPath", func(t *testing.T) {
		ctx := ApprovalContext{
			DocumentID:    "DOC-20250425-003",
			Requester:     "John Doe",
			Reviewer:      "Jane Smith",
			SubmittedAt:   time.Now(),
			LastUpdatedAt: time.Now(),
		}

		// Submit document
		state, err := stateMachine.FireEvent(Draft, Submit, ctx)
		if err != nil {
			t.Fatalf("Failed to submit document: %v", err)
		}

		// Cancel document
		state, err = stateMachine.FireEvent(state, Cancel, ctx)
		if err != nil {
			t.Fatalf("Failed to cancel document: %v", err)
		}
		if state != Cancelled {
			t.Errorf("Expected state to be %s, got %s", Cancelled, state)
		}
	})

	// Test validation failures
	t.Run("ValidationFailures", func(t *testing.T) {
		// Missing document ID
		ctx := ApprovalContext{
			Requester: "John Doe",
			Reviewer:  "Jane Smith",
		}

		_, err := stateMachine.FireEvent(Draft, Submit, ctx)
		if err == nil {
			t.Error("Expected error when submitting document with missing ID")
		}

		// Missing reviewer
		ctx = ApprovalContext{
			DocumentID: "DOC-20250425-004",
			Requester:  "John Doe",
			// No reviewer
		}

		state, _ := stateMachine.FireEvent(Draft, Submit, ctx)
		_, err = stateMachine.FireEvent(state, Review, ctx)
		if err == nil {
			t.Error("Expected error when reviewing document with no reviewer")
		}
	})

	// Test visualization
	t.Run("Visualization", func(t *testing.T) {
		// Generate all diagram formats
		diagrams := stateMachine.GenerateDiagram(fsm.PlantUML, fsm.MarkdownTable, fsm.MarkdownFlowchart, fsm.MarkdownStateDiagram)
		if diagrams == "" {
			t.Error("Expected non-empty diagrams")
		}

		t.Logf("Generated %d characters of combined diagrams", len(diagrams))
		t.Log(diagrams)
	})
}
