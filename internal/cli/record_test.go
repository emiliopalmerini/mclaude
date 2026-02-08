package cli

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/emiliopalmerini/mclaude/internal/domain"
	sqlc "github.com/emiliopalmerini/mclaude/sqlc/generated"
)

// TestRecordCommand_Memory runs the record command test with in-memory SQLite.
// This is fast and runs by default.
func TestRecordCommand_Memory(t *testing.T) {
	db, cleanup := testDB(t)
	defer cleanup()

	runRecordTest(t, db)
}

// TestRecordCommand_Turso runs the record command test with a Turso container.
// This is slower and only runs when MCLAUDE_TEST_TURSO=1.
func TestRecordCommand_Turso(t *testing.T) {
	if os.Getenv("MCLAUDE_TEST_TURSO") != "1" {
		t.Skip("Skipping Turso integration test: set MCLAUDE_TEST_TURSO=1 to run")
	}

	db, cleanup := testTursoDB(t)
	defer cleanup()

	runRecordTest(t, db)
}

// runRecordTest contains the actual test logic, parameterized by database.
func runRecordTest(t *testing.T, db *sql.DB) {
	t.Helper()

	// Override the database connection for the record command
	testDBOverride = db
	defer func() { testDBOverride = nil }()

	ctx := context.Background()
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

	// Expected: 3 user messages (2 original + 1 sub-agent result), 5 assistant messages (4 original + 1 Task tool_use)
	assertEqual(t, "metrics.MessageCountUser", int64(3), metrics.MessageCountUser)
	assertEqual(t, "metrics.MessageCountAssistant", int64(5), metrics.MessageCountAssistant)

	// Expected tokens: parent (100+150+50+40+30=370) + subagent (6000) = 6370 input
	// parent (50+75+25+20+15=185) + subagent (230) = 415 output
	assertEqual(t, "metrics.TokenInput", int64(6370), metrics.TokenInput)
	assertEqual(t, "metrics.TokenOutput", int64(415), metrics.TokenOutput)

	// Expected cache tokens: parent (20+30+10+8+5=73) + subagent (1500) = 1573 read
	// parent (10+15+5+4+3=37) + subagent (249) = 286 write
	assertEqual(t, "metrics.TokenCacheRead", int64(1573), metrics.TokenCacheRead)
	assertEqual(t, "metrics.TokenCacheWrite", int64(286), metrics.TokenCacheWrite)

	// Verify cost estimate exists (pricing was seeded by migrations)
	if !metrics.CostEstimateUsd.Valid {
		t.Error("Expected cost estimate to be set")
	} else {
		// Verify cost is reasonable (should be > 0)
		if metrics.CostEstimateUsd.Float64 <= 0 {
			t.Errorf("Expected positive cost estimate, got %f", metrics.CostEstimateUsd.Float64)
		}
		t.Logf("Cost estimate: $%.6f", metrics.CostEstimateUsd.Float64)
	}

	// Verify model ID was extracted
	if !metrics.ModelID.Valid {
		t.Error("Expected model ID to be set")
	} else {
		assertEqual(t, "metrics.ModelID", "claude-sonnet-4-20250514", metrics.ModelID.String)
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
	assertEqual(t, "tool.Task count", int64(1), toolMap["Task"])

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

	// Verify sub-agents
	subagents, err := queries.ListSessionSubagentsBySessionID(ctx, sessionID)
	if err != nil {
		t.Fatalf("Failed to get subagents: %v", err)
	}

	if len(subagents) != 1 {
		t.Fatalf("Expected 1 subagent, got %d", len(subagents))
	}
	assertEqual(t, "subagent.AgentType", "Explore", subagents[0].AgentType)
	assertEqual(t, "subagent.AgentKind", "task", subagents[0].AgentKind)
	assertEqual(t, "subagent.TotalTokens", int64(7979), subagents[0].TotalTokens)
	assertEqual(t, "subagent.TokenInput", int64(6000), subagents[0].TokenInput)
	assertEqual(t, "subagent.TokenOutput", int64(230), subagents[0].TokenOutput)
	assertEqual(t, "subagent.TokenCacheRead", int64(1500), subagents[0].TokenCacheRead)
	assertEqual(t, "subagent.TokenCacheWrite", int64(249), subagents[0].TokenCacheWrite)
	assertEqual(t, "subagent.ToolUseCount", int64(2), subagents[0].ToolUseCount)
	if !subagents[0].TotalDurationMs.Valid {
		t.Error("Expected subagent duration to be set")
	} else {
		assertEqual(t, "subagent.TotalDurationMs", int64(5000), subagents[0].TotalDurationMs.Int64)
	}
	if !subagents[0].Description.Valid {
		t.Error("Expected subagent description to be set")
	} else {
		assertEqual(t, "subagent.Description", "Search codebase", subagents[0].Description.String)
	}
	if !subagents[0].Model.Valid {
		t.Error("Expected subagent model to be set")
	} else {
		assertEqual(t, "subagent.Model", "haiku", subagents[0].Model.String)
	}

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
