package command

// ResetCommand resets the conversation history
type ResetCommand struct{}

func NewResetCommand() *ResetCommand {
	return &ResetCommand{}
}

func (c *ResetCommand) Name() string {
	return "reset"
}

func (c *ResetCommand) Aliases() []string {
	return []string{}
}

func (c *ResetCommand) Description() string {
	return "Reset conversation history"
}

func (c *ResetCommand) Usage() string {
	return "/reset"
}

func (c *ResetCommand) Execute(ctx *CommandContext) ExecutionResult {
	successColor := ctx.Config.PromptColor()
	ctx.History.Clear()

	successColor.Println("✓ Conversation history has been reset for current model")

	return REPLContinue
}
