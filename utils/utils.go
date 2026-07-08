package utils

import (
	"bufio"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/mattn/go-isatty"
)

type StreamingOutputHandler func(ctx context.Context, part string) error

// ProcessStreamingOutput drains a Bedrock ConverseStream, invoking handler
// for each text delta and reasoningHandler for each reasoning-content delta
// (pass a no-op handler if the caller doesn't support reasoning mode).
func ProcessStreamingOutput(output *bedrockruntime.ConverseStreamOutput, handler, reasoningHandler StreamingOutputHandler) (types.Message, error) {

	var combinedResult string

	msg := types.Message{}

	for event := range output.GetStream().Events() {
		switch v := event.(type) {
		case *types.ConverseStreamOutputMemberMessageStart:

			msg.Role = v.Value.Role

		case *types.ConverseStreamOutputMemberContentBlockDelta:

			switch delta := v.Value.Delta.(type) {
			case *types.ContentBlockDeltaMemberText:
				if err := handler(context.Background(), delta.Value); err != nil {
					return msg, fmt.Errorf("handler error: %w", err)
				}
				combinedResult += delta.Value

			case *types.ContentBlockDeltaMemberReasoningContent:
				if textDelta, ok := delta.Value.(*types.ReasoningContentBlockDeltaMemberText); ok {
					if err := reasoningHandler(context.Background(), textDelta.Value); err != nil {
						return msg, fmt.Errorf("handler error: %w", err)
					}
				}
				// Signature and redacted-content deltas aren't rendered as
				// visible text; prompt is one-shot so there's no next turn
				// to preserve them for (Functional Design Decision 3,
				// unit-5-extended-thinking).
			}

		case *types.UnknownUnionMember:
			fmt.Println("unknown tag:", v.Tag)
		}
	}

	msg.Content = append(msg.Content,
		&types.ContentBlockMemberText{
			Value: combinedResult,
		},
	)

	return msg, nil
}

// ValidateLocalPath confines filename resolution to the current working
// directory, returning the validated absolute path or an error if filename
// escapes it or doesn't exist. Shared by ReadImage, ReadDocument, and the
// read_file tool so path-traversal protection lives in exactly one place.
func ValidateLocalPath(filename string) (string, error) {
	baseDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("unable to get working directory: %w", err)
	}

	// Clean the filename and create the full path
	cleanFilename := filepath.Clean(filename)
	fullPath := filepath.Join(baseDir, cleanFilename)

	// Ensure the full path is within the base directory
	relPath, err := filepath.Rel(baseDir, fullPath)
	if err != nil || strings.HasPrefix(relPath, "..") || strings.HasPrefix(relPath, string(filepath.Separator)) {
		return "", fmt.Errorf("access denied: %s is outside of the allowed directory", filename)
	}

	// Check if the file exists
	if _, statErr := os.Stat(fullPath); os.IsNotExist(statErr) {
		return "", fmt.Errorf("file does not exist: %s", filename)
	}

	return fullPath, nil
}

func ReadImage(filename string) (data []byte, imageType string, err error) {

	fullPath, err := ValidateLocalPath(filename)
	if err != nil {
		return nil, "", err
	}

	// Read the file
	data, err = os.ReadFile(fullPath) // #nosec G304 - path is validated above
	if err != nil {
		return nil, "", fmt.Errorf("unable to read file: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(filename))
	if ext != "" {
		ext = ext[1:] // Remove the leading dot
	}

	// imageType is already declared as named return parameter

	switch ext {
	case "jpg":
		imageType = "jpeg"
	case "jpeg":
		imageType = "jpeg"
	case "png":
		imageType = "png"
	case "gif":
		imageType = "gif"
	case "webp":
		imageType = "webp"
	default:
		return nil, "", fmt.Errorf("unsupported file type")

	}

	return data, imageType, nil
}

// ReadDocument reads a local document file for use as a Bedrock document
// content block, mirroring ReadImage's shape. Supported formats match
// Bedrock's DocumentFormat: pdf, csv, doc, docx, xls, xlsx, html, txt, md.
func ReadDocument(filename string) (data []byte, format string, err error) {
	fullPath, err := ValidateLocalPath(filename)
	if err != nil {
		return nil, "", err
	}

	data, err = os.ReadFile(fullPath) // #nosec G304 - path is validated above
	if err != nil {
		return nil, "", fmt.Errorf("unable to read file: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(filename))
	if ext != "" {
		ext = ext[1:] // Remove the leading dot
	}

	switch ext {
	case "pdf", "csv", "doc", "docx", "xls", "xlsx", "html", "txt", "md":
		format = ext
	default:
		return nil, "", fmt.Errorf("unsupported document type: %s", ext)
	}

	return data, format, nil
}

func StringPrompt(label string) string {
	// Check if we're in a TTY - if so, use the fancy bubble input
	if isatty.IsTerminal(os.Stdin.Fd()) || isatty.IsCygwinTerminal(os.Stdin.Fd()) {
		// We don't print the prompt here anymore since it's inside the input box
		input, _ := BubbleInput()
		return input
	}

	// Fallback to simple input for non-interactive use
	var s string
	bufferSize := 8192

	r := bufio.NewReaderSize(os.Stdin, bufferSize)

	for {
		fmt.Fprint(os.Stderr, label+" ")
		s, _ = r.ReadString('\n')
		if s != "" {
			break
		}
	}

	return s
}

func DecodeImage(base64Image string) ([]byte, error) {
	decoded, err := base64.StdEncoding.DecodeString(base64Image)
	if err != nil {
		return nil, err
	}
	return decoded, nil
}

func LoadDocument() (string, error) {

	// read a document from stdin
	var document string

	if isatty.IsTerminal(os.Stdin.Fd()) || isatty.IsCygwinTerminal(os.Stdin.Fd()) {
		// do nothing
	} else {
		stdin, err := io.ReadAll(os.Stdin)

		if err != nil {
			return "", err
		}
		document = string(stdin)
	}

	if document != "" {
		document = "<document>\n\n" + document + "\n\n</document>\n\n"
	}

	return document, nil
}
