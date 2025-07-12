package agents

import (
	"fmt"
	"strings"
	"sync"
)

// DefaultRegistry is a thread-safe implementation of the Registry interface
type DefaultRegistry struct {
	agents map[string]Agent
	mutex  sync.RWMutex
}

// NewRegistry creates a new agent registry
func NewRegistry() Registry {
	return &DefaultRegistry{
		agents: make(map[string]Agent),
	}
}

// RegisterAgent adds an agent to the registry
func (r *DefaultRegistry) RegisterAgent(agent Agent) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	name := agent.Name()
	if _, exists := r.agents[name]; exists {
		return fmt.Errorf("agent with name '%s' already registered", name)
	}
	
	r.agents[name] = agent
	return nil
}

// GetAgent retrieves an agent by name
func (r *DefaultRegistry) GetAgent(name string) (Agent, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	agent, exists := r.agents[name]
	if !exists {
		return nil, fmt.Errorf("agent '%s' not found", name)
	}
	
	return agent, nil
}

// ListAgents returns all registered agents
func (r *DefaultRegistry) ListAgents() []Agent {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	agents := make([]Agent, 0, len(r.agents))
	for _, agent := range r.agents {
		agents = append(agents, agent)
	}
	
	return agents
}

// FindAgentForTask finds the best agent to handle a given task
func (r *DefaultRegistry) FindAgentForTask(task string) (Agent, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	taskLower := strings.ToLower(task)
	
	// First pass: find agents that explicitly can handle the task
	for _, agent := range r.agents {
		if agent.CanHandle(taskLower) {
			return agent, nil
		}
	}
	
	// Second pass: basic keyword matching for common tasks
	if strings.Contains(taskLower, "edit") || strings.Contains(taskLower, "file") || strings.Contains(taskLower, "modify") {
		for _, agent := range r.agents {
			if strings.Contains(strings.ToLower(agent.Name()), "file") || strings.Contains(strings.ToLower(agent.Description()), "file") {
				return agent, nil
			}
		}
	}
	
	return nil, fmt.Errorf("no agent found capable of handling task: %s", task)
}