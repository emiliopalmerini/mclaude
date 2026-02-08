package domain

import (
	"encoding/json"
	"fmt"
)

// HookEventBase contains fields common to all hook events from Claude Code.
type HookEventBase struct {
	SessionID      string `json:"session_id"`
	TranscriptPath string `json:"transcript_path"`
	Cwd            string `json:"cwd"`
	PermissionMode string `json:"permission_mode"`
	HookEventName  string `json:"hook_event_name"`
}

// SessionEndInput is sent when a Claude Code session ends.
type SessionEndInput struct {
	HookEventBase
	Reason string `json:"reason"`
}

// SessionStartInput is sent when a Claude Code session starts.
type SessionStartInput struct {
	HookEventBase
	Source    string `json:"source"`
	Model     string `json:"model"`
	AgentType string `json:"agent_type"`
}

// PostToolUseInput is sent after a tool is used in a session.
type PostToolUseInput struct {
	HookEventBase
	ToolName     string          `json:"tool_name"`
	ToolInput    json.RawMessage `json:"tool_input"`
	ToolResponse json.RawMessage `json:"tool_response"`
	ToolUseID    string          `json:"tool_use_id"`
}

// StopInput is sent when a stop event occurs in a session.
type StopInput struct {
	HookEventBase
	StopHookActive bool `json:"stop_hook_active"`
}

// SubagentStartInput is sent when a sub-agent starts.
type SubagentStartInput struct {
	HookEventBase
	AgentID   string `json:"agent_id"`
	AgentType string `json:"agent_type"`
}

// SubagentStopInput is sent when a sub-agent stops.
type SubagentStopInput struct {
	HookEventBase
	StopHookActive     bool   `json:"stop_hook_active"`
	AgentID            string `json:"agent_id"`
	AgentType          string `json:"agent_type"`
	AgentTranscriptPath string `json:"agent_transcript_path"`
}

// ParseHookEvent parses raw JSON into the appropriate typed event struct.
func ParseHookEvent(data []byte) (any, error) {
	var base HookEventBase
	if err := json.Unmarshal(data, &base); err != nil {
		return nil, fmt.Errorf("failed to parse hook event: %w", err)
	}

	if base.HookEventName == "" {
		return nil, fmt.Errorf("missing hook_event_name")
	}

	switch base.HookEventName {
	case "SessionEnd":
		var event SessionEndInput
		if err := json.Unmarshal(data, &event); err != nil {
			return nil, fmt.Errorf("failed to parse SessionEnd event: %w", err)
		}
		return &event, nil

	case "SessionStart":
		var event SessionStartInput
		if err := json.Unmarshal(data, &event); err != nil {
			return nil, fmt.Errorf("failed to parse SessionStart event: %w", err)
		}
		return &event, nil

	case "PostToolUse":
		var event PostToolUseInput
		if err := json.Unmarshal(data, &event); err != nil {
			return nil, fmt.Errorf("failed to parse PostToolUse event: %w", err)
		}
		return &event, nil

	case "Stop":
		var event StopInput
		if err := json.Unmarshal(data, &event); err != nil {
			return nil, fmt.Errorf("failed to parse Stop event: %w", err)
		}
		return &event, nil

	case "SubagentStart":
		var event SubagentStartInput
		if err := json.Unmarshal(data, &event); err != nil {
			return nil, fmt.Errorf("failed to parse SubagentStart event: %w", err)
		}
		return &event, nil

	case "SubagentStop":
		var event SubagentStopInput
		if err := json.Unmarshal(data, &event); err != nil {
			return nil, fmt.Errorf("failed to parse SubagentStop event: %w", err)
		}
		return &event, nil

	default:
		return nil, fmt.Errorf("unknown hook event: %s", base.HookEventName)
	}
}
