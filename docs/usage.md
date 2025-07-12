# Usage

(config)=
## Config

Chat-CLI provides a configuration system that allows you to set persistent default values for commonly used settings. This eliminates the need to specify the same flags repeatedly when using the `chat` and `prompt` commands.

### Managing Configuration

#### Setting Values

Use the `config set` command to store default values:

```shell
# Set a default model ID
chat-cli config set model-id "anthropic.claude-3-5-sonnet-20240620-v1:0"

# Set a custom ARN for marketplace or cross-region models  
chat-cli config set custom-arn "arn:aws:bedrock:us-west-2::foundation-model/custom-model"
```

#### Viewing Configuration

List all current configuration values:

```shell
chat-cli config list
```

Example output:
```
Current configuration:
  model-id = anthropic.claude-3-5-sonnet-20240620-v1:0
  custom-arn = arn:aws:bedrock:us-west-2::foundation-model/custom-model
```

#### Removing Values

Remove specific configuration values when no longer needed:

```shell
chat-cli config unset model-id
chat-cli config unset custom-arn
```

### Configuration Precedence

The configuration system uses a clear precedence hierarchy to determine which values to use:

1. **Command line flags** (highest priority)
   - Values specified with `--model-id` or `--custom-arn` flags
   - Always override configuration file and defaults

2. **Configuration file** (medium priority)
   - Values set using `chat-cli config set`
   - Used when no command line flag is provided

3. **Built-in defaults** (lowest priority)
   - Default model: `anthropic.claude-3-5-sonnet-20240620-v1:0`
   - Used when no configuration or flags are set

### Custom ARN Priority

When both `model-id` and `custom-arn` are configured, `custom-arn` takes precedence. This design allows you to:

- Set a default `model-id` for regular use
- Override with `custom-arn` for marketplace or cross-region models
- Use command line flags to override either setting temporarily

### Supported Settings

| Setting | Description | Example |
|---------|-------------|---------|
| `model-id` | Default model identifier for Bedrock foundation models | `anthropic.claude-3-5-sonnet-20240620-v1:0` |
| `custom-arn` | Custom ARN for marketplace or cross-region inference | `arn:aws:bedrock:us-west-2::foundation-model/custom-model` |

### Configuration Storage

Configuration values are stored in a YAML file in your system's standard configuration directory:

- **macOS**: `~/Library/Application Support/chat-cli/config.yaml`
- **Linux**: `~/.config/chat-cli/config.yaml` 
- **Windows**: `%APPDATA%\chat-cli\config.yaml`

(prompt)=
## Prompt

(chat)=
## Chat

(image)=
## Image

(agentic)=
## Agentic Operations

Chat-CLI includes autonomous agents that can perform complex file operations using natural language instructions. These agents understand tasks described in plain English and can autonomously plan and execute multi-step operations.

**Model Configuration:** Agentic commands use `anthropic.claude-3-sonnet-20240229-v1:0` by default (an older Claude version) to prevent throttle issues during intensive agent operations. You can override this with the `--model-id` flag if needed.

### Quick Agentic Commands

The `agentic` command provides the simplest way to perform file operations:

```shell
# File reading and analysis
chat-cli agentic "read the main.go file and explain what it does"
chat-cli agentic "show me the contents of all configuration files"

# File creation
chat-cli agentic "create a .gitignore file for a Go project"
chat-cli agentic "create a README.md with basic project information"

# File organization
chat-cli agentic "list all .go files in this directory"
chat-cli agentic "organize documentation files into a docs folder"

# File modification
chat-cli agentic "add proper headers to all source files"
chat-cli agentic "update the version number in package.json to 2.0.0"
```

### Advanced Agent Management

For more control over agent behavior, use the dedicated `agent` commands:

#### List Available Agents

```shell
chat-cli agent list
```

Example output:
```
Available agents (1):

Name: file_edit_agent
Description: An agent specialized in reading, writing, and modifying files in the current working directory
Tools: read_file, write_file, list_files
```

#### Get Agent Information

```shell
chat-cli agent info file_edit_agent
```

This displays detailed information about the agent including:
- Available tools and their descriptions
- Parameter schemas for each tool
- Usage examples

#### Run Specific Agents

```shell
# Auto-select best agent for task
chat-cli agent run "create test files for all Go modules"

# Use specific agent
chat-cli agent run file_edit_agent "analyze project structure and suggest improvements"
```

