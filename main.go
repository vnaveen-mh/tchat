package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"syscall"
	"time"

	"tchat/internal/appstate"
	"tchat/internal/command"
	"tchat/internal/config"
	"tchat/internal/db"
	"tchat/internal/flows"
	"tchat/internal/history"
	"tchat/internal/logging"
	"tchat/internal/media"
	ollamahelper "tchat/internal/ollama"
	"tchat/internal/utils"
	"tchat/internal/version"

	"github.com/chzyer/readline"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/ollama"
	"github.com/google/uuid"
	"golang.design/x/clipboard"
)

// lastResponse stores the last AI response for clipboard copy
var lastResponse = ""

func main() {
	fmt.Printf("Initializing...\n")

	// Initialize configuration
	fmt.Printf("• Loading configuration...\n")
	cfg, err := config.New()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("  ✓ Configuration loaded\n")

	// Initialize logging with rotation
	logsDir := filepath.Join(cfg.GetAppDir(), "logs")
	if err := logging.Init(version.Version, logging.Config{
		LogDir: logsDir,
		Level:  cfg.GetLogLevel(),
	}); err != nil {
		fmt.Printf("  ⚠ Warning: Logging initialization failed: %v\n", err)
	}
	defer logging.Close()

	slog.Info("App directory", "path", cfg.GetAppDir())
	slog.Info("Logs directory", "path", logsDir)

	// Initialize storage
	fmt.Printf("• Initializing database...\n")
	dbPath := filepath.Join(cfg.GetAppDir(), "tchat.db")
	store, err := db.New(dbPath)
	if err != nil {
		slog.Error("Failed to initialize storage", "error", err)
		fmt.Printf("  ⚠ Warning: Database storage disabled: %v\n", err)
		store = nil
	} else {
		defer store.Close()
		slog.Info("Storage initialized", "database", dbPath)
		cfg.InfoColor().Printf("  ✓ Database ready at %s\n", cfg.GetAppDir())
	}

	// Initialize clipboard
	err = clipboard.Init()
	if err != nil {
		slog.Warn("Clipboard initialization failed", "error", err)
		cfg.ErrorColor().Println("Warning: Clipboard functionality may not work properly")
	}

	ctx := context.Background()

	ollamaHost := os.Getenv("OLLAMA_HOST")
	if ollamaHost == "" {
		ollamaHost = "http://localhost:11434"
	}
	ollamaObj := &ollama.Ollama{
		ServerAddress: ollamaHost,
		Timeout:       300, // 5 minutes
	}

	// Initialize Genkit with the Google AI plugin
	fmt.Printf("• Initializing Genkit...\n")
	g := genkit.Init(ctx,
		genkit.WithPlugins(ollamaObj),
	)
	cfg.InfoColor().Printf("  ✓ Genkit ready\n")

	// Register Ollama models
	fmt.Printf("• Discovering Ollama models...\n")
	availableModels, err := ollamahelper.RegisterModels(g, ollamaObj, ollamaHost)
	if err != nil {
		slog.Error("Failed to register Ollama models", "error", err)
		cfg.ErrorColor().Printf("  x Failed to register ollama models")
		os.Exit(1)
	}

	if len(availableModels) == 0 {
		slog.Warn("No local ollama models are available",
			"message", "Run `ollama pull <model_name> to pull a model. Visit https://ollama.com for more details")
		cfg.ErrorColor().Printf("  x No local ollama models available\n")
		os.Exit(1)
	}
	cfg.InfoColor().Printf("  ✓ Found %d models\n", len(availableModels))

	// set model from config. If not available, choose one from available ollama models
	currentModel := cfg.GetModel()
	if currentModel == "" || !slices.Contains(availableModels, currentModel) {
		currentModel = availableModels[0] // Use first Ollama model by default
	}

	cfg.InfoColor().Printf("  ✓ Using model: %s\n", currentModel)

	// Initialize history manager
	historyMgr := history.NewHistoryManager(history.WithMaxMessages(5))

	// Load history from DB
	fmt.Printf("• Loading conversation history...\n")
	fmt.Printf("  ✓ Initializing history manager\n")
	if store != nil {
		fmt.Printf("  ✓ Attempting to store history from database\n")
		msgs, err := store.LoadHistory(context.Background())
		if err != nil {
			slog.Warn("Failed to load history from database", "error", err)
			fmt.Printf("  ⚠ Failed to load messages from database: %s\n", err)
		} else {
			historyMgr.Set(msgs)
			count := historyMgr.Count()
			if count > 0 {
				slog.Info("History loaded from database", "messages", count)
				fmt.Printf("  ✓ Loaded %d messages\n", count)
			} else {
				slog.Info("History is empty")
				fmt.Printf("  ✓ History is empty. Starting fresh conversation\n")
			}
		}
	} else {
		fmt.Printf("  ⚠ Database unavailable, starting fresh\n")
	}
	cfg.InfoColor().Printf("  ✓ History manager ready\n")

	// Initialize app state
	fmt.Printf("• Initializing app state...\n")
	state, err := appstate.New(
		appstate.WithModel(currentModel),
		appstate.WithSystemPrompt(cfg.GetSystemPrompt()),
	)
	if err != nil {
		slog.Error("App state creation faile", "error", err)
		cfg.ErrorColor().Printf("  ⚠ App state creation failed\n")
		os.Exit(1)
	}
	fmt.Printf("  ✓ App state created and initialized\n")

	// Initialize command registry
	cmdRegistry := command.InitializeRegistry(availableModels, store)

	// Initialize chat flow with dependencies
	chatFlow := flows.NewChatFlow(g)

	// Setup readline with history
	historyFile := filepath.Join(cfg.GetAppDir(), "history")

	rl, err := readline.NewEx(&readline.Config{
		Prompt:          cfg.PromptColor().Sprint("tchat> "),
		HistoryFile:     historyFile,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})
	if err != nil {
		slog.Error("Failed to initialize readline", "error", err)
		fmt.Printf("Error: Could not initialize input handler: %v\n", err)
		return
	}
	defer rl.Close()

	// Setup signal handling for Ctrl-C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGINT)

	// Track generation state for cancellation
	var mu sync.Mutex
	var genCancel context.CancelFunc

	// Handle Ctrl-C in background
	go func() {
		for range sigChan {
			mu.Lock()
			if genCancel != nil {
				// Cancel ongoing generation
				genCancel()
				fmt.Println() // New line after ^C
				cfg.ErrorColor().Println("Generation canceled by user")
			}
			mu.Unlock()
		}
	}()

	// Show ready message
	fmt.Println()
	fmt.Printf("\nReady! Type /help for available commands\n")
	fmt.Printf("Use ↑/↓ arrow keys to navigate command history\n\n")

	sessionId := uuid.NewString()
	session := db.Session{
		SessionId: sessionId,
		ModelName: state.GetModel(),
	}
	store.CreateSession(session)

	// Print asciiart and welcome message
	cfg.AsciiArtColor().Println(utils.AsciiArt)
	fmt.Printf("TChat - Your Terminal Chat AI Assistant\n")

	// Main read loop
	for {
		line, err := rl.Readline()
		if err != nil {
			if err == readline.ErrInterrupt {
				fmt.Println("Press Ctrl-C to cancel an operation in progress or press Ctrl-D to exit")
				continue
			} else if err == io.EOF {
				// Ctrl-D pressed
				break
			}
			slog.Error("Readline error", "error", err)
			break
		}

		userInput := strings.TrimSpace(line)
		if userInput == "" {
			continue
		}
		// Special commands
		if cmdRegistry.IsCommand(userInput) {
			cmd, _ := cmdRegistry.Get(userInput)
			cmdCtx := &command.CommandContext{
				Ctx:          ctx,
				Config:       cfg,
				State:        state,
				Readline:     rl,
				History:      historyMgr,
				LastResponse: &lastResponse,
			}
			result := cmd.Execute(cmdCtx)
			if result == command.REPLExit {
				break
			}
			continue
		}

		// Check for unrecognized commands (anything starting with /)
		if strings.HasPrefix(userInput, "/") {
			cfg.ErrorColor().Printf("Unknown command: %s\n", userInput)
			fmt.Println("Type /help to see available commands")
			continue
		}

		// Detect images in user input
		imagePaths := media.ExtractImagePaths(userInput)
		if len(imagePaths) > 0 {
			cfg.InfoColor().Printf("📷 Detected %d image(s): %v\n", len(imagePaths), imagePaths)
		}

		// Create cancellable context for this generation
		mu.Lock()
		genCtx, cancel := context.WithCancel(ctx)
		genCancel = cancel
		mu.Unlock()

		cleanup := func() {
			cancel()
			mu.Lock()
			genCancel = nil
			mu.Unlock()
		}

		// Log generation start
		startTime := time.Now()
		slog.Info("Generation started",
			"model", state.GetModel(),
			"input_length", len(userInput),
			"history_messages", historyMgr.Count(),
			"images", len(imagePaths),
			"input", userInput,
		)

		// Prepare streaming callback
		firstChunk := true
		streamCallback := flows.StreamCallback(func(ctx context.Context, chunk *ai.ModelResponseChunk) error {
			if firstChunk {
				fmt.Println()
				firstChunk = false
			}
			cfg.OutputColor().Printf("%s", chunk.Text())
			return nil
		})

		// Execute chat flow with streaming
		resp, err := chatFlow.RunWithStreaming(genCtx, flows.ChatRequest{
			UserInput:    userInput,
			Model:        state.GetModel(),
			SystemPrompt: state.GetSystemPrompt(),
			History:      historyMgr.GetAll(),
			ImagePaths:   imagePaths,
		}, streamCallback)

		if err != nil {
			// Check if it was cancelled
			if genCtx.Err() == context.Canceled {
				slog.Info("Generation cancelled by user",
					"model", state.GetModel(),
					"duration_ms", resp.DurationMs,
				)
				cleanup()
				continue
			}

			// some other error
			slog.Error("Generation failed",
				"error", err,
				"duration_ms", resp.DurationMs,
				"model", state.GetModel(),
			)
			cfg.ErrorColor().Printf("Error generating response: %v\n", err)
			cleanup()
			continue
		}

		cfg.OutputColor().Println()

		// Show images loaded if any
		if resp.ImagesLoaded > 0 {
			cfg.InfoColor().Printf("✓ Processed %d image(s)\n", resp.ImagesLoaded)
		}

		// Log generation success with metadata
		slog.Info("Generation completed",
			"model", state.GetModel(),
			"duration_ms", resp.DurationMs,
			"ttfc_ms", resp.TTFCMs,
			"chunks", resp.Chunks,
			"output_length", len(resp.Output),
			"input_length", len(userInput),
			"images_loaded", resp.ImagesLoaded,
		)

		// Save to database
		if store != nil {
			turn := db.ConversationTurn{
				SessionId:    sessionId,
				Timestamp:    startTime,
				UserInput:    userInput,
				ModelOutput:  resp.Output,
				DurationMs:   resp.DurationMs,
				TTFCMs:       resp.TTFCMs,
				Chunks:       resp.Chunks,
				InputLength:  len(userInput),
				OutputLength: len(resp.Output),
			}
			if id, err := store.SaveTurn(turn); err != nil {
				slog.Error("Failed to save conversation to database", "error", err)
			} else {
				slog.Debug("Conversation saved", "id", id)
			}
		}

		// Update conversation history
		historyMgr.AddUserMessage(state.GetModel(), userInput)
		historyMgr.AddAssistantMessage(state.GetModel(), resp.Output)

		// Store last response for clipboard copy
		lastResponse = resp.Output

		// Cleanup generation state
		cleanup()
	}

	// Save history on exit
	if store != nil {
		slog.Debug("Saving history on exit")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		msgs := historyMgr.GetAll()
		if err := store.SaveHistory(shutdownCtx, msgs); err != nil {
			slog.Error("Failed to save history on exit", "error", err)
		} else {
			slog.Info("History saved successfully")
		}
	}

	fmt.Println("Goodbye!")
}
