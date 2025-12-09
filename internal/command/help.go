package command

import "fmt"

// HelpCommand displays all available commands
type HelpCommand struct {
	registry *Registry
}

func NewHelpCommand(registry *Registry) *HelpCommand {
	return &HelpCommand{
		registry: registry,
	}
}

func (c *HelpCommand) Name() string {
	return "help"
}

func (c *HelpCommand) Aliases() []string {
	return []string{"?"}
}

func (c *HelpCommand) Description() string {
	return "Display this help message"
}

func (c *HelpCommand) Usage() string {
	return "/help"
}

func (c *HelpCommand) Execute(ctx *CommandContext) ExecutionResult {
	fmt.Println(c.registry.Help())
	return REPLContinue
}
