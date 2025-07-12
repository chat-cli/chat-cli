# Agents

Chat-CLI includes an extensible agent system that allows autonomous AI agents to perform complex tasks using natural language instructions. This document provides comprehensive information about using and extending the agent system.

## Overview

The agent system consists of several key components:

- **Agents**: Autonomous AI entities that can understand tasks and plan their execution
- **Tools**: Specific capabilities that agents can use to interact with the system
- **Registry**: A management system for organizing and accessing available agents
- **CLI Integration**: Commands that make agent functionality accessible from the command line

**Model Configuration:** The agent system uses `anthropic.claude-3-sonnet-20240229-v1:0` by default to avoid throttle issues that can occur with newer models during intensive operations. This ensures reliable performance for complex multi-step tasks.

## Quick Start

### Simple Agent Commands

The easiest way to use agents is through the `agentic` command:

```shell
# Read a file
chat-cli agentic "read the package.json file and tell me about the dependencies"

# Create a file
chat-cli agentic "create a .gitignore file for a Node.js project"

# List files with criteria
chat-cli agentic "show me all markdown files in this directory"

# Modify files
chat-cli agentic "add a license header to all .go files"
```

### Advanced Agent Management

For more control, use the `agent` command:

```shell
# List all available agents
chat-cli agent list

# Get information about a specific agent
chat-cli agent info file_edit_agent

# Run an agent with a specific task
chat-cli agent run file_edit_agent "organize my documentation files"
```

## Available Agents

### File Edit Agent

The File Edit Agent (`file_edit_agent`) is specialized for file system operations.

**Capabilities:**
- Reading file contents
- Writing and creating files
- Listing directory contents
- File organization and management
- Text manipulation and editing

**Available Tools:**
- `read_file`: Read the contents of a file
- `write_file`: Write content to a file (creates directories as needed)
- `list_files`: List files and directories with metadata

**Example Tasks:**
```shell
# File analysis
chat-cli agent run file_edit_agent "analyze the main.go file and create documentation"

# Batch operations
chat-cli agent run file_edit_agent "create README files for each subdirectory"

# Code organization
chat-cli agent run file_edit_agent "move all test files to a tests/ directory"
```

## Agent Architecture

### Agent Interface

All agents implement the following interface:

```go
type Agent interface {
    Name() string                    // Unique identifier
    Description() string             // Human-readable description
    Tools() []Tool                   // Available tools
    CanHandle(task string) bool      // Task compatibility check
    Execute(ctx context.Context, task string, context map[string]interface{}) (*AgentResult, error)
}
```

### Tool Interface

Tools provide specific capabilities to agents:

```go
type Tool interface {
    Name() string                    // Tool identifier
    Description() string             // What the tool does
    Execute(ctx context.Context, params map[string]interface{}) (interface{}, error)
    Schema() map[string]interface{}  // Parameter schema
}
```

## Security Model

The agent system implements several security measures:

### Directory Restrictions

All file operations are restricted to the current working directory and its subdirectories. Agents cannot:
- Access files outside the current directory tree
- Follow symlinks that point outside the allowed area
- Perform operations on system files

### Safe Defaults

- File permissions are set conservatively (0644 for files, 0755 for directories)
- Existing files are only modified when explicitly requested
- Error handling prevents partial or corrupted operations

### Transparent Operations

- All tool executions are logged and reported
- Failed operations provide detailed error messages
- Users can see exactly what actions were performed

## Usage Patterns

### Interactive File Management

```shell
# Explore project structure
chat-cli agentic "give me an overview of this project's file structure"

# Clean up files
chat-cli agentic "remove any temporary or backup files"

# Standardize formatting
chat-cli agentic "ensure all markdown files have proper headers"
```

### Code Development Tasks

```shell
# Generate boilerplate
chat-cli agentic "create a basic Go module structure"

# Documentation generation
chat-cli agentic "create API documentation based on the code comments"

# Test file creation
chat-cli agentic "generate test files for all exported functions"
```

### Project Organization

```shell
# Directory structure
chat-cli agentic "organize source files into appropriate directories"

# Configuration management
chat-cli agentic "update all config files to use the new API endpoint"

# Build system
chat-cli agentic "create a Makefile for this Go project"
```

