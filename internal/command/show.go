package command

import "fmt"

// ShowCommand displays current settings
type ShowCommand struct{}

func NewShowCommand() *ShowCommand {
	return &ShowCommand{}
}

func (c *ShowCommand) Name() string {
	return "show"
}

func (c *ShowCommand) Aliases() []string {
	return []string{}
}

func (c *ShowCommand) Description() string {
	return "Display current system prompt and model"
}

func (c *ShowCommand) Usage() string {
	return "/show"
}

func (c *ShowCommand) Execute(ctx *CommandContext) ExecutionResult {
	ctx.Config.InfoColor().Printf("\nCurrent system prompt: ")
	fmt.Printf("%s\n", ctx.State.GetSystemPrompt())
	ctx.Config.InfoColor().Printf("Current model: ")
	fmt.Printf("%s\n", ctx.State.GetModel())
	fmt.Println()
	return REPLContinue
}
