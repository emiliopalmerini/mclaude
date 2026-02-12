package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	sqlc "github.com/emiliopalmerini/mclaude/sqlc/generated"
)

func TestHookDispatcher_SessionEnd(t *testing.T) {
	db, cleanup := testDB(t)
	defer cleanup()

	testDBOverride = db
	defer func() { testDBOverride = nil }()

	ctx := context.Background()
	queries := sqlc.New(db)

	sessionID := "test-hook-se-" + fmt.Sprintf("%d", time.Now().UnixNano())

	transcriptPath, err := filepath.Abs("testdata/transcript.jsonl")
	if err != nil {
		t.Fatalf("Failed to get transcript path: %v", err)
	}

	input := map[string]string{
		"session_id":      sessionID,
		"transcript_path": transcriptPath,
		"cwd":             "/test/project",
		"permission_mode": "default",
		"hook_event_name": "SessionEnd",
		"reason":          "exit",
	}
	inputJSON, _ := json.Marshal(input)

	// Pipe stdin
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()
	r, w, _ := os.Pipe()
	os.Stdin = r
	go func() {
		_, _ = w.Write(inputJSON)
		_ = w.Close()
	}()

	// Capture stdout
	oldStdout := os.Stdout
	rOut, wOut, _ := os.Pipe()
	os.Stdout = wOut

	err = runHook(nil, nil)

	_ = wOut.Close()
	os.Stdout = oldStdout
	var stdout bytes.Buffer
	_, _ = stdout.ReadFrom(rOut)

	if err != nil {
		t.Fatalf("hook dispatcher failed: %v\nOutput: %s", err, stdout.String())
	}

	// Verify session was created (same as record test)
	session, err := queries.GetSessionByID(ctx, sessionID)
	if err != nil {
		t.Fatalf("Failed to get session: %v", err)
	}

	assertEqual(t, "session.ExitReason", "exit", session.ExitReason)
	assertEqual(t, "session.Cwd", "/test/project", session.Cwd)

	// Verify metrics
	metrics, err := queries.GetSessionMetricsBySessionID(ctx, sessionID)
	if err != nil {
		t.Fatalf("Failed to get metrics: %v", err)
	}

	if metrics.TokenInput <= 0 {
		t.Error("Expected positive token input")
	}
}

func TestHookDispatcher_UnknownEvent(t *testing.T) {
	input := []byte(`{"session_id":"abc","hook_event_name":"FutureEvent"}`)

	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()
	r, w, _ := os.Pipe()
	os.Stdin = r
	go func() {
		_, _ = w.Write(input)
		_ = w.Close()
	}()

	err := runHook(nil, nil)
	if err == nil {
		t.Fatal("expected error for unknown event type")
	}
}

func TestHookDispatcher_InvalidJSON(t *testing.T) {
	input := []byte(`not valid json`)

	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()
	r, w, _ := os.Pipe()
	os.Stdin = r
	go func() {
		_, _ = w.Write(input)
		_ = w.Close()
	}()

	err := runHook(nil, nil)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}
