package command

// QuitCommand handles exiting the application
type QuitCommand struct{}

func NewQuitCommand() *QuitCommand {
	return &QuitCommand{}
}

func (c *QuitCommand) Name() string {
	return "quit"
}

func (c *QuitCommand) Aliases() []string {
	return []string{"exit", "bye"}
}

func (c *QuitCommand) Description() string {
	return "Exit the AI assistant"
}

func (c *QuitCommand) Usage() string {
	return "/quit, /exit, or /bye"
}

func (c *QuitCommand) Execute(ctx *CommandContext) ExecutionResult {
	return REPLExit
}
