package command

import (
	"fmt"
	"tchat/internal/db"
)

// StatsCommand displays database statistics
type StatsCommand struct {
	store *db.Store
}

func NewStatsCommand(store *db.Store) *StatsCommand {
	return &StatsCommand{
		store: store,
	}
}

func (c *StatsCommand) Name() string {
	return "stats"
}

func (c *StatsCommand) Aliases() []string {
	return []string{}
}

func (c *StatsCommand) Description() string {
	return "Display conversation statistics from database"
}

func (c *StatsCommand) Usage() string {
	return "/stats"
}

func (c *StatsCommand) Execute(ctx *CommandContext) ExecutionResult {
	if c.store == nil {
		fmt.Println("Database storage is not available")
		return REPLContinue
	}

	stats, err := c.store.GetStats()
	if err != nil {
		ctx.Config.ErrorColor().Printf("Failed to retrieve statistics: %v\n", err)
		return REPLContinue
	}

	ctx.Config.InfoColor().Println("\nConversation Statistics")
	ctx.Config.InfoColor().Println("======================")

	if total, ok := stats["total_conversations"].(int); ok && total == 0 {
		fmt.Println("No conversations stored yet")
		return REPLContinue
	}

	// Display statistics
	if val, ok := stats["total_conversations"]; ok {
		fmt.Printf("Total conversations:  %d\n", val)
	}
	if val, ok := stats["unique_models"]; ok {
		fmt.Printf("Unique models used:   %d\n", val)
	}
	if val, ok := stats["avg_duration_ms"]; ok {
		fmt.Printf("Avg response time:    %.0f ms\n", val)
	}
	if val, ok := stats["min_duration_ms"]; ok {
		fmt.Printf("Min response time:    %.0f ms\n", val)
	}
	if val, ok := stats["max_duration_ms"]; ok {
		fmt.Printf("Max response time:    %.0f ms\n", val)
	}
	if val, ok := stats["avg_input_length"]; ok {
		fmt.Printf("Avg input length:     %.0f chars\n", val)
	}
	if val, ok := stats["avg_output_length"]; ok {
		fmt.Printf("Avg output length:    %.0f chars\n", val)
	}

	fmt.Println()

	return REPLContinue
}
