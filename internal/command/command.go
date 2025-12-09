package command

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"tchat/internal/appstate"
	"tchat/internal/config"
	"tchat/internal/history"

	"github.com/chzyer/readline"
)

// ExecutionResult represents the outcome of a command execution
type ExecutionResult int

const (
	// Continue indicates normal execution, continue the REPL loop
	REPLContinue ExecutionResult = iota
	// Exit indicates the REPL should terminate
	REPLExit
)

// CommandContext provides the runtime context for command execution
type CommandContext struct {
	Ctx          context.Context
	Config       *config.Config
	State        *appstate.State
	Readline     *readline.Instance
	History      *history.HistoryManager
	LastResponse *string
}

// Command represents a special command that can be executed in the REPL
type Command interface {
	// Name returns the primary name of the command (e.g., "quit")
	Name() string

	// Aliases returns alternative names for the command (e.g., ["exit", "bye"])
	Aliases() []string

	// Description returns a brief description of what the command does
	Description() string

	// Usage returns usage information for the command (optional)
	Usage() string

	// Execute runs the command with the given context
	Execute(ctx *CommandContext) ExecutionResult
}

// ReadInputWithoutHistory reads user input without adding it to command history.
// This is useful for selections and confirmations that shouldn't clutter history.
func ReadInputWithoutHistory(prompt string) (string, error) {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}
