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
	"github.com/chat-cli/chat-cli/errors"
	"github.com/mattn/go-isatty"
)

type StreamingOutputHandler func(ctx context.Context, part string) error

func ProcessStreamingOutput(output *bedrockruntime.ConverseStreamOutput, handler StreamingOutputHandler) (types.Message, error) {
	if output == nil {
		appErr := errors.NewValidationError(
			"invalid_input",
			"ProcessStreamingOutput received nil output",
			"Unable to process streaming response - invalid input received",
			nil,
		).WithOperation("ProcessStreamingOutput").WithComponent("utils")
		return types.Message{}, errors.Handle(appErr)
	}

	if handler == nil {
		appErr := errors.NewValidationError(
			"invalid_handler",
			"ProcessStreamingOutput received nil handler",
			"Unable to process streaming response - no output handler provided",
			nil,
		).WithOperation("ProcessStreamingOutput").WithComponent("utils")
		return types.Message{}, errors.Handle(appErr)
	}

	var combinedResult string
	msg := types.Message{}

	defer func() {
		if r := recover(); r != nil {
			appErr := errors.NewAWSError(
				"streaming_panic",
				fmt.Sprintf("Panic during streaming output processing: %v", r),
				"An unexpected error occurred while processing the streaming response. Please try again.",
				nil,
			).WithOperation("ProcessStreamingOutput").WithComponent("utils")
			errors.Handle(appErr)
		}
	}()

	for event := range output.GetStream().Events() {
		switch v := event.(type) {
		case *types.ConverseStreamOutputMemberMessageStart:
			msg.Role = v.Value.Role

		case *types.ConverseStreamOutputMemberContentBlockDelta:
			textResponse := v.Value.Delta.(*types.ContentBlockDeltaMemberText)
			if err := handler(context.Background(), textResponse.Value); err != nil {
				appErr := errors.NewAWSError(
					"handler_error",
					fmt.Sprintf("Streaming handler error: %v", err),
					"Error occurred while processing streaming response. The response may be incomplete.",
					err,
				).WithOperation("ProcessStreamingOutput").WithComponent("utils").WithRecoverable(true)
				return msg, errors.Handle(appErr)
			}
			combinedResult += textResponse.Value

		case *types.UnknownUnionMember:
			// Log unknown union members but continue processing
			appErr := errors.NewAWSError(
				"unknown_stream_event",
				fmt.Sprintf("Unknown streaming event type: %s", v.Tag),
				"Received unknown streaming event type - continuing with available content",
				nil,
			).WithOperation("ProcessStreamingOutput").WithComponent("utils").
				WithSeverity(errors.ErrorSeverityLow).WithRecoverable(true)
			errors.Handle(appErr)
		}
	}

	msg.Content = append(msg.Content,
		&types.ContentBlockMemberText{
			Value: combinedResult,
		},
	)

	return msg, nil
}

