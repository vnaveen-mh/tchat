package command

import (
	"fmt"
	"strconv"
	"strings"
)

// ModelCommand handles model switching
type ModelCommand struct {
	availableModels []string
}

func NewModelCommand(models []string) *ModelCommand {
	return &ModelCommand{
		availableModels: models,
	}
}

func (c *ModelCommand) Name() string {
	return "model"
}

func (c *ModelCommand) Aliases() []string {
	return []string{"models"}
}

func (c *ModelCommand) Description() string {
	return "Switch between available AI models"
}

func (c *ModelCommand) Usage() string {
	return "/model - then select a model from the list"
}

func (c *ModelCommand) Execute(ctx *CommandContext) ExecutionResult {
	if len(c.availableModels) == 0 {
		fmt.Println("No models available")
		return REPLContinue
	}

	currentModel := ctx.State.GetModel()
	fmt.Println("Current model:")
	fmt.Printf("  %s\n", currentModel)

	ctx.Config.InfoColor().Printf("\nAvailable Models:\n\n")
	c.displayModels(ctx, currentModel)

	// Read selection without adding to command history
	selection, err := ReadInputWithoutHistory("Enter a number to select the model: ")
	if err != nil {
		return REPLExit
	}
	if selection == "" {
		fmt.Println("Model unchanged")
		return REPLContinue
	}

	selectedModel := c.parseSelection(selection)
	if selectedModel == "" {
		fmt.Println("Invalid selection. Please enter a number between 1 and", len(c.availableModels))
		return REPLContinue
	}

	c.switchModel(ctx, selectedModel)
	return REPLContinue
}

// displayModels shows all available models with highlighting for the current one
func (c *ModelCommand) displayModels(ctx *CommandContext, currentModel string) {
	highlightColor := ctx.Config.InfoColor()
	for i, model := range c.availableModels {
		if model == currentModel {
			highlightColor.Printf("  [%d] %s (current)\n", i+1, model)
		} else {
			fmt.Printf("  [%d] %s\n", i+1, model)
		}
	}
	fmt.Println()
}

// parseSelection converts user input to a model name
// Supports both numeric selection (1, 2, 3...) and direct model name
func (c *ModelCommand) parseSelection(selection string) string {
	// Try parsing as number first
	if num, err := strconv.Atoi(selection); err == nil {
		if num >= 1 && num <= len(c.availableModels) {
			return c.availableModels[num-1]
		}
		return ""
	}

	// Check if it matches a model name directly
	for _, model := range c.availableModels {
		if strings.EqualFold(model, selection) || strings.Contains(strings.ToLower(model), strings.ToLower(selection)) {
			return model
		}
	}

	return ""
}

// switchModel changes to a new model and clears history if different
func (c *ModelCommand) switchModel(ctx *CommandContext, newModel string) {
	if ctx.State.GetModel() == newModel {
		fmt.Printf("Model is already set to %s\n", newModel)
		return
	}

	ctx.State.SetModel(newModel)

	// TBD: Shoudl I save config to persist changes like model selection or system prompt update?

	// Clear conversation history for the new model to start fresh
	ctx.History.Clear()

	ctx.Config.InfoColor().Printf("✓ Switched to %s (conversation history cleared for this model)\n", newModel)
}
