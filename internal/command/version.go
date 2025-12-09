package command

import (
	"fmt"

	"tchat/internal/version"
)

// VersionCommand displays version and build information
type VersionCommand struct{}

func NewVersionCommand() *VersionCommand {
	return &VersionCommand{}
}

func (c *VersionCommand) Name() string {
	return "version"
}

func (c *VersionCommand) Aliases() []string {
	return []string{"v"}
}

func (c *VersionCommand) Description() string {
	return "Display version and build information"
}

func (c *VersionCommand) Usage() string {
	return "/version or /v"
}

func (c *VersionCommand) Execute(ctx *CommandContext) ExecutionResult {
	titleColor := ctx.Config.PromptColor()
	titleColor.Println("\nTerminal AI Assistant")
	titleColor.Println("====================")

	info := version.Get()
	fmt.Println(info.String())
	fmt.Println()

	return REPLContinue
}
