package command

import (
	"fmt"
)

// SystemCommand handles updating the system prompt
type SystemCommand struct{}

func NewSystemCommand() *SystemCommand {
	return &SystemCommand{}
}

func (c *SystemCommand) Name() string {
	return "system"
}

func (c *SystemCommand) Aliases() []string {
	return []string{"sysprompt"}
}

func (c *SystemCommand) Description() string {
	return "View or update the system prompt"
}

func (c *SystemCommand) Usage() string {
	return "/system or /sysprompt - then enter new prompt or press Enter to keep current"
}

func (c *SystemCommand) Execute(ctx *CommandContext) ExecutionResult {
	ctx.Config.InfoColor().Printf("\nCurrent system prompt:")
	fmt.Println(ctx.Config.GetSystemPrompt())

	msg := "Enter a new system prompt (press Enter to keep current): "

	// Read input without adding to command history
	newPrompt, err := ReadInputWithoutHistory(msg)
	if err != nil {
		return REPLExit
	}
	if newPrompt != "" {
		ctx.State.SetSystemPrompt(newPrompt)
		fmt.Println("System prompt updated successfully")
	} else {
		fmt.Println("System prompt unchanged")
	}

	return REPLContinue
}
