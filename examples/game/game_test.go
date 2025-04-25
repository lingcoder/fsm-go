package game

import (
	"testing"
	"time"

	"github.com/lingcoder/fsm-go"
)

// Game states
type GameState string

const (
	MainMenu  GameState = "MAIN_MENU"
	Loading   GameState = "LOADING"
	Playing   GameState = "PLAYING"
	Paused    GameState = "PAUSED"
	GameOver  GameState = "GAME_OVER"
	Victory   GameState = "VICTORY"
	Settings  GameState = "SETTINGS"
	Inventory GameState = "INVENTORY"
)

// Game events
type GameEvent string

const (
	StartGame      GameEvent = "START_GAME"
	PauseGame      GameEvent = "PAUSE_GAME"
	ResumeGame     GameEvent = "RESUME_GAME"
	PlayerDied     GameEvent = "PLAYER_DIED"
	LevelComplete  GameEvent = "LEVEL_COMPLETE"
	OpenSettings   GameEvent = "OPEN_SETTINGS"
	CloseSettings  GameEvent = "CLOSE_SETTINGS"
	OpenInventory  GameEvent = "OPEN_INVENTORY"
	CloseInventory GameEvent = "CLOSE_INVENTORY"
	ReturnToMenu   GameEvent = "RETURN_TO_MENU"
)

// Game context
type GameContext struct {
	PlayerID       string
	CurrentLevel   int
	Score          int
	Health         int
	IsLoadingSaved bool
	LastSaveTime   time.Time
}

// Resource loading condition
type ResourceLoadedCondition struct{}

func (c *ResourceLoadedCondition) IsSatisfied(ctx GameContext) bool {
	// Simulate resource loading check
	return true
}

// Player alive condition
type PlayerAliveCondition struct{}

func (c *PlayerAliveCondition) IsSatisfied(ctx GameContext) bool {
	return ctx.Health > 0
}

// Game state transition action
type GameStateAction struct{}

func (a *GameStateAction) Execute(from, to GameState, event GameEvent, ctx GameContext) error {
	// In a real game, this would handle transition logic
	return nil
}

// Game over action
type GameOverAction struct{}

func (a *GameOverAction) Execute(from, to GameState, event GameEvent, ctx GameContext) error {
	// In a real game, this would handle game over logic
	return nil
}

// Victory action
type VictoryAction struct{}

func (a *VictoryAction) Execute(from, to GameState, event GameEvent, ctx GameContext) error {
	// In a real game, this would handle victory logic
	return nil
}

