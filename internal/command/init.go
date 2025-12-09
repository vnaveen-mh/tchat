package command

import "tchat/internal/db"

// InitializeRegistry creates and registers all available commands
func InitializeRegistry(availableModels []string, store *db.Store) *Registry {
	registry := NewRegistry()

	// Create help command with registry reference (will be set after other commands)
	helpCmd := NewHelpCommand(registry)

	// Register all commands
	registry.Register(NewQuitCommand())
	registry.Register(NewSystemCommand())
	registry.Register(NewModelCommand(availableModels))
	registry.Register(NewShowCommand())
	registry.Register(NewConfigCommand())
	registry.Register(NewClearCommand())
	registry.Register(NewResetCommand())
	registry.Register(NewHistoryCommand())
	registry.Register(NewCopyCommand())
	registry.Register(NewVersionCommand())
	registry.Register(NewStatsCommand(store))
	registry.Register(helpCmd)

	return registry
}
