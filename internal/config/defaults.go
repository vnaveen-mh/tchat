package config

const (
	// DefaultSystemPrompt is the default system prompt for the AI
	DefaultSystemPrompt = "You are a helpful assistant"

	// DefaultConversationID is the default conversation identifier
	DefaultConversationID = "default"

	// DefaultMaxMessages is the default maximum number of messages to keep in history
	DefaultMaxMessages = 5

	// DefaultLogLevel is the default logging level
	DefaultLogLevel = "info"

	// DefaultLogToFile enables file logging
	DefaultLogToFile = true

	// DefaultLogMaxSizeMB is the max log file size in MB before rotation
	DefaultLogMaxSizeMB = 1

	// DefaultLogMaxBackups is the number of old log files to keep
	DefaultLogMaxBackups = 3

	// DefaultLogMaxAgeDays is the max days to keep old logs (0 = no age limit)
	DefaultLogMaxAgeDays = 30

	// Default colors for UI elements
	DefaultColorPrompt = "cyan"
	DefaultColorInfo   = "green"
	DefaultColorError  = "red"
	DefaultColorOutput = "green"
)
