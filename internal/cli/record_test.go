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

	"github.com/emiliopalmerini/claude-watcher/internal/adapters/turso"
	"github.com/emiliopalmerini/claude-watcher/internal/domain"
	sqlc "github.com/emiliopalmerini/claude-watcher/sqlc/generated"
)

func TestRecordCommand_Integration(t *testing.T) {
	// Skip if not in integration test mode
	if os.Getenv("CLAUDE_WATCHER_DATABASE_URL") == "" {
		t.Skip("Skipping integration test: CLAUDE_WATCHER_DATABASE_URL not set")
	}

	ctx := context.Background()

	// Connect to database
	db, err := turso.NewDB()
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	queries := sqlc.New(db)

	// Generate unique session ID for this test
	sessionID := "test-record-" + randomID()

	// Get path to test transcript
	transcriptPath, err := filepath.Abs("testdata/transcript.jsonl")
	if err != nil {
		t.Fatalf("Failed to get transcript path: %v", err)
	}

	// Verify transcript exists
	if _, err := os.Stat(transcriptPath); os.IsNotExist(err) {
		t.Fatalf("Test transcript not found at %s", transcriptPath)
	}

	// Create hook input
	hookInput := domain.HookInput{
		SessionID:      sessionID,
		TranscriptPath: transcriptPath,
		Cwd:            "/test/project",
		PermissionMode: "default",
		HookEventName:  "SessionEnd",
		Reason:         "exit",
	}

	inputJSON, err := json.Marshal(hookInput)
	if err != nil {
		t.Fatalf("Failed to marshal hook input: %v", err)
	}

	// Save original stdin and restore after test
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	// Create a pipe to simulate stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	os.Stdin = r

	// Write input and close writer
	go func() {
		w.Write(inputJSON)
		w.Close()
	}()

	// Capture stdout
	oldStdout := os.Stdout
	rOut, wOut, _ := os.Pipe()
	os.Stdout = wOut

	// Run the record command
	err = runRecord(nil, nil)

	// Restore stdout and read output
	wOut.Close()
	os.Stdout = oldStdout
	var stdout bytes.Buffer
	stdout.ReadFrom(rOut)

	if err != nil {
		t.Fatalf("Record command failed: %v\nOutput: %s", err, stdout.String())
	}

	t.Logf("Record output: %s", stdout.String())

	// Cleanup function
	defer func() {
		// Delete test data
		queries.DeleteSession(ctx, sessionID)
		// Note: CASCADE should handle metrics, tools, files, commands
	}()

	// Verify session was created
	session, err := queries.GetSessionByID(ctx, sessionID)
	if err != nil {
		t.Fatalf("Failed to get session: %v", err)
	}

	// Verify session fields
	assertEqual(t, "session.ExitReason", "exit", session.ExitReason)
	assertEqual(t, "session.PermissionMode", "default", session.PermissionMode)
	assertEqual(t, "session.Cwd", "/test/project", session.Cwd)

	// Verify metrics
	metrics, err := queries.GetSessionMetricsBySessionID(ctx, sessionID)
	if err != nil {
		t.Fatalf("Failed to get metrics: %v", err)
	}

	// Expected: 2 user messages, 4 assistant messages
	assertEqual(t, "metrics.MessageCountUser", int64(2), metrics.MessageCountUser)
	assertEqual(t, "metrics.MessageCountAssistant", int64(4), metrics.MessageCountAssistant)

	// Expected tokens: 100+150+50+30 = 330 input, 50+75+25+15 = 165 output
	assertEqual(t, "metrics.TokenInput", int64(330), metrics.TokenInput)
	assertEqual(t, "metrics.TokenOutput", int64(165), metrics.TokenOutput)

	// Expected cache tokens: 20+30+10+5 = 65 read, 10+15+5+3 = 33 write
	assertEqual(t, "metrics.TokenCacheRead", int64(65), metrics.TokenCacheRead)
	assertEqual(t, "metrics.TokenCacheWrite", int64(33), metrics.TokenCacheWrite)

	// Verify cost estimate exists
	if !metrics.CostEstimateUsd.Valid {
		t.Error("Expected cost estimate to be set")
	}

	// Verify tools
	tools, err := queries.ListSessionToolsBySessionID(ctx, sessionID)
	if err != nil {
		t.Fatalf("Failed to get tools: %v", err)
	}

	toolMap := make(map[string]int64)
	for _, tool := range tools {
		toolMap[tool.ToolName] = tool.InvocationCount
	}

	assertEqual(t, "tool.Read count", int64(1), toolMap["Read"])
	assertEqual(t, "tool.Edit count", int64(1), toolMap["Edit"])
	assertEqual(t, "tool.Bash count", int64(1), toolMap["Bash"])

	// Verify files
	files, err := queries.ListSessionFilesBySessionID(ctx, sessionID)
	if err != nil {
		t.Fatalf("Failed to get files: %v", err)
	}

	fileMap := make(map[string]string)
	for _, file := range files {
		fileMap[file.FilePath+":"+file.Operation] = file.Operation
	}

	if _, ok := fileMap["/test/file.go:read"]; !ok {
		t.Error("Expected file /test/file.go:read to be tracked")
	}
	if _, ok := fileMap["/test/file.go:edit"]; !ok {
		t.Error("Expected file /test/file.go:edit to be tracked")
	}

	// Verify commands
	commands, err := queries.ListSessionCommandsBySessionID(ctx, sessionID)
	if err != nil {
		t.Fatalf("Failed to get commands: %v", err)
	}

	if len(commands) != 1 {
		t.Fatalf("Expected 1 command, got %d", len(commands))
	}
	assertEqual(t, "command.Command", "go build ./...", commands[0].Command)

	// Verify transcript was stored
	storedPath := session.TranscriptStoredPath
	if !storedPath.Valid {
		t.Error("Expected transcript stored path to be set")
	} else {
		if _, err := os.Stat(storedPath.String); os.IsNotExist(err) {
			t.Errorf("Stored transcript not found at %s", storedPath.String)
		}
		// Cleanup stored transcript
		defer os.Remove(storedPath.String)
	}

	t.Log("All assertions passed!")
}

func assertEqual[T comparable](t *testing.T, name string, expected, actual T) {
	t.Helper()
	if expected != actual {
		t.Errorf("%s: expected %v, got %v", name, expected, actual)
	}
}

func randomID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
