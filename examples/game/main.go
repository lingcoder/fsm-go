package main

import (
	"fmt"
	"github.com/lingcoder/fsm-go"
	"log"
	"time"
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
	// In a real application, this would check if game resources are fully loaded
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
	fmt.Printf("Game state transitioning from %s to %s, event: %s\n", from, to, event)
	fmt.Printf("  Player ID: %s, Current Level: %d, Score: %d, Health: %d\n",
		ctx.PlayerID, ctx.CurrentLevel, ctx.Score, ctx.Health)

	// Add specific game state transition logic here
	// For example: loading resources, saving game state, updating UI, etc.

	return nil
}

// Game over action
type GameOverAction struct{}

func (a *GameOverAction) Execute(from, to GameState, event GameEvent, ctx GameContext) error {
	fmt.Printf("Game Over! Final score: %d\n", ctx.Score)
	fmt.Printf("  Player ID: %s, Level: %d\n", ctx.PlayerID, ctx.CurrentLevel)

	// Add game over logic here
	// For example: save score, show leaderboard, etc.

	return nil
}

// Victory action
type VictoryAction struct{}

func (a *VictoryAction) Execute(from, to GameState, event GameEvent, ctx GameContext) error {
	fmt.Printf("Congratulations! You completed the game! Final score: %d\n", ctx.Score)
	fmt.Printf("  Player ID: %s, Level: %d\n", ctx.PlayerID, ctx.CurrentLevel)

	// Add victory logic here
	// For example: unlock achievements, save progress, etc.

	return nil
}

func main() {
	// Create game state machine
	builder := fsm.NewStateMachineBuilder[GameState, GameEvent, GameContext]()

	// Define state transitions
	// From main menu to loading
	builder.ExternalTransition().
		From(MainMenu).
		To(Loading).
		On(StartGame).
		Perform(&GameStateAction{})

	// From loading to playing
	builder.ExternalTransition().
		From(Loading).
		To(Playing).
		On(StartGame).
		When(&ResourceLoadedCondition{}).
		Perform(&GameStateAction{})

	// From playing to paused
	builder.ExternalTransition().
		From(Playing).
		To(Paused).
		On(PauseGame).
		Perform(&GameStateAction{})

	// From paused to playing
	builder.ExternalTransition().
		From(Paused).
		To(Playing).
		On(ResumeGame).
		Perform(&GameStateAction{})

	// From playing to game over
	builder.ExternalTransition().
		From(Playing).
		To(GameOver).
		On(PlayerDied).
		Perform(&GameOverAction{})

	// From playing to victory
	builder.ExternalTransition().
		From(Playing).
		To(Victory).
		On(LevelComplete).
		Perform(&VictoryAction{})

	// From playing to settings
	builder.ExternalTransition().
		From(Playing).
		To(Settings).
		On(OpenSettings).
		Perform(&GameStateAction{})

	// From settings to playing
	builder.ExternalTransition().
		From(Settings).
		To(Playing).
		On(CloseSettings).
		Perform(&GameStateAction{})

	// From playing to inventory
	builder.ExternalTransition().
		From(Playing).
		To(Inventory).
		On(OpenInventory).
		Perform(&GameStateAction{})

	// From inventory to playing
	builder.ExternalTransition().
		From(Inventory).
		To(Playing).
		On(CloseInventory).
		Perform(&GameStateAction{})

	// Multiple states can return to main menu
	builder.ExternalTransitions().
		FromAmong(Paused, GameOver, Victory, Settings).
		To(MainMenu).
		On(ReturnToMenu).
		Perform(&GameStateAction{})

	// Build the state machine
	sm, err := builder.Build("GameStateMachine")
	if err != nil {
		log.Fatalf("Failed to create game state machine: %v", err)
	}

	// Display state machine structure
	fmt.Println("Game State Machine:")
	fmt.Println(sm.ShowStateMachine())

	// Create game context
	ctx := GameContext{
		PlayerID:     "PLAYER-001",
		CurrentLevel: 1,
		Score:        0,
		Health:       100,
		LastSaveTime: time.Now(),
	}

	// Simulate game flow
	fmt.Println("\nStarting game flow simulation:")

	// Start game
	currentState := MainMenu

	// From main menu to loading
	newState, err := sm.FireEvent(currentState, StartGame, ctx)
	if err != nil {
		log.Fatalf("State transition failed: %v", err)
	}
	currentState = newState
	fmt.Printf("\nCurrent state: %s\n", currentState)

	// From loading to playing
	newState, err = sm.FireEvent(currentState, StartGame, ctx)
	if err != nil {
		log.Fatalf("State transition failed: %v", err)
	}
	currentState = newState
	fmt.Printf("\nCurrent state: %s\n", currentState)

	// Open inventory
	newState, err = sm.FireEvent(currentState, OpenInventory, ctx)
	if err != nil {
		log.Fatalf("State transition failed: %v", err)
	}
	currentState = newState
	fmt.Printf("\nCurrent state: %s\n", currentState)

	// Close inventory, return to game
	newState, err = sm.FireEvent(currentState, CloseInventory, ctx)
	if err != nil {
		log.Fatalf("State transition failed: %v", err)
	}
	currentState = newState
	fmt.Printf("\nCurrent state: %s\n", currentState)

	// Update game context
	ctx.Score = 1000
	ctx.CurrentLevel = 2

	// Complete level
	newState, err = sm.FireEvent(currentState, LevelComplete, ctx)
	if err != nil {
		log.Fatalf("State transition failed: %v", err)
	}
	currentState = newState
	fmt.Printf("\nCurrent state: %s\n", currentState)

	// Return to main menu
	newState, err = sm.FireEvent(currentState, ReturnToMenu, ctx)
	if err != nil {
		log.Fatalf("State transition failed: %v", err)
	}
	currentState = newState
	fmt.Printf("\nCurrent state: %s\n", currentState)

	// Generate PlantUML diagram
	fmt.Println("\nGame State Machine Diagram:")
	fmt.Println(sm.GeneratePlantUML())
}
