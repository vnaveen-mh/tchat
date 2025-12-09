package command

import "fmt"

// ConfigCommand displays current configuration
type ConfigCommand struct{}

func NewConfigCommand() *ConfigCommand {
	return &ConfigCommand{}
}

func (c *ConfigCommand) Name() string {
	return "config"
}

func (c *ConfigCommand) Aliases() []string {
	return []string{"settings", "cfg"}
}

func (c *ConfigCommand) Description() string {
	return "View current configuration"
}

func (c *ConfigCommand) Usage() string {
	return "/config - Display all configuration settings"
}

func (c *ConfigCommand) Execute(ctx *CommandContext) ExecutionResult {
	titleColor := ctx.Config.PromptColor()
	titleColor.Println("\n═══════════════════════════════════")
	titleColor.Println("        Configuration")
	titleColor.Println("═══════════════════════════════════")

	titleColor.Printf("\n📋 AI Settings:\n")
	fmt.Printf("  System Prompt    : %s\n", ctx.Config.GetSystemPrompt())
	fmt.Printf("  Current Model    : %s\n", ctx.Config.GetModel())

	titleColor.Printf("\n💾 Storage Settings:\n")
	fmt.Printf("  Config File      : %s\n", ctx.Config.ConfigPath())
	fmt.Printf("  App Directory    : %s\n", ctx.Config.GetAppDir())

	titleColor.Printf("\n📊 History Settings:\n")
	fmt.Printf("  Max Messages     : %d\n", ctx.Config.GetMaxMessages())
	fmt.Printf("  Current Messages : %d\n", ctx.History.Count())

	titleColor.Printf("\n📝 Logging Settings:\n")
	fmt.Printf("  Log Level        : %s\n", ctx.Config.GetLogLevel())

	titleColor.Printf("\n🎨 Color Settings:\n")
	fmt.Printf("  Prompt           : %s\n", ctx.Config.Colors().Prompt)
	fmt.Printf("  Info             : %s\n", ctx.Config.Colors().Info)
	fmt.Printf("  Error            : %s\n", ctx.Config.Colors().Error)
	fmt.Printf("  Output           : %s\n", ctx.Config.Colors().Output)

	titleColor.Println("\n═══════════════════════════════════")
	fmt.Println("\nUse /system to change system prompt")
	fmt.Println("Use /model to switch models")
	fmt.Println()

	return REPLContinue
}
