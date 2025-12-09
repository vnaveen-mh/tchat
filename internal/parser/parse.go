package parser

import (
	"os"
	"path/filepath"
	"strings"
)

type ParsedInput struct {
	Mode       string
	Text       string
	ImagePaths []string
	Raw        string
	Command    string
	Args       []string
}

func ParseLine(line string) ParsedInput {
	trimmed := strings.TrimSpace(line)
	out := ParsedInput{Raw: line}

	if trimmed == "" {
		return out
	}

	tokens := strings.Fields(trimmed)

	var textTokens []string
	var imagePaths []string

	for _, tok := range tokens {
		if isImageFile(tok) {
			imagePaths = append(imagePaths, tok)
		} else {
			textTokens = append(textTokens, tok)
		}
	}

	out.ImagePaths = imagePaths
	if len(imagePaths) > 0 {
		out.Mode = "vision"
	} else {
		out.Mode = "chat"
	}

	text := strings.TrimSpace(strings.Join(textTokens, " "))
	if text == "" && len(imagePaths) > 0 {
		text = "Describe this image in detail."
	}
	out.Text = text

	return out
}

func isImageFile(path string) bool {
	// remove trailing punctuation like ",", "?" etc
	path = strings.TrimRight(path, ".,!?;:")
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return false
	}
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".png", ".jpg", ".jpeg", ".webp", ".gif", ".bmp":
		return true
	}
	return false
}
