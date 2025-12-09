package config

import (
	"github.com/fatih/color"
)

// PromptColor returns the color for prompt text (>)
func (c *Config) PromptColor() *color.Color {
	return c.promptColor
}

// OutputColor returns the color for AI output text
func (c *Config) OutputColor() *color.Color {
	return c.outputColor
}

// ErrorColor returns the color for error messages
func (c *Config) ErrorColor() *color.Color {
	return c.errorColor
}

// InfoColor returns the color for info messages
func (c *Config) InfoColor() *color.Color {
	return c.infoColor
}

// AsciiArtColor returns the color for ASCII art (always yellow/bold)
func (c *Config) AsciiArtColor() *color.Color {
	return c.asciiArtColor
}
