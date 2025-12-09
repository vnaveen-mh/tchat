package logging

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	logger  *slog.Logger
	once    sync.Once
	logFile *os.File
	logDir  string
)

// Config holds logging configuration
type Config struct {
	LogDir string
	Level  string
}

// Init initializes the logging system with rotation support
// If the initialization fails, fallback to os.Stderr
func Init(version string, cfg Config) error {
	var initErr error

	once.Do(func() {
		logDir = cfg.LogDir

		// Create logs directory if it doesn't exist
		if err := os.MkdirAll(logDir, 0755); err != nil {
			initErr = fmt.Errorf("failed to create log directory: %w", err)
			return
		}

		// Open current log file
		logPath := filepath.Join(logDir, "tchat.log")
		var err error
		logFile, err = os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			initErr = fmt.Errorf("failed to open log file: %w", err)
			return
		}

		// Parse log level
		level := slog.LevelInfo
		switch strings.ToLower(cfg.Level) {
		case "debug":
			level = slog.LevelDebug
		case "info":
			level = slog.LevelInfo
		case "warn", "warning":
			level = slog.LevelWarn
		case "error":
			level = slog.LevelError
		}

		// Create logger with file output
		handler := slog.NewTextHandler(logFile, &slog.HandlerOptions{
			Level: level,
		})
		logger = slog.New(handler)

		logger.Info("Starting application", "version", version)
	})

	if initErr != nil {
		// Set default logger to write to stderr as fallback
		stderrHandler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
		slog.SetDefault(slog.New(stderrHandler))
		return initErr
	}
	slog.SetDefault(logger)
	return nil
}

// Close closes the log file
func Close() {
	if logFile != nil {
		logFile.Close()
	}
}

// Writer returns an io.Writer for the log file
func Writer() io.Writer {
	if logFile != nil {
		return logFile
	}
	return os.Stderr
}
