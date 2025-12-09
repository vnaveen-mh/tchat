package command

import (
	"fmt"
)

// HistoryCommand displays history statistics
type HistoryCommand struct{}

func NewHistoryCommand() *HistoryCommand {
	return &HistoryCommand{}
}

func (c *HistoryCommand) Name() string {
	return "history"
}

func (c *HistoryCommand) Aliases() []string {
	return []string{"hist"}
}

func (c *HistoryCommand) Description() string {
	return "Display conversation history statistics"
}

func (c *HistoryCommand) Usage() string {
	return "/history or /hist"
}

func (c *HistoryCommand) Execute(ctx *CommandContext) ExecutionResult {
	if ctx.History.IsEmpty() {
		fmt.Println("No conversation history")
		return REPLContinue
	}

	stats := ctx.History.GetStats()

	ctx.Config.InfoColor().Println("\nConversation History")
	ctx.Config.InfoColor().Println("===================")

	fmt.Printf("Total messages:     %d\n", stats.TotalMessages)
	fmt.Printf("User messages:      %d\n", stats.UserMessages)
	fmt.Printf("Assistant messages: %d\n", stats.AssistantMessages)
	fmt.Printf("Conversation pairs: %d\n", stats.UserMessages) // User messages = pairs

	fmt.Println()
	ctx.Config.InfoColor().Println("Commands:")
	fmt.Println("  /reset  - Clear conversation history")
	fmt.Println()

	return REPLContinue
}
