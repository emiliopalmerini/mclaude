package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/emiliopalmerini/mclaude/internal/domain"
)

const maxToolResponseSize = 10 * 1024 // 10KB

func handlePostToolUse(event *domain.PostToolUseInput) error {
	sqlDB, tursoDB, closeDB, err := hookDB()
	if err != nil {
		return err
	}
	defer func() { syncAndClose(tursoDB, closeDB) }()

	ctx := context.Background()

	toolInput := string(event.ToolInput)
	toolResponse := truncateString(string(event.ToolResponse), maxToolResponseSize)

	// Compact JSON if it's valid JSON (reduces storage)
	if compacted, err := compactJSON(event.ToolInput); err == nil {
		toolInput = compacted
	}
	if compacted, err := compactJSON(json.RawMessage(toolResponse)); err == nil {
		toolResponse = compacted
	}

	_, err = sqlDB.ExecContext(ctx,
		"INSERT OR IGNORE INTO tool_events (session_id, tool_name, tool_use_id, tool_input, tool_response, captured_at) VALUES (?, ?, ?, ?, ?, ?)",
		event.SessionID, event.ToolName, event.ToolUseID, toolInput, toolResponse, time.Now().UTC().Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("failed to insert tool event: %w", err)
	}

	return nil
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "...[truncated]"
}

func compactJSON(data json.RawMessage) (string, error) {
	if len(data) == 0 {
		return "", fmt.Errorf("empty data")
	}
	var buf json.RawMessage
	if err := json.Unmarshal(data, &buf); err != nil {
		return "", err
	}
	compacted, err := json.Marshal(buf)
	if err != nil {
		return "", err
	}
	return string(compacted), nil
}
