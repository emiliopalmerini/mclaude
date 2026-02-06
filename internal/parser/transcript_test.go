package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseTranscript_SubagentDetection(t *testing.T) {
	// Create a temp transcript with sub-agent entries
	content := `{"type":"user","timestamp":"2025-01-17T10:00:00Z","message":{"role":"user","content":[{"type":"text","text":"Hello"}]}}
{"type":"assistant","model":"claude-sonnet-4-20250514","timestamp":"2025-01-17T10:00:05Z","message":{"role":"assistant","content":[{"type":"text","text":"Let me search."},{"type":"tool_use","id":"task1","name":"Task","input":{"description":"Find files","subagent_type":"Explore","model":"haiku","prompt":"Search"}}]},"usage":{"input_tokens":100,"output_tokens":50,"cache_read_input_tokens":0,"cache_creation_input_tokens":0}}
{"type":"user","timestamp":"2025-01-17T10:00:10Z","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"task1","content":"Found files"}]},"toolUseResult":{"status":"completed","totalDurationMs":3000,"totalTokens":5000,"totalToolUseCount":3,"usage":{"input_tokens":4000,"output_tokens":500,"cache_read_input_tokens":400,"cache_creation_input_tokens":100}}}
{"type":"assistant","timestamp":"2025-01-17T10:00:15Z","message":{"role":"assistant","content":[{"type":"text","text":"Now commit."},{"type":"tool_use","id":"skill1","name":"Skill","input":{"skill":"commit","args":"-m fix"}}]},"usage":{"input_tokens":80,"output_tokens":30,"cache_read_input_tokens":0,"cache_creation_input_tokens":0}}
{"type":"user","timestamp":"2025-01-17T10:00:20Z","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"skill1","content":"Committed"}]},"toolUseResult":{"status":"completed","totalDurationMs":2000,"totalTokens":3000,"totalToolUseCount":1,"usage":{"input_tokens":2000,"output_tokens":800,"cache_read_input_tokens":150,"cache_creation_input_tokens":50}}}
`

	dir := t.TempDir()
	path := filepath.Join(dir, "transcript.jsonl")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test transcript: %v", err)
	}

	result, err := ParseTranscript("test-session", path)
	if err != nil {
		t.Fatalf("ParseTranscript failed: %v", err)
	}

	// Verify sub-agents were detected
	if len(result.Subagents) != 2 {
		t.Fatalf("Expected 2 subagents, got %d", len(result.Subagents))
	}

	// Verify Task sub-agent
	sa0 := result.Subagents[0]
	assertEqual(t, "sa0.AgentType", "Explore", sa0.AgentType)
	assertEqual(t, "sa0.AgentKind", "task", sa0.AgentKind)
	assertEqual(t, "sa0.TotalTokens", int64(5000), sa0.TotalTokens)
	assertEqual(t, "sa0.TokenInput", int64(4000), sa0.TokenInput)
	assertEqual(t, "sa0.TokenOutput", int64(500), sa0.TokenOutput)
	assertEqual(t, "sa0.TokenCacheRead", int64(400), sa0.TokenCacheRead)
	assertEqual(t, "sa0.TokenCacheWrite", int64(100), sa0.TokenCacheWrite)
	assertEqual(t, "sa0.ToolUseCount", int64(3), sa0.ToolUseCount)
	if sa0.Description == nil || *sa0.Description != "Find files" {
		t.Errorf("Expected description 'Find files', got %v", sa0.Description)
	}
	if sa0.Model == nil || *sa0.Model != "haiku" {
		t.Errorf("Expected model 'haiku', got %v", sa0.Model)
	}
	if sa0.TotalDurationMs == nil || *sa0.TotalDurationMs != 3000 {
		t.Errorf("Expected duration 3000, got %v", sa0.TotalDurationMs)
	}

	// Verify Skill sub-agent
	sa1 := result.Subagents[1]
	assertEqual(t, "sa1.AgentType", "commit", sa1.AgentType)
	assertEqual(t, "sa1.AgentKind", "skill", sa1.AgentKind)
	assertEqual(t, "sa1.TotalTokens", int64(3000), sa1.TotalTokens)
	assertEqual(t, "sa1.TokenInput", int64(2000), sa1.TokenInput)
	assertEqual(t, "sa1.TokenOutput", int64(800), sa1.TokenOutput)
	if sa1.Model != nil {
		t.Errorf("Expected nil model for skill, got %v", sa1.Model)
	}

	// Verify sub-agent tokens were accumulated into session totals
	// Parent tokens: 100+80=180 input, 50+30=80 output
	// Sub-agent tokens: 4000+2000=6000 input, 500+800=1300 output, 400+150=550 cache read, 100+50=150 cache write
	// Total: 180+6000=6180 input, 80+1300=1380 output, 550 cache read, 150 cache write
	assertEqual(t, "metrics.TokenInput", int64(6180), result.Metrics.TokenInput)
	assertEqual(t, "metrics.TokenOutput", int64(1380), result.Metrics.TokenOutput)
	assertEqual(t, "metrics.TokenCacheRead", int64(550), result.Metrics.TokenCacheRead)
	assertEqual(t, "metrics.TokenCacheWrite", int64(150), result.Metrics.TokenCacheWrite)
}

func TestParseTranscript_NoSubagents(t *testing.T) {
	// Transcript with no sub-agent entries
	content := `{"type":"user","timestamp":"2025-01-17T10:00:00Z","message":{"role":"user","content":[{"type":"text","text":"Hello"}]}}
{"type":"assistant","timestamp":"2025-01-17T10:00:05Z","message":{"role":"assistant","content":[{"type":"text","text":"Hi!"}]},"usage":{"input_tokens":100,"output_tokens":50,"cache_read_input_tokens":0,"cache_creation_input_tokens":0}}
`
	dir := t.TempDir()
	path := filepath.Join(dir, "transcript.jsonl")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test transcript: %v", err)
	}

	result, err := ParseTranscript("test-session", path)
	if err != nil {
		t.Fatalf("ParseTranscript failed: %v", err)
	}

	if len(result.Subagents) != 0 {
		t.Errorf("Expected 0 subagents, got %d", len(result.Subagents))
	}
}

func assertEqual[T comparable](t *testing.T, name string, expected, actual T) {
	t.Helper()
	if expected != actual {
		t.Errorf("%s: expected %v, got %v", name, expected, actual)
	}
}