## Extending the Agent System

### Creating New Agents

To create a new agent:

1. **Implement the Agent Interface**:
```go
type MyCustomAgent struct {
    tools []Tool
}

func (a *MyCustomAgent) Name() string {
    return "my_custom_agent"
}

func (a *MyCustomAgent) Description() string {
    return "Performs custom operations"
}

func (a *MyCustomAgent) Tools() []Tool {
    return a.tools
}

func (a *MyCustomAgent) CanHandle(task string) bool {
    // Implement task detection logic
    return strings.Contains(strings.ToLower(task), "custom")
}

func (a *MyCustomAgent) Execute(ctx context.Context, task string, context map[string]interface{}) (*AgentResult, error) {
    // Implement agent logic
}
```

2. **Register the Agent**:
```go
agent := &MyCustomAgent{tools: myTools}
agentRegistry.RegisterAgent(agent)
```

### Creating New Tools

To create a new tool:

```go
type MyTool struct{}

func (t *MyTool) Name() string {
    return "my_tool"
}

func (t *MyTool) Description() string {
    return "Performs a specific operation"
}

func (t *MyTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
    // Implement tool functionality
    return result, nil
}

func (t *MyTool) Schema() map[string]interface{} {
    return map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "param1": map[string]interface{}{
                "type": "string",
                "description": "Parameter description",
            },
        },
        "required": []string{"param1"},
    }
}
```

### Integration with LLMs

Agents use the configured LLM (via Bedrock) to:
- Parse and understand natural language tasks
- Plan multi-step operations
- Generate appropriate tool calls
- Provide human-readable result summaries

The system prompt guides the LLM to:
- Use available tools appropriately
- Follow security restrictions
- Provide clear feedback
- Handle errors gracefully

## Configuration

Agents inherit configuration from the main chat-cli settings:

- **Model Selection**: Uses `anthropic.claude-3-sonnet-20240229-v1:0` by default for throttle prevention, or configured model-id/custom-arn
- **Region**: Inherits AWS region settings
- **Credentials**: Uses the same AWS credential chain

You can override these settings using command-line flags:

```shell
# Override the default model if needed
chat-cli agentic "task description" --model-id anthropic.claude-3-5-sonnet-20240620-v1:0

# Use different AWS region
chat-cli agentic "task description" --region us-west-2
```

## Troubleshooting

### Common Issues

**Agent not found:**
```
Error: No suitable agent found for task
```
- Check available agents with `chat-cli agent list`
- Use more specific task descriptions
- Try the general `agentic` command instead

**Permission denied:**
```
Error: access denied: file must be within current working directory
```
- Ensure you're in the correct directory
- Use relative paths for file operations
- Check file permissions

**Tool execution failed:**
```
Tool read_file failed: file not found
```
- Verify file paths are correct
- Check if files exist with `chat-cli agentic "list files in current directory"`
- Ensure proper file permissions

### Debug Mode

For detailed execution information, you can examine the agent result output which includes:
- Success/failure status
- Detailed error messages
- Individual tool execution results
- Timing information

## Best Practices

### Task Description

Write clear, specific task descriptions:

**Good:**
- "Create a README.md file with project description and installation instructions"
- "Read all .go files and generate a summary of exported functions"
- "Organize source files into src/, docs/, and tests/ directories"

**Avoid:**
- "Fix everything"
- "Make it better"
- "Do something with files"

### File Operations

- Always work from the intended directory
- Use relative paths when possible
- Be specific about file locations and names
- Verify results after complex operations

### Error Handling

- Check agent results for success status
- Review tool execution details when operations fail
- Use simpler tasks to debug complex operations
- Verify AWS credentials and region settings

## Future Extensions

The agent system is designed to be extended with additional capabilities:

- **Network Tools**: HTTP requests, API interactions
- **Database Tools**: Query execution, schema management
- **Development Tools**: Code compilation, testing, deployment
- **Analysis Tools**: Log parsing, metrics collection
- **Integration Tools**: Git operations, CI/CD workflows

Each new tool type follows the same interface pattern and security model, ensuring consistent behavior across the system.