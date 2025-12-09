package appstate

import (
	"fmt"
	"sync"
)

type Option func(*State) error

type State struct {
	mu sync.RWMutex

	model        string
	systemPrompt string

	// What about History? should I keep it here?
}

func New(options ...Option) (*State, error) {
	state := &State{}
	for _, opt := range options {
		if err := opt(state); err != nil {
			return nil, err
		}
	}
	return state, nil
}

func WithModel(modelname string) Option {
	return func(s *State) error {
		if modelname == "" {
			return fmt.Errorf("model name cannot be empty")
		}
		s.model = modelname
		return nil
	}
}

func WithSystemPrompt(systemPrompt string) Option {
	return func(s *State) error {
		s.systemPrompt = systemPrompt
		return nil
	}
}

// GetModel returns current model set
func (s *State) GetModel() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.model
}

// GetSystemPrompt returns the current system prompt
func (s *State) GetSystemPrompt() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.systemPrompt
}

// SetModel sets/updates the model
func (s *State) SetModel(newModel string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.model = newModel
	// TBD - this change can optionally be peristed to user preferences
}

// SetSystemPrompt sets/updates system prompt
func (s *State) SetSystemPrompt(prompt string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.systemPrompt = prompt
	// TBD - this change can optionally be peristed to user preferences
}
