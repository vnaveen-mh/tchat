package history

import (
	"fmt"
	"sync"
	"time"

	"github.com/firebase/genkit/go/ai"
)

// config contains configuration for history management
type config struct {
	maxMessages int
}

type Option func(*config)

func WithMaxMessages(n int) Option {
	return func(cfg *config) {
		cfg.maxMessages = n
	}
}

// Manager manages conversation history
type HistoryManager struct {
	config   config
	messages []*ai.Message
	mu       sync.RWMutex
}

// New creates a new history manager
func NewHistoryManager(opts ...Option) *HistoryManager {
	cfg := config{
		maxMessages: 5,
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	if cfg.maxMessages < 0 {
		cfg.maxMessages = 5
	}
	return &HistoryManager{
		config:   cfg,
		messages: []*ai.Message{},
	}
}

// Add adds a message to history and enforces limits
func (h *HistoryManager) Add(model string, msg *ai.Message) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.messages = append(h.messages, msg)
	h.enforceLimits()
}

// AddUserMessage is a convenience method to add a user message
func (h *HistoryManager) AddUserMessage(model, text string) {
	h.Add(model, ai.NewUserTextMessage(text))
}

// AddAssistantMessage is a convenience method to add an assistant message
func (h *HistoryManager) AddAssistantMessage(model, text string) {
	h.Add(model, ai.NewModelTextMessage(text))
}

// GetAll returns all messages for a model
func (h *HistoryManager) GetAll() []*ai.Message {
	h.mu.RLock()
	defer h.mu.RUnlock()

	result := make([]*ai.Message, len(h.messages))
	copy(result, h.messages)
	return result
}

// Set sets history to messages
func (h *HistoryManager) Set(msgs []*ai.Message) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(msgs) == 0 {
		h.messages = []*ai.Message{}
		return
	}

	h.messages = make([]*ai.Message, len(msgs))
	copy(h.messages, msgs)
}

// Clear removes all messages
func (h *HistoryManager) Clear() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.messages = []*ai.Message{}
}

// Count returns the number of messages
func (h *HistoryManager) Count() int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return len(h.messages)
}

// IsEmpty returns true if history is empty
func (h *HistoryManager) IsEmpty() bool {
	return h.Count() == 0
}

// GetLast returns the last N messages
func (h *HistoryManager) GetLast(model string, n int) []*ai.Message {
	h.mu.RLock()
	defer h.mu.RUnlock()

	msgs := h.messages
	total := len(msgs)
	if n <= 0 || n >= total {
		result := make([]*ai.Message, total)
		copy(result, msgs)
		return result
	}

	start := total - n
	result := make([]*ai.Message, n)
	copy(result, msgs[start:])
	return result
}

// enforceLimits removes old messages if limits are exceeded
// the caller must have already locked the mutex
func (h *HistoryManager) enforceLimits() {
	// mutex must be locked by the caller of this function
	msgs := h.messages
	if h.config.maxMessages > 0 && len(msgs) > h.config.maxMessages {
		// Keep only the last MaxMessages
		msgs = msgs[len(msgs)-h.config.maxMessages:]
	}
	h.messages = msgs
}

// Statistics contains history statistics
type Statistics struct {
	TotalMessages     int
	UserMessages      int
	AssistantMessages int
	OldestTimestamp   *time.Time
	NewestTimestamp   *time.Time
}

// GetStats returns statistics about the history
func (h *HistoryManager) GetStats() Statistics {
	h.mu.RLock()
	defer h.mu.RUnlock()

	msgs := h.messages
	stats := Statistics{
		TotalMessages: len(msgs),
	}

	for _, msg := range msgs {
		switch msg.Role {
		case "user":
			stats.UserMessages++
		case "model":
			stats.AssistantMessages++
		}
	}

	return stats
}

// String returns a human-readable representation
func (h *HistoryManager) String() string {
	stats := h.GetStats()
	return fmt.Sprintf("History: %d messages (%d user, %d assistant)",
		stats.TotalMessages, stats.UserMessages, stats.AssistantMessages)
}
