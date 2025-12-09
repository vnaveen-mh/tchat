package command

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

// ClearCommand clears the terminal screen
type ClearCommand struct{}

func NewClearCommand() *ClearCommand {
	return &ClearCommand{}
}

func (c *ClearCommand) Name() string {
	return "clear"
}

func (c *ClearCommand) Aliases() []string {
	return []string{"cls"}
}

func (c *ClearCommand) Description() string {
	return "Clear the terminal screen"
}

func (c *ClearCommand) Usage() string {
	return "/clear or /cls"
}

func (c *ClearCommand) Execute(ctx *CommandContext) ExecutionResult {
	clearScreen()
	return REPLContinue
}

// clearScreen clears the terminal screen using platform-specific methods
func clearScreen() {
	// Try ANSI escape codes first (works on most modern terminals)
	fmt.Print("\033[H\033[2J")

	// For Windows, also try the cls command as fallback
	if runtime.GOOS == "windows" {
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	} else {
		// For Unix-like systems, use clear command as fallback
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}
