package media

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/firebase/genkit/go/ai"
)

// SupportedImageFormats maps file extensions to MIME types
var SupportedImageFormats = map[string]string{
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".gif":  "image/gif",
	".webp": "image/webp",
	".bmp":  "image/bmp",
}

// ImageReference represents a reference to an image in user input
type ImageReference struct {
	Path     string // Original path/URL from user input
	MimeType string
	Data     []byte // Base64 encoded data
}

// ExtractImagePaths extracts potential image file paths from user input
// Looks for patterns like: /path/to/image.jpg, ~/image.png, ./image.jpg, image://path
func ExtractImagePaths(input string) []string {
	var paths []string
	words := strings.Fields(input)

	for _, word := range words {
		// Check if it has a supported image extension
		ext := strings.ToLower(filepath.Ext(word))
		if _, supported := SupportedImageFormats[ext]; supported {
			paths = append(paths, word)
		}
	}

	return paths
}

// LoadImage loads an image from a local file path or URL
func LoadImage(path string) (*ImageReference, error) {
	// Expand home directory
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to expand home directory: %w", err)
		}
		path = filepath.Join(home, path[2:])
	}

	// Check if it's a URL
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return loadImageFromURL(path)
	}

	// Load from local file
	return loadImageFromFile(path)
}

// loadImageFromFile loads an image from the local filesystem
func loadImageFromFile(path string) (*ImageReference, error) {
	// Get MIME type from extension
	ext := strings.ToLower(filepath.Ext(path))
	mimeType, ok := SupportedImageFormats[ext]
	if !ok {
		return nil, fmt.Errorf("unsupported image format: %s", ext)
	}

	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read image file: %w", err)
	}

	return &ImageReference{
		Path:     path,
		MimeType: mimeType,
		Data:     data,
	}, nil
}

// loadImageFromURL downloads an image from a URL
func loadImageFromURL(url string) (*ImageReference, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download image: status %d", resp.StatusCode)
	}

	// Read response body
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read image data: %w", err)
	}

	// Try to get MIME type from Content-Type header
	mimeType := resp.Header.Get("Content-Type")
	if mimeType == "" {
		// Fallback to extension-based detection
		ext := strings.ToLower(filepath.Ext(url))
		var ok bool
		mimeType, ok = SupportedImageFormats[ext]
		if !ok {
			mimeType = "image/jpeg" // Default fallback
		}
	}

	return &ImageReference{
		Path:     url,
		MimeType: mimeType,
		Data:     data,
	}, nil
}

// ToBase64 converts image data to base64 string
func (img *ImageReference) ToBase64() string {
	return base64.StdEncoding.EncodeToString(img.Data)
}

// ToDataURI converts image to a data URI (data:image/jpeg;base64,...)
func (img *ImageReference) ToDataURI() string {
	return fmt.Sprintf("data:%s;base64,%s", img.MimeType, img.ToBase64())
}

// ToMediaPart converts the image to a Genkit media part
func (img *ImageReference) ToMediaPart() *ai.Part {
	return ai.NewMediaPart(img.MimeType, img.ToDataURI())
}

// BuildMultimodalMessage creates a message with text and images
func BuildMultimodalMessage(text string, images []*ImageReference) *ai.Message {
	parts := make([]*ai.Part, 0, len(images)+1)

	// Add text part first
	parts = append(parts, ai.NewTextPart(text))

	// Add image parts
	for _, img := range images {
		parts = append(parts, img.ToMediaPart())
	}

	return ai.NewUserMessage(parts...)
}
