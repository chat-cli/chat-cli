package tools

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/document"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
)

// fakeTool is a test double implementing Tool.
type fakeTool struct {
	name    string
	result  string
	execErr error
}

func (f *fakeTool) Name() string        { return f.name }
func (f *fakeTool) Description() string { return "a fake tool for tests" }
func (f *fakeTool) InputSchema() document.Interface {
	return document.NewLazyDocument(map[string]interface{}{"type": "object"})
}
func (f *fakeTool) Execute(_ context.Context, _ json.RawMessage) (string, error) {
	if f.execErr != nil {
		return "", f.execErr
	}
	return f.result, nil
}
func (f *fakeTool) RequiresConfirmation() bool { return false }
func (f *fakeTool) ConfirmationSummary(_ json.RawMessage) (string, string, error) {
	return "", "", nil
}

// fakeDestructiveTool is a test double for a destructive tool that requires
// confirmation, with a controllable ConfirmationSummary outcome.
type fakeDestructiveTool struct {
	name       string
	result     string
	summaryErr error
}

func (f *fakeDestructiveTool) Name() string        { return f.name }
func (f *fakeDestructiveTool) Description() string { return "a fake destructive tool for tests" }
func (f *fakeDestructiveTool) InputSchema() document.Interface {
	return document.NewLazyDocument(map[string]interface{}{"type": "object"})
}
func (f *fakeDestructiveTool) Execute(_ context.Context, _ json.RawMessage) (string, error) {
	return f.result, nil
}
func (f *fakeDestructiveTool) RequiresConfirmation() bool { return true }
func (f *fakeDestructiveTool) ConfirmationSummary(_ json.RawMessage) (string, string, error) {
	if f.summaryErr != nil {
		return "", "", f.summaryErr
	}
	return "will do something destructive", "pattern-key", nil
}

// fakeGate is a test double for PermissionGate returning a fixed Decision
// and recording whether Check was called.
type fakeGate struct {
	decision Decision
	called   bool
}

func (g *fakeGate) Check(_, _, _ string) Decision {
	g.called = true
	return g.decision
}

func TestNewRegistry_EmptyToolConfiguration(t *testing.T) {
	r := NewRegistry()

	if cfg := r.ToolConfiguration(); cfg != nil {
		t.Errorf("expected nil ToolConfiguration for an empty registry, got %v", cfg)
	}
}

func TestNewRegistry_NonEmptyToolConfiguration(t *testing.T) {
	r := NewRegistry()
	r.Register(&fakeTool{name: "fake_tool"})

	cfg := r.ToolConfiguration()
	if cfg == nil {
		t.Fatal("expected non-nil ToolConfiguration once a tool is registered")
	}
	if len(cfg.Tools) != 1 {
		t.Fatalf("expected 1 tool in ToolConfiguration, got %d", len(cfg.Tools))
	}
}

func TestRegistry_Dispatch_UnknownTool(t *testing.T) {
	r := NewRegistry()

	result := r.Dispatch(context.Background(), ToolCall{
		Name:      "nonexistent",
		ToolUseID: "abc123",
		Input:     []byte(`{}`),
	}, nil)

	if result.Status != types.ToolResultStatusError {
		t.Errorf("expected error status for unknown tool, got %v", result.Status)
	}
	if result.ToolUseId == nil || *result.ToolUseId != "abc123" {
		t.Errorf("expected ToolUseId to be echoed back, got %v", result.ToolUseId)
	}
}

func TestRegistry_Dispatch_Success(t *testing.T) {
	r := NewRegistry()
	r.Register(&fakeTool{name: "fake_tool", result: "hello from fake tool"})

	result := r.Dispatch(context.Background(), ToolCall{
		Name:      "fake_tool",
		ToolUseID: "call-1",
		Input:     []byte(`{}`),
	}, nil)

	if result.Status != types.ToolResultStatusSuccess {
		t.Errorf("expected success status, got %v", result.Status)
	}
	if len(result.Content) != 1 {
		t.Fatalf("expected 1 content block, got %d", len(result.Content))
	}
	textBlock, ok := result.Content[0].(*types.ToolResultContentBlockMemberText)
	if !ok {
		t.Fatalf("expected text content block, got %T", result.Content[0])
	}
	if textBlock.Value != "hello from fake tool" {
		t.Errorf("expected result text %q, got %q", "hello from fake tool", textBlock.Value)
	}
}

