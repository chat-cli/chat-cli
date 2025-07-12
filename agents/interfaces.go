package agents

import (
	"context"
)

// Tool represents a capability that an agent can use
type Tool interface {
	// Name returns the unique identifier for this tool
	Name() string
	
	// Description returns a human-readable description of what this tool does
	Description() string
	
	// Execute runs the tool with the given parameters and returns the result
	Execute(ctx context.Context, params map[string]interface{}) (interface{}, error)
	
	// Schema returns the JSON schema for the tool's parameters
	Schema() map[string]interface{}
}

// Agent represents an autonomous agent that can perform tasks
type Agent interface {
	// Name returns the unique identifier for this agent
	Name() string
	
	// Description returns a human-readable description of the agent's capabilities
	Description() string
	
	// Tools returns the list of tools this agent can use
	Tools() []Tool
	
	// Execute runs the agent with the given task and context
	Execute(ctx context.Context, task string, context map[string]interface{}) (*AgentResult, error)
	
	// CanHandle returns true if this agent can handle the given task
	CanHandle(task string) bool
}

// AgentResult represents the result of an agent execution
type AgentResult struct {
	Success     bool                   `json:"success"`
	Message     string                 `json:"message"`
	Data        interface{}            `json:"data,omitempty"`
	ToolResults []ToolResult          `json:"tool_results,omitempty"`
	Error       string                 `json:"error,omitempty"`
}

// ToolResult represents the result of a tool execution
type ToolResult struct {
	ToolName   string      `json:"tool_name"`
	Success    bool        `json:"success"`
	Result     interface{} `json:"result,omitempty"`
	Error      string      `json:"error,omitempty"`
}

// Registry manages available agents and their registration
type Registry interface {
	// RegisterAgent adds an agent to the registry
	RegisterAgent(agent Agent) error
	
	// GetAgent retrieves an agent by name
	GetAgent(name string) (Agent, error)
	
	// ListAgents returns all registered agents
	ListAgents() []Agent
	
	// FindAgentForTask finds the best agent to handle a given task
	FindAgentForTask(task string) (Agent, error)
}