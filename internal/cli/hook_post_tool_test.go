package cli

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestHandlePostToolUse_SavesEvent(t *testing.T) {
	db, cleanup := testDB(t)
	defer cleanup()

	testDBOverride = db
	defer func() { testDBOverride = nil }()

	input := map[string]any{
		"session_id":      "sess-ptu-1",
		"transcript_path": "/tmp/transcript.jsonl",
		"cwd":             "/project",
		"permission_mode": "default",
		"hook_event_name": "PostToolUse",
		"tool_name":       "Bash",
		"tool_input":      map[string]string{"command": "go test ./..."},
		"tool_response":   map[string]string{"stdout": "PASS"},
		"tool_use_id":     "tool_abc123",
	}

	_, err := runHookWithInput(t, input)
	if err != nil {
		t.Fatalf("PostToolUse handler failed: %v", err)
	}

	// Verify tool event was saved
	var toolName, toolUseID string
	var toolInput, toolResponse string
	err = db.QueryRowContext(context.Background(),
		"SELECT tool_name, tool_use_id, tool_input, tool_response FROM tool_events WHERE tool_use_id = ?",
		"tool_abc123",
	).Scan(&toolName, &toolUseID, &toolInput, &toolResponse)
	if err != nil {
		t.Fatalf("Failed to query tool event: %v", err)
	}

	assertEqual(t, "toolName", "Bash", toolName)
	assertEqual(t, "toolUseID", "tool_abc123", toolUseID)

	// Verify input/response are stored as JSON
	var parsedInput map[string]string
	if err := json.Unmarshal([]byte(toolInput), &parsedInput); err != nil {
		t.Fatalf("tool_input is not valid JSON: %v", err)
	}
	assertEqual(t, "toolInput.command", "go test ./...", parsedInput["command"])
}

func TestHandlePostToolUse_TruncatesLargeResponse(t *testing.T) {
	db, cleanup := testDB(t)
	defer cleanup()

	testDBOverride = db
	defer func() { testDBOverride = nil }()

	// Create a response larger than 10KB
	largeResponse := strings.Repeat("x", 100*1024) // 100KB

	input := map[string]any{
		"session_id":      "sess-ptu-2",
		"transcript_path": "/tmp/transcript.jsonl",
		"cwd":             "/project",
		"permission_mode": "default",
		"hook_event_name": "PostToolUse",
		"tool_name":       "Read",
		"tool_input":      map[string]string{"file_path": "/big/file.go"},
		"tool_response":   largeResponse,
		"tool_use_id":     "tool_large",
	}

	_, err := runHookWithInput(t, input)
	if err != nil {
		t.Fatalf("PostToolUse handler failed: %v", err)
	}

	// Verify response was truncated
	var toolResponse string
	err = db.QueryRowContext(context.Background(),
		"SELECT tool_response FROM tool_events WHERE tool_use_id = ?",
		"tool_large",
	).Scan(&toolResponse)
	if err != nil {
		t.Fatalf("Failed to query tool event: %v", err)
	}

	if len(toolResponse) > 10*1024+100 { // 10KB + some buffer for truncation message
		t.Errorf("Expected response to be truncated to ~10KB, got %d bytes", len(toolResponse))
	}
}

func TestHandlePostToolUse_DuplicateToolUseID(t *testing.T) {
	db, cleanup := testDB(t)
	defer cleanup()

	testDBOverride = db
	defer func() { testDBOverride = nil }()

	input := map[string]any{
		"session_id":      "sess-ptu-3",
		"transcript_path": "/tmp/transcript.jsonl",
		"cwd":             "/project",
		"permission_mode": "default",
		"hook_event_name": "PostToolUse",
		"tool_name":       "Bash",
		"tool_input":      map[string]string{"command": "echo hello"},
		"tool_response":   map[string]string{"stdout": "hello"},
		"tool_use_id":     "tool_dup",
	}

	// First insert should succeed
	_, err := runHookWithInput(t, input)
	if err != nil {
		t.Fatalf("First PostToolUse failed: %v", err)
	}

	// Second insert with same tool_use_id should not error
	_, err = runHookWithInput(t, input)
	if err != nil {
		t.Fatalf("Duplicate PostToolUse should not error: %v", err)
	}

	// Verify only one row exists
	var count int
	err = db.QueryRowContext(context.Background(),
		"SELECT COUNT(*) FROM tool_events WHERE tool_use_id = ?",
		"tool_dup",
	).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count: %v", err)
	}
	assertEqual(t, "count", 1, count)
}
