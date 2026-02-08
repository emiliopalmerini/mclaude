package cli

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	sqlc "github.com/emiliopalmerini/mclaude/sqlc/generated"
)

func TestHandleStop_StopHookActive(t *testing.T) {
	db, cleanup := testDB(t)
	defer cleanup()

	testDBOverride = db
	defer func() { testDBOverride = nil }()

	input := map[string]any{
		"session_id":       "sess-stop-1",
		"transcript_path":  "/tmp/transcript.jsonl",
		"cwd":              "/project",
		"permission_mode":  "default",
		"hook_event_name":  "Stop",
		"stop_hook_active": true,
	}

	_, err := runHookWithInput(t, input)
	if err != nil {
		t.Fatalf("Stop handler failed: %v", err)
	}

	// Verify no session was created (should return immediately)
	queries := sqlc.New(db)
	_, err = queries.GetSessionByID(context.Background(), "sess-stop-1")
	if err == nil {
		t.Error("Expected no session to be created when stop_hook_active is true")
	}
}

func TestHandleStop_SnapshotSession(t *testing.T) {
	db, cleanup := testDB(t)
	defer cleanup()

	testDBOverride = db
	defer func() { testDBOverride = nil }()

	transcriptPath, err := filepath.Abs("testdata/transcript.jsonl")
	if err != nil {
		t.Fatalf("Failed to get transcript path: %v", err)
	}

	sessionID := "sess-stop-2-" + fmt.Sprintf("%d", time.Now().UnixNano())

	input := map[string]any{
		"session_id":       sessionID,
		"transcript_path":  transcriptPath,
		"cwd":              "/test/project",
		"permission_mode":  "default",
		"hook_event_name":  "Stop",
		"stop_hook_active": false,
	}

	_, err = runHookWithInput(t, input)
	if err != nil {
		t.Fatalf("Stop handler failed: %v", err)
	}

	// Verify session was created
	queries := sqlc.New(db)
	session, err := queries.GetSessionByID(context.Background(), sessionID)
	if err != nil {
		t.Fatalf("Failed to get session: %v", err)
	}

	assertEqual(t, "session.Cwd", "/test/project", session.Cwd)
	assertEqual(t, "session.PermissionMode", "default", session.PermissionMode)

	// Verify metrics were saved
	metrics, err := queries.GetSessionMetricsBySessionID(context.Background(), sessionID)
	if err != nil {
		t.Fatalf("Failed to get metrics: %v", err)
	}

	if metrics.TokenInput <= 0 {
		t.Error("Expected positive token input")
	}
}

func TestHandleStop_UpdatesExistingSession(t *testing.T) {
	db, cleanup := testDB(t)
	defer cleanup()

	testDBOverride = db
	defer func() { testDBOverride = nil }()

	ctx := context.Background()
	sessionID := "sess-stop-3-" + fmt.Sprintf("%d", time.Now().UnixNano())

	// Pre-create project and session
	_, err := db.ExecContext(ctx,
		"INSERT INTO projects (id, path, name, created_at) VALUES (?, ?, ?, ?)",
		"proj-stop-test", "/test/project", "test-project", time.Now().UTC().Format(time.RFC3339),
	)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	_, err = db.ExecContext(ctx,
		"INSERT INTO sessions (id, project_id, transcript_path, cwd, permission_mode, exit_reason, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
		sessionID, "proj-stop-test", "/tmp/transcript.jsonl", "/test/project", "default", "", time.Now().UTC().Format(time.RFC3339),
	)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	transcriptPath, err := filepath.Abs("testdata/transcript.jsonl")
	if err != nil {
		t.Fatalf("Failed to get transcript path: %v", err)
	}

	input := map[string]any{
		"session_id":       sessionID,
		"transcript_path":  transcriptPath,
		"cwd":              "/test/project",
		"permission_mode":  "default",
		"hook_event_name":  "Stop",
		"stop_hook_active": false,
	}

	_, err = runHookWithInput(t, input)
	if err != nil {
		t.Fatalf("Stop handler failed: %v", err)
	}

	// Verify session was updated (upserted)
	queries := sqlc.New(db)
	session, err := queries.GetSessionByID(ctx, sessionID)
	if err != nil {
		t.Fatalf("Failed to get session: %v", err)
	}

	assertEqual(t, "session.Cwd", "/test/project", session.Cwd)

	// Verify metrics exist
	metrics, err := queries.GetSessionMetricsBySessionID(ctx, sessionID)
	if err != nil {
		t.Fatalf("Failed to get metrics: %v", err)
	}

	if metrics.TokenInput <= 0 {
		t.Error("Expected positive token input after stop update")
	}
}