func ReadImage(filename string) (data []byte, imageType string, err error) {
	// Validate input
	if filename == "" {
		appErr := errors.NewValidationError(
			"empty_filename",
			"ReadImage received empty filename",
			"Please provide a valid image file path",
			nil,
		).WithOperation("ReadImage").WithComponent("utils")
		return nil, "", errors.Handle(appErr)
	}

	// Define a base directory for allowed images
	baseDir, err := os.Getwd()
	if err != nil {
		appErr := errors.NewFileSystemError(
			"working_directory_error",
			fmt.Sprintf("Unable to get working directory: %v", err),
			"Unable to determine current directory. Please check your file system permissions.",
			err,
		).WithOperation("ReadImage").WithComponent("utils")
		return nil, "", errors.Handle(appErr)
	}

	// Clean the filename and create the full path
	cleanFilename := filepath.Clean(filename)
	fullPath := filepath.Join(baseDir, cleanFilename)

	// Ensure the full path is within the base directory (security validation)
	relPath, err := filepath.Rel(baseDir, fullPath)
	if err != nil || strings.HasPrefix(relPath, "..") || strings.HasPrefix(relPath, string(filepath.Separator)) {
		appErr := errors.NewValidationError(
			"path_traversal_denied",
			fmt.Sprintf("Path traversal attempt detected: %s", filename),
			fmt.Sprintf("Access denied: '%s' is outside of the allowed directory. Please use a file within the current directory.", filename),
			nil,
		).WithOperation("ReadImage").WithComponent("utils").
			WithMetadata("attempted_path", filename).
			WithMetadata("resolved_path", fullPath)
		return nil, "", errors.Handle(appErr)
	}

	// Check if the file exists and get file info first
	fileInfo, statErr := os.Stat(fullPath)
	if os.IsNotExist(statErr) {
		appErr := errors.NewFileSystemError(
			"file_not_found",
			fmt.Sprintf("Image file does not exist: %s", filename),
			fmt.Sprintf("Image file '%s' not found. Please check the file path and try again.", filename),
			statErr,
		).WithOperation("ReadImage").WithComponent("utils").
			WithMetadata("filename", filename).
			WithMetadata("full_path", fullPath)
		return nil, "", errors.Handle(appErr)
	}
	if statErr != nil {
		appErr := errors.NewFileSystemError(
			"file_stat_error",
			fmt.Sprintf("Unable to access file information for %s: %v", filename, statErr),
			fmt.Sprintf("Unable to access file '%s'. Please check file permissions.", filename),
			statErr,
		).WithOperation("ReadImage").WithComponent("utils").
			WithMetadata("filename", filename)
		return nil, "", errors.Handle(appErr)
	}

	// Check if it's actually a file (not a directory)
	if fileInfo.IsDir() {
		appErr := errors.NewValidationError(
			"path_is_directory",
			fmt.Sprintf("Path is a directory, not a file: %s", filename),
			fmt.Sprintf("'%s' is a directory, not an image file. Please specify a valid image file.", filename),
			nil,
		).WithOperation("ReadImage").WithComponent("utils").
			WithMetadata("filename", filename)
		return nil, "", errors.Handle(appErr)
	}

	// Validate file extension after confirming it's a file
	ext := strings.ToLower(filepath.Ext(filename))
	if ext != "" {
		ext = ext[1:] // Remove the leading dot
	}

	var validImageType string
	switch ext {
	case "jpg":
		validImageType = "jpeg"
	case "jpeg":
		validImageType = "jpeg"
	case "png":
		validImageType = "png"
	case "gif":
		validImageType = "gif"
	case "webp":
		validImageType = "webp"
	default:
		appErr := errors.NewValidationError(
			"unsupported_image_format",
			fmt.Sprintf("Unsupported image format: %s", ext),
			fmt.Sprintf("Unsupported image format '%s'. Supported formats are: jpg, jpeg, png, gif, webp", ext),
			nil,
		).WithOperation("ReadImage").WithComponent("utils").
			WithMetadata("file_extension", ext).
			WithMetadata("filename", filename)
		return nil, "", errors.Handle(appErr)
	}

	// Check file size (reasonable limit for images)
	const maxImageSize = 50 * 1024 * 1024 // 50MB
	if fileInfo.Size() > maxImageSize {
		appErr := errors.NewValidationError(
			"file_too_large",
			fmt.Sprintf("Image file too large: %d bytes (max: %d)", fileInfo.Size(), maxImageSize),
			fmt.Sprintf("Image file '%s' is too large (%.1fMB). Maximum supported size is 50MB.", filename, float64(fileInfo.Size())/(1024*1024)),
			nil,
		).WithOperation("ReadImage").WithComponent("utils").
			WithMetadata("filename", filename).
			WithMetadata("file_size", fileInfo.Size()).
			WithMetadata("max_size", maxImageSize)
		return nil, "", errors.Handle(appErr)
	}

	// Read the file
	data, err = os.ReadFile(fullPath) // #nosec G304 - path is validated above
	if err != nil {
		appErr := errors.NewFileSystemError(
			"file_read_error",
			fmt.Sprintf("Unable to read image file %s: %v", filename, err),
			fmt.Sprintf("Unable to read image file '%s'. Please check file permissions and try again.", filename),
			err,
		).WithOperation("ReadImage").WithComponent("utils").
			WithMetadata("filename", filename).
			WithMetadata("full_path", fullPath)
		return nil, "", errors.Handle(appErr)
	}

	// Validate that we actually read some data
	if len(data) == 0 {
		appErr := errors.NewValidationError(
			"empty_image_file",
			fmt.Sprintf("Image file is empty: %s", filename),
			fmt.Sprintf("Image file '%s' is empty. Please provide a valid image file.", filename),
			nil,
		).WithOperation("ReadImage").WithComponent("utils").
			WithMetadata("filename", filename)
		return nil, "", errors.Handle(appErr)
	}

	return data, validImageType, nil
}

func StringPrompt(label string) string {
	// Validate input
	if label == "" {
		appErr := errors.NewValidationError(
			"empty_prompt_label",
			"StringPrompt received empty label",
			"Internal error: prompt label is missing",
			nil,
		).WithOperation("StringPrompt").WithComponent("utils").
			WithSeverity(errors.ErrorSeverityLow)
		errors.Handle(appErr)
		label = "Input" // Provide fallback label
	}

	// Check if we're in a TTY - if so, use the fancy bubble input
	if isatty.IsTerminal(os.Stdin.Fd()) || isatty.IsCygwinTerminal(os.Stdin.Fd()) {
		// We don't print the prompt here anymore since it's inside the input box
		input, success := BubbleInput()
		if !success {
			// Graceful degradation - fall back to simple input
			appErr := errors.NewFileSystemError(
				"bubble_input_failed",
				"Bubble input returned failure status",
				"Interactive input failed, falling back to simple input mode",
				nil,
			).WithOperation("StringPrompt").WithComponent("utils").
				WithSeverity(errors.ErrorSeverityLow).WithRecoverable(true)
			errors.Handle(appErr)
			
			// Fall back to simple input
			return fallbackStringPrompt(label)
		}
		return input
	}

	// Fallback to simple input for non-interactive use
	return fallbackStringPrompt(label)
}

