package command

import (
	"fmt"

	"golang.design/x/clipboard"
)

// CopyCommand copies the last AI response to clipboard
type CopyCommand struct{}

func NewCopyCommand() *CopyCommand {
	return &CopyCommand{}
}

func (c *CopyCommand) Name() string {
	return "copy"
}

func (c *CopyCommand) Aliases() []string {
	return []string{"cp"}
}

func (c *CopyCommand) Description() string {
	return "Copy the last AI response to clipboard"
}

func (c *CopyCommand) Usage() string {
	return "/copy or /cp"
}

func (c *CopyCommand) Execute(ctx *CommandContext) ExecutionResult {
	successColor := ctx.Config.PromptColor()
	if ctx.LastResponse == nil || *ctx.LastResponse == "" {
		fmt.Println("No response to copy yet")
		return REPLContinue
	}

	// Write to clipboard
	clipboard.Write(clipboard.FmtText, []byte(*ctx.LastResponse))

	responseLen := len(*ctx.LastResponse)
	successColor.Printf("✓ Copied last response to clipboard (%d characters)\n", responseLen)

	return REPLContinue
}
