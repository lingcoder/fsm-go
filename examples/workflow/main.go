package main

import (
	"fmt"
	"github.com/lingcoder/fsm-go"
	"log"
	"time"
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
	// Check if a reviewer is assigned
	return ctx.Reviewer != ""
}

// Approval action
type ApprovalAction struct{}

func (a *ApprovalAction) Execute(from, to ApprovalState, event ApprovalEvent, ctx ApprovalContext) error {
	fmt.Printf("Document %s transitioning from %s to %s, event: %s\n",
		ctx.DocumentID, from, to, event)
	fmt.Printf("  Requester: %s, Reviewer: %s\n", ctx.Requester, ctx.Reviewer)
	fmt.Printf("  Comments: %s\n", ctx.Comments)
	fmt.Printf("  Last updated: %s\n", time.Now().Format(time.RFC3339))

	return nil
}

func main() {
	// Create approval workflow state machine
	builder := fsm.NewStateMachineBuilder[ApprovalState, ApprovalEvent, ApprovalContext]()

	// Define state transitions
	// From Draft to Submitted
	builder.ExternalTransition().
		From(Draft).
		To(Submitted).
		On(Submit).
		When(&SubmitCondition{}).
		Perform(&ApprovalAction{})

	// From Submitted to InReview
	builder.ExternalTransition().
		From(Submitted).
		To(InReview).
		On(Review).
		When(&ReviewerCondition{}).
		Perform(&ApprovalAction{})

	// From InReview to Approved
	builder.ExternalTransition().
		From(InReview).
		To(Approved).
		On(Approve).
		When(&ReviewerCondition{}).
		Perform(&ApprovalAction{})

	// From InReview to Rejected
	builder.ExternalTransition().
		From(InReview).
		To(Rejected).
		On(Reject).
		When(&ReviewerCondition{}).
		Perform(&ApprovalAction{})

	// From Rejected to Submitted (resubmission)
	builder.ExternalTransition().
		From(Rejected).
		To(Submitted).
		On(Resubmit).
		When(&SubmitCondition{}).
		Perform(&ApprovalAction{})

	// Multiple states can be cancelled
	builder.ExternalTransitions().
		FromAmong(Draft, Submitted, InReview).
		To(Cancelled).
		On(Cancel).
		Perform(&ApprovalAction{})

	// Build the state machine
	sm, err := builder.Build("ApprovalWorkflow")
	if err != nil {
		log.Fatalf("Failed to create approval workflow state machine: %v", err)
	}

	// Display state machine structure
	fmt.Println("Approval Workflow State Machine:")
	fmt.Println(sm.ShowStateMachine())

	// Create context
	ctx := ApprovalContext{
		DocumentID:  "DOC-2025-001",
		Requester:   "John Doe",
		Comments:    "Please approve my expense report",
		SubmittedAt: time.Now(),
	}

	// Simulate approval workflow
	fmt.Println("\nStarting approval workflow simulation:")

	// Submit document
	newState, err := sm.FireEvent(Draft, Submit, ctx)
	if err != nil {
		log.Fatalf("Failed to submit document: %v", err)
	}
	fmt.Printf("\nDocument submitted, current state: %s\n", newState)

	// Update context
	ctx.Reviewer = "Jane Smith"
	ctx.LastUpdatedAt = time.Now()

	// Start review
	newState, err = sm.FireEvent(newState, Review, ctx)
	if err != nil {
		log.Fatalf("Failed to start review: %v", err)
	}
	fmt.Printf("\nReview started, current state: %s\n", newState)

	// Add review comments
	ctx.Comments = "Amount is reasonable, expense approved"
	ctx.LastUpdatedAt = time.Now()

	// Approve document
	newState, err = sm.FireEvent(newState, Approve, ctx)
	if err != nil {
		log.Fatalf("Failed to approve document: %v", err)
	}
	fmt.Printf("\nDocument approved, current state: %s\n", newState)

	// Generate PlantUML diagram
	fmt.Println("\nApproval Workflow State Machine Diagram:")
	fmt.Println(sm.GeneratePlantUML())
}
