package ollama

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"time"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/ollama"
)

// ListModelsResponse represents the response from the List Local Models API (/api/tags)
type ListModelsResponse struct {
	Models []struct {
		Name       string    `json:"name"`
		ModifiedAt time.Time `json:"modified_at"`
		Size       int64     `json:"size"`
		Digest     string    `json:"digest"`
		Details    struct {
			Format        string   `json:"format"`
			Family        string   `json:"family"`
			Families      []string `json:"families"`
			ParameterSize string   `json:"parameter_size"`
		} `json:"details"`
	} `json:"models"`
}

// FetchModelDetailsRequest represents the request to the fetch Model Details API (/api/show)
type FetchModelDetailsRequest struct {
	Model   string `json:"model"`
	Verbose bool   `json:"verbose,omitempty"`
}

// FetchModelDetailsResponse represents the response from the fetch Model Details API (/api/show)
type FetchModelDetailsResponse struct {
	ModifiedAt   time.Time `json:"modified_at"`
	Template     string    `json:"template"`
	Parameters   string    `json:"parameters"`
	Capabilities []string  `json:"capabilities"`
	Details      struct {
		ParentModel       string   `json:"parent_model"`
		Format            string   `json:"format"`
		Family            string   `json:"family"`
		Families          []string `json:"families"`
		ParameterSize     string   `json:"parameter_size"`
		QuantizationLevel string   `json:"quantization_level"`
	} `json:"details"`
}

// ListModels lists available models from Ollama API endpoint /api/tags
func ListModels(serverAddress string) ([]string, error) {
	resp, err := http.Get(serverAddress + "/api/tags")
	if err != nil {
		return nil, fmt.Errorf("failed to list models: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama API returned status %d", resp.StatusCode)
	}

	var listResp ListModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	for _, m := range listResp.Models {
		slog.Info("ListModels", slog.Any(m.Name, m))
	}
	models := make([]string, 0, len(listResp.Models))
	for _, model := range listResp.Models {
		models = append(models, model.Name)
	}

	return models, nil
}

// FetchModelDetals fetches detailed information about a specific model
// ollama endpoint: /api/show
func FetchModelDetals(serverAddress, modelName string) (*FetchModelDetailsResponse, error) {
	reqBody := FetchModelDetailsRequest{
		Model:   modelName,
		Verbose: false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(serverAddress+"/api/show", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch model details: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama API returned status %d", resp.StatusCode)
	}

	var modelDetails FetchModelDetailsResponse
	if err := json.NewDecoder(resp.Body).Decode(&modelDetails); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	slog.Debug("Model Details", slog.Any(modelName, modelDetails))
	return &modelDetails, nil
}

// BuildModelOptions converts model capabilities to genkit's ai.ModelOptions
func BuildModelOptions(modelName string, capabilities []string) *ai.ModelOptions {
	modelOpts := &ai.ModelOptions{
		Label: modelName,
		Supports: &ai.ModelSupports{
			Multiturn:  true,
			SystemRole: true,
			Media:      slices.Contains(capabilities, "vision"),
			Tools:      slices.Contains(capabilities, "tools"),
		},
	}
	return modelOpts
}

// RegisterModels lists and registers all available Ollama models with Genkit
// Returns a list of model identifiers with "ollama/" prefix
func RegisterModels(g *genkit.Genkit, ollamaObj *ollama.Ollama, serverAddress string) ([]string, error) {
	slog.Info("Listing available Ollama models...")

	modelNames, err := ListModels(serverAddress)
	if err != nil {
		slog.Warn("Failed to list Ollama models", "error", err)
		return nil, err
	}
	slog.Info("Available Ollama models", "count", len(modelNames), "models", modelNames)

	// Define all Ollama models with Genkit
	registeredModels := make([]string, 0, len(modelNames))
	for i, modelName := range modelNames {
		// Fetch model capabilities by querying ollama endpoint /api/show
		fmt.Printf("  • Fetching details for %s (%d/%d)...\n", modelName, i+1, len(modelNames))
		var modelOpts *ai.ModelOptions
		modelDetails, err := FetchModelDetals(serverAddress, modelName)
		if err != nil {
			slog.Warn("Failed to fetch model capabilities, using defaults",
				"model", modelName,
				"error", err,
			)
		} else {
			slog.Info("Model capabilities",
				"model", modelName,
				"capabilities", modelDetails.Capabilities,
			)
			modelOpts = BuildModelOptions(modelName, modelDetails.Capabilities)
		}

		// Define model with options when available
		model := ollamaObj.DefineModel(g, ollama.ModelDefinition{
			Name: modelName,
			Type: "chat",
		}, modelOpts)

		slog.Info("Registered Ollama model", "name", model.Name())

		registeredModels = append(registeredModels, "ollama/"+modelName)
	}

	return registeredModels, nil
}