### Available Agents

#### File Edit Agent

**Name:** `file_edit_agent`

**Capabilities:**
- Read file contents with detailed analysis
- Create new files with appropriate content
- List directory contents with filtering
- Modify existing files while preserving structure
- Organize files and directories

**Available Tools:**

1. **read_file** - Read and analyze file contents
   ```shell
   chat-cli agentic "read config.yaml and explain each setting"
   ```

2. **write_file** - Create or overwrite files
   ```shell
   chat-cli agentic "create a Docker file for this Go application"
   ```

3. **list_files** - Browse directory structure
   ```shell
   chat-cli agentic "show me all files larger than 1MB"
   ```

### Common Use Cases

#### Development Workflow

```shell
# Project initialization
chat-cli agentic "create a basic Go module structure with main.go and go.mod"

# Code analysis
chat-cli agentic "read all .go files and create a summary of the project's functionality"

# Documentation generation
chat-cli agentic "create API documentation based on function comments"

# Testing setup
chat-cli agentic "generate test files for all exported functions"
```

#### File Management

```shell
# Cleanup operations
chat-cli agentic "remove all temporary and backup files"

# Organization tasks
chat-cli agentic "move all documentation to a docs/ directory"

# Standardization
chat-cli agentic "ensure all source files have consistent formatting"

# Configuration management
chat-cli agentic "update all config files to use the new database connection string"
```

#### Content Operations

```shell
# Content analysis
chat-cli agentic "analyze all markdown files and create a table of contents"

# Batch operations
chat-cli agentic "add license headers to all source files"

# Template generation
chat-cli agentic "create boilerplate files for a new microservice"
```

### Security and Safety

All agent operations are subject to security restrictions:

- **Directory Scope**: Operations are limited to the current working directory and subdirectories
- **Safe Defaults**: File permissions and operations use conservative defaults
- **Transparent Actions**: All operations are logged and reported
- **Error Prevention**: Robust validation prevents unsafe operations

### Configuration

Agents inherit configuration from your chat-cli settings:

```shell
# Use specific model for agent operations (default is claude-3-sonnet-20240229 for throttle prevention)
chat-cli agentic "task description" --model-id anthropic.claude-3-5-haiku-20241022-v1:0

# Use different AWS region
chat-cli agentic "task description" --region us-west-2
```

Configuration precedence (highest to lowest):
1. Command line flags
2. Configuration file settings
3. Built-in defaults

### Advanced Context Usage

You can provide additional context to agents using JSON:

```shell
chat-cli agent run file_edit_agent "organize files by type" --context '{"exclude_dirs": ["node_modules", ".git"], "preserve_structure": true}'
```

### Troubleshooting

**Common Issues:**

- **"No suitable agent found"**: Use more specific task descriptions or try the `agentic` command
- **"Access denied"**: Ensure you're in the correct directory; agents can only access the current working directory
- **"Tool execution failed"**: Check file paths and permissions; verify files exist

**Debug Information:**

Agent execution results include:
- Success/failure status
- Detailed error messages
- Individual tool execution results
- Summary of actions performed

**Best Practices:**

1. **Be Specific**: Use clear, detailed task descriptions
   - Good: "Create a README.md with installation instructions and usage examples"
   - Avoid: "Make documentation"

2. **Work from Correct Directory**: Always run agent commands from your project root
3. **Verify Results**: Check agent output to confirm operations completed successfully
4. **Start Simple**: Break complex tasks into smaller, manageable steps

### Examples by Category

#### File Analysis
```shell
# Code review assistance
chat-cli agentic "analyze main.go for potential improvements"

# Dependency analysis
chat-cli agentic "examine go.mod and list all external dependencies"

# Security review
chat-cli agentic "check all configuration files for hardcoded secrets"
```

#### Project Setup
```shell
# New project structure
chat-cli agentic "create a standard Go project layout with cmd/, pkg/, and internal/ directories"

# CI/CD setup
chat-cli agentic "create a GitHub Actions workflow for Go testing and building"

# Documentation framework
chat-cli agentic "set up comprehensive project documentation structure"
```

#### Maintenance Tasks
```shell
# Code cleanup
chat-cli agentic "remove unused imports from all Go files"

# License compliance
chat-cli agentic "add SPDX license identifiers to all source files"

# Version management
chat-cli agentic "update version numbers across all configuration files"
```