// TestGameStateMachine tests the basic functionality of a game state machine
func TestGameStateMachine(t *testing.T) {
	// Create state machine builder
	builder := fsm.NewStateMachineBuilder[GameState, GameEvent, GameContext]()

	// Create conditions and actions
	resourceLoaded := &ResourceLoadedCondition{}
	playerAlive := &PlayerAliveCondition{}
	gameStateAction := &GameStateAction{}
	gameOverAction := &GameOverAction{}
	victoryAction := &VictoryAction{}

	// Main menu to loading
	builder.ExternalTransition().
		From(MainMenu).
		To(Loading).
		On(StartGame).
		Perform(gameStateAction)

	// Loading to playing
	builder.ExternalTransition().
		From(Loading).
		To(Playing).
		When(resourceLoaded).
		Perform(gameStateAction)

	// Playing to paused
	builder.ExternalTransition().
		From(Playing).
		To(Paused).
		On(PauseGame).
		Perform(gameStateAction)

	// Paused to playing
	builder.ExternalTransition().
		From(Paused).
		To(Playing).
		On(ResumeGame).
		Perform(gameStateAction)

	// Playing to game over
	builder.ExternalTransition().
		From(Playing).
		To(GameOver).
		On(PlayerDied).
		Perform(gameOverAction)

	// Playing to victory
	builder.ExternalTransition().
		From(Playing).
		To(Victory).
		On(LevelComplete).
		When(playerAlive).
		Perform(victoryAction)

	// Playing to settings
	builder.ExternalTransition().
		From(Playing).
		To(Settings).
		On(OpenSettings).
		Perform(gameStateAction)

	// Settings to playing
	builder.ExternalTransition().
		From(Settings).
		To(Playing).
		On(CloseSettings).
		Perform(gameStateAction)

	// Playing to inventory
	builder.ExternalTransition().
		From(Playing).
		To(Inventory).
		On(OpenInventory).
		Perform(gameStateAction)

	// Inventory to playing
	builder.ExternalTransition().
		From(Inventory).
		To(Playing).
		On(CloseInventory).
		Perform(gameStateAction)

	// Return to main menu from various states
	builder.ExternalTransitions().
		FromAmong(Paused, GameOver, Victory).
		To(MainMenu).
		On(ReturnToMenu).
		Perform(gameStateAction)

	// Build the state machine
	stateMachine, err := builder.Build("GameStateMachine")
	if err != nil {
		t.Fatalf("Failed to build state machine: %v", err)
	}

	// Test game flow
	t.Run("GameFlow", func(t *testing.T) {
		ctx := GameContext{
			PlayerID:     "player1",
			CurrentLevel: 1,
			Score:        0,
			Health:       100,
		}

		// Start game
		state, err := stateMachine.FireEvent(MainMenu, StartGame, ctx)
		if err != nil {
			t.Fatalf("Failed to start game: %v", err)
		}
		if state != Loading {
			t.Errorf("Expected state to be %s, got %s", Loading, state)
		}

		// Load resources and start playing
		state, err = stateMachine.FireEvent(state, GameEvent(""), ctx) // Auto-transition
		if err != nil {
			t.Fatalf("Failed to load resources: %v", err)
		}
		if state != Playing {
			t.Errorf("Expected state to be %s, got %s", Playing, state)
		}

		// Pause game
		state, err = stateMachine.FireEvent(state, PauseGame, ctx)
		if err != nil {
			t.Fatalf("Failed to pause game: %v", err)
		}
		if state != Paused {
			t.Errorf("Expected state to be %s, got %s", Paused, state)
		}

		// Resume game
		state, err = stateMachine.FireEvent(state, ResumeGame, ctx)
		if err != nil {
			t.Fatalf("Failed to resume game: %v", err)
		}
		if state != Playing {
			t.Errorf("Expected state to be %s, got %s", Playing, state)
		}

		// Open inventory
		state, err = stateMachine.FireEvent(state, OpenInventory, ctx)
		if err != nil {
			t.Fatalf("Failed to open inventory: %v", err)
		}
		if state != Inventory {
			t.Errorf("Expected state to be %s, got %s", Inventory, state)
		}

		// Close inventory
		state, err = stateMachine.FireEvent(state, CloseInventory, ctx)
		if err != nil {
			t.Fatalf("Failed to close inventory: %v", err)
		}
		if state != Playing {
			t.Errorf("Expected state to be %s, got %s", Playing, state)
		}

		// Complete level
		state, err = stateMachine.FireEvent(state, LevelComplete, ctx)
		if err != nil {
			t.Fatalf("Failed to complete level: %v", err)
		}
		if state != Victory {
			t.Errorf("Expected state to be %s, got %s", Victory, state)
		}

		// Return to main menu
		state, err = stateMachine.FireEvent(state, ReturnToMenu, ctx)
		if err != nil {
			t.Fatalf("Failed to return to main menu: %v", err)
		}
		if state != MainMenu {
			t.Errorf("Expected state to be %s, got %s", MainMenu, state)
		}
	})

	// Test game over
	t.Run("GameOver", func(t *testing.T) {
		ctx := GameContext{
			PlayerID:     "player2",
			CurrentLevel: 1,
			Score:        0,
			Health:       0, // Player is dead
		}

		// Start game and get to playing state
		state, _ := stateMachine.FireEvent(MainMenu, StartGame, ctx)
		state, _ = stateMachine.FireEvent(state, GameEvent(""), ctx) // Auto-transition

		// Player dies
		state, err := stateMachine.FireEvent(state, PlayerDied, ctx)
		if err != nil {
			t.Fatalf("Failed to process player death: %v", err)
		}
		if state != GameOver {
			t.Errorf("Expected state to be %s, got %s", GameOver, state)
		}

		// Return to main menu
		state, err = stateMachine.FireEvent(state, ReturnToMenu, ctx)
		if err != nil {
			t.Fatalf("Failed to return to main menu: %v", err)
		}
		if state != MainMenu {
			t.Errorf("Expected state to be %s, got %s", MainMenu, state)
		}
	})

	// Test visualization
	t.Run("Visualization", func(t *testing.T) {
		// Generate diagrams
		plantUML := stateMachine.GenerateDiagram(fsm.PlantUML)
		markdown := stateMachine.GenerateDiagram(fsm.MarkdownTable)
		flow := stateMachine.GenerateDiagram(fsm.MarkdownFlowchart)

		if plantUML == "" || markdown == "" || flow == "" {
			t.Error("Expected non-empty diagrams")
		}

		t.Logf("Generated %d characters of PlantUML diagram", len(plantUML))
		t.Logf("Generated %d characters of Markdown table", len(markdown))
		t.Logf("Generated %d characters of Markdown flow chart", len(flow))
	})
}
