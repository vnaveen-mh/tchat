package command

import (
	"fmt"
	"strings"
)

// Registry manages all available commands
type Registry struct {
	commands map[string]Command
}

// NewRegistry creates a new command registry
func NewRegistry() *Registry {
	return &Registry{
		commands: make(map[string]Command),
	}
}

// Register adds a command to the registry
func (r *Registry) Register(cmd Command) {
	// Register the primary name
	r.commands["/"+cmd.Name()] = cmd

	// Register all aliases
	for _, alias := range cmd.Aliases() {
		r.commands["/"+alias] = cmd
	}
}

// Get retrieves a command by name or alias
func (r *Registry) Get(name string) (Command, bool) {
	cmd, ok := r.commands[name]
	return cmd, ok
}

// IsCommand checks if the input is a registered command
func (r *Registry) IsCommand(input string) bool {
	_, ok := r.commands[input]
	return ok
}

// AllCommands returns all unique commands (without duplicates from aliases)
func (r *Registry) AllCommands() []Command {
	seen := make(map[string]bool)
	var commands []Command

	for _, cmd := range r.commands {
		name := cmd.Name()
		if !seen[name] {
			seen[name] = true
			commands = append(commands, cmd)
		}
	}

	return commands
}

// Help returns formatted help text for all commands
func (r *Registry) Help() string {
	var sb strings.Builder
	sb.WriteString("\nAvailable commands:\n")

	for _, cmd := range r.AllCommands() {
		// Build the command names (name + aliases)
		names := []string{"/" + cmd.Name()}
		for _, alias := range cmd.Aliases() {
			names = append(names, "/"+alias)
		}

		fmt.Fprintf(&sb, "  %-30s %s\n", strings.Join(names, ", "), cmd.Description())
	}

	return sb.String()
}
