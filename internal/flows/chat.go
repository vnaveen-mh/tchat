package flows

import (
	"context"
	"log/slog"
	"time"

	"tchat/internal/media"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"
)

// ChatFlow encapsulates the chat flow with its dependencies
type ChatFlow struct {
	genkit *genkit.Genkit
	flow   *core.Flow[ChatRequest, ChatResponse, struct{}]
}

// NewChatFlow creates a new chat flow with dependencies
func NewChatFlow(g *genkit.Genkit) *ChatFlow {
	cf := &ChatFlow{
		genkit: g,
	}

	// Define the flow
	cf.flow = genkit.DefineFlow(g, "chat-flow", cf.execute)

	return cf
}

// execute is the main flow execution function (without streaming)
func (cf *ChatFlow) execute(ctx context.Context, req ChatRequest) (ChatResponse, error) {
	return cf.generate(ctx, req, nil)
}

// generate handles the actual AI generation
func (cf *ChatFlow) generate(ctx context.Context, req ChatRequest, streamCallback StreamCallback) (ChatResponse, error) {
	response := ChatResponse{}
	startTime := time.Now()

	// Use model from request (required)
	model := req.Model

	// Track streaming metrics
	chunkCount := 0
	var firstChunkTime time.Time

	// Build message list: prior history + current user turn
	messages := make([]*ai.Message, 0, len(req.History)+1)
	messages = append(messages, req.History...)

	// Handle multimodal message if images are provided
	var currentMessage *ai.Message
	if len(req.ImagePaths) > 0 {
		// Load images
		images := make([]*media.ImageReference, 0, len(req.ImagePaths))
		for _, path := range req.ImagePaths {
			img, err := media.LoadImage(path)
			if err != nil {
				slog.Warn("Failed to load image", "path", path, "error", err)
				continue
			}
			images = append(images, img)
			slog.Info("Loaded image", "path", path, "size", len(img.Data))
		}

		if len(images) > 0 {
			// Build multimodal message with text and images
			currentMessage = media.BuildMultimodalMessage(req.UserInput, images)
			response.ImagesLoaded = len(images)
		} else {
			// Fallback to text-only if all images failed to load
			currentMessage = ai.NewUserTextMessage(req.UserInput)
		}
	} else {
		// Text-only message
		currentMessage = ai.NewUserTextMessage(req.UserInput)
	}

	messages = append(messages, currentMessage)

	// Build generation options
	opts := []ai.GenerateOption{
		ai.WithSystem(req.SystemPrompt),
		ai.WithModelName(model),
		ai.WithMessages(messages...),
	}

	// Add streaming handler if callback provided
	if streamCallback != nil {
		opts = append(opts, ai.WithStreaming(func(ctx context.Context, chunk *ai.ModelResponseChunk) error {
			if chunkCount == 0 {
				firstChunkTime = time.Now()
			}
			//slog.Info("chunk callback", "chunk id", chunkCount)
			chunkCount++
			return streamCallback(ctx, chunk)
		}))
	}

	// Generate response
	output, err := genkit.GenerateText(ctx, cf.genkit, opts...)

	duration := time.Since(startTime)
	response.DurationMs = duration.Milliseconds()
	response.Chunks = chunkCount

	if !firstChunkTime.IsZero() {
		response.TTFCMs = firstChunkTime.Sub(startTime).Milliseconds()
	}

	if err != nil {
		response.Error = err
		return response, err
	}

	response.Output = output
	return response, nil
}

// Run executes the flow with the given request (no streaming support due to serialization)
func (cf *ChatFlow) Run(ctx context.Context, req ChatRequest) (ChatResponse, error) {
	return cf.flow.Run(ctx, req)
}

// RunWithStreaming executes the generation with streaming support
// This bypasses the flow serialization to allow callbacks
func (cf *ChatFlow) RunWithStreaming(ctx context.Context, req ChatRequest, streamCallback StreamCallback) (ChatResponse, error) {
	return cf.generate(ctx, req, streamCallback)
}
