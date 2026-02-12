package cli

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	sqlc "github.com/emiliopalmerini/mclaude/sqlc/generated"
)

func seedActiveExperiment(t *testing.T, db *sql.DB, name, hypothesis string) string {
	t.Helper()
	ctx := context.Background()
	queries := sqlc.New(db)
	expID := "exp-" + fmt.Sprintf("%d", time.Now().UnixNano())
	now := time.Now().UTC().Format(time.RFC3339)
	err := queries.CreateExperiment(ctx, sqlc.CreateExperimentParams{
		ID:          expID,
		Name:        name,
		Description: sql.NullString{String: "test description", Valid: true},
		Hypothesis:  sql.NullString{String: hypothesis, Valid: true},
		StartedAt:   now,
		CreatedAt:   now,
		IsActive:    1,
	})
	if err != nil {
		t.Fatalf("Failed to create experiment: %v", err)
	}
	return expID
}

func runHookWithInput(t *testing.T, input any) (string, error) {
	t.Helper()
	inputJSON, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal input: %v", err)
	}

	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()
	r, w, _ := os.Pipe()
	os.Stdin = r
	go func() {
		_, _ = w.Write(inputJSON)
		_ = w.Close()
	}()

	oldStdout := os.Stdout
	rOut, wOut, _ := os.Pipe()
	os.Stdout = wOut

	hookErr := runHook(nil, nil)

	_ = wOut.Close()
	os.Stdout = oldStdout
	var stdout bytes.Buffer
	_, _ = stdout.ReadFrom(rOut)

	return stdout.String(), hookErr
}

func TestHandleSessionStart_WithActiveExperiment(t *testing.T) {
	db, cleanup := testDB(t)
	defer cleanup()

	testDBOverride = db
	defer func() { testDBOverride = nil }()

	seedActiveExperiment(t, db, "test-experiment", "Testing improves quality")

	input := map[string]string{
		"session_id":      "sess-123",
		"transcript_path": "/tmp/transcript.jsonl",
		"cwd":             "/project",
		"permission_mode": "default",
		"hook_event_name": "SessionStart",
		"source":          "cli",
		"model":           "claude-opus-4-6",
	}

	output, err := runHookWithInput(t, input)
	if err != nil {
		t.Fatalf("SessionStart handler failed: %v", err)
	}

	var resp HookResponse
	if err := json.Unmarshal([]byte(output), &resp); err != nil {
		t.Fatalf("Failed to parse response JSON: %v\nOutput: %s", err, output)
	}

	if resp.AdditionalContext == "" {
		t.Fatal("Expected additionalContext to contain experiment info")
	}

	if !bytes.Contains([]byte(resp.AdditionalContext), []byte("test-experiment")) {
		t.Errorf("Expected additionalContext to contain experiment name, got: %s", resp.AdditionalContext)
	}
	if !bytes.Contains([]byte(resp.AdditionalContext), []byte("Testing improves quality")) {
		t.Errorf("Expected additionalContext to contain hypothesis, got: %s", resp.AdditionalContext)
	}
}

func TestHandleSessionStart_NoActiveExperiment(t *testing.T) {
	db, cleanup := testDB(t)
	defer cleanup()

	testDBOverride = db
	defer func() { testDBOverride = nil }()

	// Ensure no active experiments exist
	queries := sqlc.New(db)
	_ = queries.DeactivateAllExperiments(context.Background())

	input := map[string]string{
		"session_id":      "sess-456",
		"transcript_path": "/tmp/transcript.jsonl",
		"cwd":             "/project",
		"permission_mode": "default",
		"hook_event_name": "SessionStart",
		"source":          "cli",
	}

	output, err := runHookWithInput(t, input)
	if err != nil {
		t.Fatalf("SessionStart handler failed: %v", err)
	}

	// No experiment → no output (or empty JSON)
	if output != "" {
		var resp HookResponse
		if err := json.Unmarshal([]byte(output), &resp); err != nil {
			t.Fatalf("Failed to parse response: %v\nOutput: %s", err, output)
		}
		if resp.AdditionalContext != "" {
			t.Errorf("Expected empty additionalContext, got: %s", resp.AdditionalContext)
		}
	}
}

func TestHandleSessionStart_OutputFormat(t *testing.T) {
	db, cleanup := testDB(t)
	defer cleanup()

	testDBOverride = db
	defer func() { testDBOverride = nil }()

	seedActiveExperiment(t, db, "format-test", "Output is valid JSON")

	input := map[string]string{
		"session_id":      "sess-789",
		"transcript_path": "/tmp/transcript.jsonl",
		"cwd":             "/project",
		"permission_mode": "default",
		"hook_event_name": "SessionStart",
	}

	output, err := runHookWithInput(t, input)
	if err != nil {
		t.Fatalf("SessionStart handler failed: %v", err)
	}

	// Verify output is valid JSON
	var raw json.RawMessage
	if err := json.Unmarshal([]byte(output), &raw); err != nil {
		t.Fatalf("Output is not valid JSON: %v\nOutput: %s", err, output)
	}

	// Verify it has the expected fields
	var resp map[string]any
	if err := json.Unmarshal([]byte(output), &resp); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if _, ok := resp["additionalContext"]; !ok {
		t.Error("Expected 'additionalContext' key in response")
	}
}
