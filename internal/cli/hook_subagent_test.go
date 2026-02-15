package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"
	"time"
)

func TestHandleSubagentStart_CreatesTracking(t *testing.T) {
	db, cleanup := testDB(t)
	defer cleanup()

	testDBOverride = db
	defer func() { testDBOverride = nil }()

	input := map[string]string{
		"session_id":      "sess-sa-1",
		"transcript_path": "/tmp/transcript.jsonl",
		"cwd":             "/project",
		"permission_mode": "default",
		"hook_event_name": "SubagentStart",
		"agent_id":        "agent-001",
		"agent_type":      "Explore",
	}

	_, err := runHookWithInput(t, input)
	if err != nil {
		t.Fatalf("SubagentStart handler failed: %v", err)
	}

	// Verify tracking row was created
	var agentID, sessionID, agentType string
	err = db.QueryRowContext(context.Background(),
		"SELECT agent_id, session_id, agent_type FROM hook_subagent_tracking WHERE agent_id = ?",
		"agent-001",
	).Scan(&agentID, &sessionID, &agentType)
	if err != nil {
		t.Fatalf("Failed to query tracking row: %v", err)
	}

	assertEqual(t, "agentID", "agent-001", agentID)
	assertEqual(t, "sessionID", "sess-sa-1", sessionID)
	assertEqual(t, "agentType", "Explore", agentType)
}

func TestHandleSubagentStop_ParsesAndSaves(t *testing.T) {
	db, cleanup := testDB(t)
	defer cleanup()

	testDBOverride = db
	defer func() { testDBOverride = nil }()

	ctx := context.Background()
	sessionID := "sess-sa-2-" + fmt.Sprintf("%d", time.Now().UnixNano())

	// First create a project and session so we have a valid session_id for FK
	_, err := db.ExecContext(ctx,
		"INSERT INTO projects (id, path, name, created_at) VALUES (?, ?, ?, ?)",
		"proj-sa-test", "/project", "test-project", time.Now().UTC().Format(time.RFC3339),
	)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}
	_, err = db.ExecContext(ctx,
		"INSERT INTO sessions (id, project_id, transcript_path, cwd, permission_mode, exit_reason, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
		sessionID, "proj-sa-test", "/tmp/transcript.jsonl", "/project", "default", "exit", time.Now().UTC().Format(time.RFC3339),
	)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Seed a tracking row (simulates SubagentStart having run)
	_, err = db.ExecContext(ctx,
		"INSERT INTO hook_subagent_tracking (agent_id, session_id, agent_type, started_at) VALUES (?, ?, ?, ?)",
		"agent-002", sessionID, "Explore", time.Now().UTC().Format(time.RFC3339),
	)
	if err != nil {
		t.Fatalf("Failed to insert tracking row: %v", err)
	}

	transcriptPath, err := filepath.Abs("testdata/subagent_transcript.jsonl")
	if err != nil {
		t.Fatalf("Failed to get transcript path: %v", err)
	}

	input := map[string]any{
		"session_id":            sessionID,
		"transcript_path":       "/tmp/main_transcript.jsonl",
		"cwd":                   "/project",
		"permission_mode":       "default",
		"hook_event_name":       "SubagentStop",
		"stop_hook_active":      false,
		"agent_id":              "agent-002",
		"agent_type":            "Explore",
		"agent_transcript_path": transcriptPath,
	}

	_, err = runHookWithInput(t, input)
	if err != nil {
		t.Fatalf("SubagentStop handler failed: %v", err)
	}

	// Verify tracking row was deleted
	var count int
	err = db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM hook_subagent_tracking WHERE agent_id = ?",
		"agent-002",
	).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query tracking: %v", err)
	}
	if count != 0 {
		t.Error("Expected tracking row to be deleted after SubagentStop")
	}

	// Verify session_subagents row was created
	var agentType, agentKind string
	var tokenInput, tokenOutput int64
	err = db.QueryRowContext(ctx,
		"SELECT agent_type, agent_kind, token_input, token_output FROM session_subagents WHERE session_id = ?",
		sessionID,
	).Scan(&agentType, &agentKind, &tokenInput, &tokenOutput)
	if err != nil {
		t.Fatalf("Failed to query session_subagents: %v", err)
	}

	assertEqual(t, "agentType", "Explore", agentType)
	assertEqual(t, "agentKind", "hook", agentKind)
	// From subagent_transcript.jsonl: 200+100=300 input, 80+30=110 output
	assertEqual(t, "tokenInput", int64(300), tokenInput)
	assertEqual(t, "tokenOutput", int64(110), tokenOutput)
}

func TestHandleSubagentStop_NoTrackingRow(t *testing.T) {
	db, cleanup := testDB(t)
	defer cleanup()

	testDBOverride = db
	defer func() { testDBOverride = nil }()

	input := map[string]any{
		"session_id":            "sess-sa-3",
		"transcript_path":       "/tmp/transcript.jsonl",
		"cwd":                   "/project",
		"permission_mode":       "default",
		"hook_event_name":       "SubagentStop",
		"stop_hook_active":      false,
		"agent_id":              "agent-nonexistent",
		"agent_type":            "Explore",
		"agent_transcript_path": "/tmp/subagent.jsonl",
	}

	// Should not crash, just handle gracefully
	_, err := runHookWithInput(t, input)
	if err != nil {
		t.Fatalf("SubagentStop handler should handle missing tracking row gracefully: %v", err)
	}
}

func TestHandleSubagentStop_MissingTranscript(t *testing.T) {
	db, cleanup := testDB(t)
	defer cleanup()

	testDBOverride = db
	defer func() { testDBOverride = nil }()

	ctx := context.Background()

	// Seed tracking row
	_, err := db.ExecContext(ctx,
		"INSERT INTO hook_subagent_tracking (agent_id, session_id, agent_type, started_at) VALUES (?, ?, ?, ?)",
		"agent-003", "sess-sa-4", "Explore", time.Now().UTC().Format(time.RFC3339),
	)
	if err != nil {
		t.Fatalf("Failed to insert tracking row: %v", err)
	}

	input := map[string]any{
		"session_id":            "sess-sa-4",
		"transcript_path":       "/tmp/transcript.jsonl",
		"cwd":                   "/project",
		"permission_mode":       "default",
		"hook_event_name":       "SubagentStop",
		"stop_hook_active":      false,
		"agent_id":              "agent-003",
		"agent_type":            "Explore",
		"agent_transcript_path": "/nonexistent/transcript.jsonl",
	}

	// Should handle missing transcript gracefully (log warning, still delete tracking)
	_, err = runHookWithInput(t, input)
	if err != nil {
		t.Fatalf("SubagentStop should handle missing transcript gracefully: %v", err)
	}

	// Tracking row should still be cleaned up
	var count int
	err = db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM hook_subagent_tracking WHERE agent_id = ?",
		"agent-003",
	).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query tracking: %v", err)
	}
	if count != 0 {
		t.Error("Expected tracking row to be deleted even with missing transcript")
	}
}

// runHookWithInput is a helper - use the shared one from hook_session_start_test.go
// It's already defined there. These tests use it directly.

func TestSubagentInputJSON(t *testing.T) {
	// Verify our test inputs produce valid JSON for the dispatcher
	input := map[string]any{
		"session_id":      "test",
		"hook_event_name": "SubagentStart",
		"agent_id":        "a1",
		"agent_type":      "Explore",
	}
	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("Expected non-empty JSON")
	}
}
