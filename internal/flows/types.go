package flows

import (
	"context"

	"github.com/firebase/genkit/go/ai"
)

// StreamCallback is called for each chunk received during streaming
type StreamCallback func(ctx context.Context, chunk *ai.ModelResponseChunk) error

// ChatRequest represents the input to the chat flow (must be serializable)
type ChatRequest struct {
	UserInput    string
	Model        string
	SystemPrompt string
	History      []*ai.Message
	ImagePaths   []string // Optional image paths for vision models
}

// ChatResponse represents the output from the chat flow
type ChatResponse struct {
	Output       string
	DurationMs   int64
	TTFCMs       int64
	Chunks       int
	Error        error
	ImagesLoaded int // Number of images successfully loaded
}
