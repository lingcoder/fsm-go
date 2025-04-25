package benchmark

import (
	"fmt"
	"testing"
	"time"

	"github.com/lingcoder/fsm-go"
)

// BenchmarkDiagramGeneration tests the performance of diagram generation
func BenchmarkDiagramGeneration(b *testing.B) {
	b.ReportAllocs()

	// Create a state machine
	builder := fsm.NewStateMachineBuilder[string, string, interface{}]()

	// Define multiple transitions
	builder.ExternalTransition().
		From("A").
		To("B").
		On("EVENT1")

	builder.ExternalTransition().
		From("B").
		To("C").
		On("EVENT2")

	builder.ExternalTransition().
		From("C").
		To("D").
		On("EVENT3")

	builder.ExternalTransition().
		From("D").
		To("A").
		On("EVENT1")

	stateMachine, err := builder.Build(fmt.Sprintf("DiagramMachine-%d", time.Now().UnixNano()))
	if err != nil {
		b.Fatalf("Failed to build state machine: %v", err)
	}

	// Run sub-benchmarks for different formats
	b.Run("PlantUML", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = stateMachine.GenerateDiagram(fsm.PlantUML)
		}
	})

	b.Run("MarkdownTable", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = stateMachine.GenerateDiagram(fsm.MarkdownTable)
		}
	})

	b.Run("MarkdownFlow", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = stateMachine.GenerateDiagram(fsm.MarkdownFlow)
		}
	})

	// Test combined formats
	b.Run("AllFormats", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = stateMachine.GenerateDiagram(fsm.PlantUML, fsm.MarkdownTable, fsm.MarkdownFlow)
		}
	})
}
