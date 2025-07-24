# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Development Commands

### Building the Project
```bash
# Build the CLI binary
make

# Alternative: Direct Go build
go build -o ./bin/chat-cli main.go
```

### Running the Application
```bash
# Run from build directory
./bin/chat-cli <command> <args> <flags>

# Run directly with Go (for development)
go run main.go <command> <args> <flags>
```

### Testing
The project doesn't appear to have automated tests configured. Manual testing should be done by building and running the CLI commands.

## Architecture Overview

This is a Go CLI application built with Cobra that provides an interface to Amazon Bedrock LLMs. The architecture follows a clean separation of concerns:

### Core Components

**CLI Layer (`/cmd/`)**
- `root.go` - Base Cobra command that launches chat functionality directly when no subcommands are provided
- `chat.go` - Interactive chat sessions with persistent storage (called from root or as subcommand for management)
- `prompt.go` - One-shot prompt commands with stdin support
- `config.go` - Configuration management (set/get/unset model-id and custom-arn)
- `image.go` - Image generation commands
- `models*.go` - Model listing and management

**Configuration (`/config/`)**
- `config.go` - OS-specific file management using Viper
- Handles config.yaml storage in user's config directory
- Manages database path resolution

**Database Layer (`/db/` and `/repository/`)**
- SQLite-based persistence for chat history
- Repository pattern with base and chat-specific implementations
- Migration system for database schema management
- Chat sessions identified by UUID and stored with timestamps

**AWS Integration**
- Direct AWS SDK v2 integration for Bedrock services
- Supports both foundation models and custom ARNs
- Region configuration with us-east-1 as default
- Streaming and non-streaming response modes

### Key Data Flow

1. **Root command** (no args) → **Chat functionality** → Parse flags/config → **AWS SDK** → **Bedrock API**
2. **Subcommands** → Direct command execution (prompt, config, image, etc.)
3. **Chat sessions** → **Repository** → **SQLite DB** for persistence
4. **Configuration** precedence: CLI flags > config file > defaults

### Important Patterns

**Model Selection Priority:**
1. `--custom-arn` flag (highest)
2. `--model-id` flag
3. Config file values
4. Default: `anthropic.claude-3-5-sonnet-20240620-v1:0`

**Chat Session Management:**
- Auto-saves all chat interactions to SQLite
- `chat list` shows recent sessions with timestamps and previews
- `--chat-id` flag resumes specific conversations (works with root command)
- UUIDs track individual chat sessions
- Chat flags are available at root level for direct access

**Document Input:**
- Stdin piping supported: `cat file.go | chat-cli prompt "explain"`
- Wraps documents in `<document></document>` tags
- Image attachments via `--image` flag (PNG/JPG, <5MB)

### File Organization
- `main.go` - Entry point, delegates to cmd package
- `factory/` - Database factory pattern
- `utils/` - Utility functions for document loading
- `docs/` - Sphinx documentation with Python requirements

### AWS Requirements
- AWS CLI configured with credentials
- Bedrock model access enabled in AWS Console
- Region-specific model availability varies