func TestRegistry_Dispatch_ExecutionError(t *testing.T) {
	r := NewRegistry()
	r.Register(&fakeTool{name: "fake_tool", execErr: errors.New("boom")})

	result := r.Dispatch(context.Background(), ToolCall{
		Name:      "fake_tool",
		ToolUseID: "call-2",
		Input:     []byte(`{}`),
	}, nil)

	if result.Status != types.ToolResultStatusError {
		t.Errorf("expected error status when tool execution fails, got %v", result.Status)
	}
	textBlock, ok := result.Content[0].(*types.ToolResultContentBlockMemberText)
	if !ok {
		t.Fatalf("expected text content block, got %T", result.Content[0])
	}
	if textBlock.Value != "boom" {
		t.Errorf("expected error text %q, got %q", "boom", textBlock.Value)
	}
}

func TestRegistry_Dispatch_NonDestructiveToolNeverConsultsGate(t *testing.T) {
	r := NewRegistry()
	r.Register(&fakeTool{name: "fake_tool", result: "ok"})
	gate := &fakeGate{decision: DecisionDeny} // if consulted, would deny - proves it wasn't consulted

	result := r.Dispatch(context.Background(), ToolCall{
		Name:      "fake_tool",
		ToolUseID: "call-3",
		Input:     []byte(`{}`),
	}, gate)

	if gate.called {
		t.Error("expected the gate to never be consulted for a non-destructive tool")
	}
	if result.Status != types.ToolResultStatusSuccess {
		t.Errorf("expected success status, got %v", result.Status)
	}
}

func TestRegistry_Dispatch_ConfirmationSummaryErrorNeverReachesGateOrExecute(t *testing.T) {
	r := NewRegistry()
	r.Register(&fakeDestructiveTool{name: "destructive_tool", summaryErr: errors.New("bad input")})
	gate := &fakeGate{decision: DecisionAllowOnce}

	result := r.Dispatch(context.Background(), ToolCall{
		Name:      "destructive_tool",
		ToolUseID: "call-4",
		Input:     []byte(`{}`),
	}, gate)

	if gate.called {
		t.Error("expected the gate to never be consulted when ConfirmationSummary fails")
	}
	if result.Status != types.ToolResultStatusError {
		t.Errorf("expected error status, got %v", result.Status)
	}
}

func TestRegistry_Dispatch_GateDenyBlocksExecute(t *testing.T) {
	r := NewRegistry()
	r.Register(&fakeDestructiveTool{name: "destructive_tool", result: "should never see this"})
	gate := &fakeGate{decision: DecisionDeny}

	result := r.Dispatch(context.Background(), ToolCall{
		Name:      "destructive_tool",
		ToolUseID: "call-5",
		Input:     []byte(`{}`),
	}, gate)

	if !gate.called {
		t.Error("expected the gate to be consulted for a destructive tool")
	}
	if result.Status != types.ToolResultStatusError {
		t.Errorf("expected a declined call to return error status, got %v", result.Status)
	}
	textBlock, ok := result.Content[0].(*types.ToolResultContentBlockMemberText)
	if !ok {
		t.Fatalf("expected text content block, got %T", result.Content[0])
	}
	if textBlock.Value == "should never see this" {
		t.Error("Execute must not run when the gate denies the call")
	}
}

func TestRegistry_Dispatch_GateAllowRunsExecute(t *testing.T) {
	for _, decision := range []Decision{DecisionAllowOnce, DecisionAllowSession, DecisionAllowAlways} {
		r := NewRegistry()
		r.Register(&fakeDestructiveTool{name: "destructive_tool", result: "executed"})
		gate := &fakeGate{decision: decision}

		result := r.Dispatch(context.Background(), ToolCall{
			Name:      "destructive_tool",
			ToolUseID: "call-6",
			Input:     []byte(`{}`),
		}, gate)

		if !gate.called {
			t.Errorf("decision %v: expected the gate to be consulted", decision)
		}
		if result.Status != types.ToolResultStatusSuccess {
			t.Errorf("decision %v: expected success status, got %v", decision, result.Status)
		}
	}
}