// fallbackStringPrompt provides a simple input mechanism when bubble input fails
func fallbackStringPrompt(label string) string {
	var s string
	bufferSize := 8192

	r := bufio.NewReaderSize(os.Stdin, bufferSize)

	for {
		fmt.Fprint(os.Stderr, label+" ")
		var err error
		s, err = r.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				// Handle EOF gracefully
				break
			}
			appErr := errors.NewFileSystemError(
				"stdin_read_error",
				fmt.Sprintf("Failed to read from stdin: %v", err),
				"Unable to read input. Please try again.",
				err,
			).WithOperation("fallbackStringPrompt").WithComponent("utils").
				WithSeverity(errors.ErrorSeverityLow)
			errors.Handle(appErr)
			break
		}
		if s != "" {
			break
		}
	}

	return s
}

func DecodeImage(base64Image string) ([]byte, error) {
	// Validate input
	if base64Image == "" {
		appErr := errors.NewValidationError(
			"empty_base64_input",
			"DecodeImage received empty base64 string",
			"Please provide a valid base64-encoded image string",
			nil,
		).WithOperation("DecodeImage").WithComponent("utils")
		return nil, errors.Handle(appErr)
	}

	// Validate base64 string format (basic check)
	if len(base64Image)%4 != 0 {
		appErr := errors.NewValidationError(
			"invalid_base64_format",
			fmt.Sprintf("Invalid base64 format: length %d is not divisible by 4", len(base64Image)),
			"Invalid base64 image format. Please provide a properly encoded base64 string.",
			nil,
		).WithOperation("DecodeImage").WithComponent("utils").
			WithMetadata("input_length", len(base64Image))
		return nil, errors.Handle(appErr)
	}

	// Attempt to decode
	decoded, err := base64.StdEncoding.DecodeString(base64Image)
	if err != nil {
		appErr := errors.NewValidationError(
			"base64_decode_error",
			fmt.Sprintf("Failed to decode base64 string: %v", err),
			"Unable to decode base64 image data. Please check that the image data is properly encoded.",
			err,
		).WithOperation("DecodeImage").WithComponent("utils")
		return nil, errors.Handle(appErr)
	}

	// Validate that we got some data
	if len(decoded) == 0 {
		appErr := errors.NewValidationError(
			"empty_decoded_data",
			"Base64 decoding resulted in empty data",
			"Base64 image data decoded to empty content. Please provide valid image data.",
			nil,
		).WithOperation("DecodeImage").WithComponent("utils")
		return nil, errors.Handle(appErr)
	}

	// Basic validation - check for reasonable image size
	const maxDecodedSize = 100 * 1024 * 1024 // 100MB
	if len(decoded) > maxDecodedSize {
		appErr := errors.NewValidationError(
			"decoded_image_too_large",
			fmt.Sprintf("Decoded image too large: %d bytes (max: %d)", len(decoded), maxDecodedSize),
			fmt.Sprintf("Decoded image is too large (%.1fMB). Maximum supported size is 100MB.", float64(len(decoded))/(1024*1024)),
			nil,
		).WithOperation("DecodeImage").WithComponent("utils").
			WithMetadata("decoded_size", len(decoded)).
			WithMetadata("max_size", maxDecodedSize)
		return nil, errors.Handle(appErr)
	}

	return decoded, nil
}

func LoadDocument() (string, error) {
	var document string

	// Check if we're in a terminal (interactive mode)
	if isatty.IsTerminal(os.Stdin.Fd()) || isatty.IsCygwinTerminal(os.Stdin.Fd()) {
		// In interactive mode, no document to load from stdin
		return "", nil
	}

	// Read from stdin (non-interactive mode)
	stdin, err := io.ReadAll(os.Stdin)
	if err != nil {
		appErr := errors.NewFileSystemError(
			"stdin_read_error",
			fmt.Sprintf("Failed to read document from stdin: %v", err),
			"Unable to read document from input. Please check your input source and try again.",
			err,
		).WithOperation("LoadDocument").WithComponent("utils")
		return "", errors.Handle(appErr)
	}

	document = string(stdin)

	// Validate document size
	const maxDocumentSize = 10 * 1024 * 1024 // 10MB
	if len(document) > maxDocumentSize {
		appErr := errors.NewValidationError(
			"document_too_large",
			fmt.Sprintf("Document too large: %d bytes (max: %d)", len(document), maxDocumentSize),
			fmt.Sprintf("Input document is too large (%.1fMB). Maximum supported size is 10MB.", float64(len(document))/(1024*1024)),
			nil,
		).WithOperation("LoadDocument").WithComponent("utils").
			WithMetadata("document_size", len(document)).
			WithMetadata("max_size", maxDocumentSize)
		return "", errors.Handle(appErr)
	}

	// Wrap document in XML tags if content exists
	if document != "" {
		document = "<document>\n\n" + document + "\n\n</document>\n\n"
	}

	return document, nil
}
