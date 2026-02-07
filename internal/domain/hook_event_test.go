package domain

import (
	"testing"
)

func TestParseHookEvent_SessionEnd(t *testing.T) {
	input := []byte(`{
		"session_id": "abc123",
		"transcript_path": "/tmp/transcript.jsonl",
		"cwd": "/home/user/project",
		"permission_mode": "default",
		"hook_event_name": "SessionEnd",
		"reason": "exit"
	}`)

	event, err := ParseHookEvent(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	se, ok := event.(*SessionEndInput)
	if !ok {
		t.Fatalf("expected *SessionEndInput, got %T", event)
	}

	assertEqual(t, "SessionID", "abc123", se.SessionID)
	assertEqual(t, "TranscriptPath", "/tmp/transcript.jsonl", se.TranscriptPath)
	assertEqual(t, "Cwd", "/home/user/project", se.Cwd)
	assertEqual(t, "PermissionMode", "default", se.PermissionMode)
	assertEqual(t, "HookEventName", "SessionEnd", se.HookEventName)
	assertEqual(t, "Reason", "exit", se.Reason)
}

func TestParseHookEvent_SessionStart(t *testing.T) {
	input := []byte(`{
		"session_id": "abc123",
		"transcript_path": "/tmp/transcript.jsonl",
		"cwd": "/home/user/project",
		"permission_mode": "default",
		"hook_event_name": "SessionStart",
		"source": "cli",
		"model": "claude-opus-4-6",
		"agent_type": "main"
	}`)

	event, err := ParseHookEvent(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ss, ok := event.(*SessionStartInput)
	if !ok {
		t.Fatalf("expected *SessionStartInput, got %T", event)
	}

	assertEqual(t, "SessionID", "abc123", ss.SessionID)
	assertEqual(t, "HookEventName", "SessionStart", ss.HookEventName)
	assertEqual(t, "Source", "cli", ss.Source)
	assertEqual(t, "Model", "claude-opus-4-6", ss.Model)
	assertEqual(t, "AgentType", "main", ss.AgentType)
}

func TestParseHookEvent_PostToolUse(t *testing.T) {
	input := []byte(`{
		"session_id": "abc123",
		"transcript_path": "/tmp/transcript.jsonl",
		"cwd": "/home/user/project",
		"permission_mode": "default",
		"hook_event_name": "PostToolUse",
		"tool_name": "Bash",
		"tool_input": {"command": "go test ./..."},
		"tool_response": {"stdout": "PASS"},
		"tool_use_id": "tool_123"
	}`)

	event, err := ParseHookEvent(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ptu, ok := event.(*PostToolUseInput)
	if !ok {
		t.Fatalf("expected *PostToolUseInput, got %T", event)
	}

	assertEqual(t, "SessionID", "abc123", ptu.SessionID)
	assertEqual(t, "HookEventName", "PostToolUse", ptu.HookEventName)
	assertEqual(t, "ToolName", "Bash", ptu.ToolName)
	assertEqual(t, "ToolUseID", "tool_123", ptu.ToolUseID)

	if ptu.ToolInput == nil {
		t.Fatal("expected ToolInput to be set")
	}
	if ptu.ToolResponse == nil {
		t.Fatal("expected ToolResponse to be set")
	}
}

func TestParseHookEvent_Stop(t *testing.T) {
	input := []byte(`{
		"session_id": "abc123",
		"transcript_path": "/tmp/transcript.jsonl",
		"cwd": "/home/user/project",
		"permission_mode": "default",
		"hook_event_name": "Stop",
		"stop_hook_active": true
	}`)

	event, err := ParseHookEvent(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	s, ok := event.(*StopInput)
	if !ok {
		t.Fatalf("expected *StopInput, got %T", event)
	}

	assertEqual(t, "SessionID", "abc123", s.SessionID)
	assertEqual(t, "HookEventName", "Stop", s.HookEventName)
	assertEqual(t, "StopHookActive", true, s.StopHookActive)
}

func TestParseHookEvent_SubagentStart(t *testing.T) {
	input := []byte(`{
		"session_id": "abc123",
		"transcript_path": "/tmp/transcript.jsonl",
		"cwd": "/home/user/project",
		"permission_mode": "default",
		"hook_event_name": "SubagentStart",
		"agent_id": "agent_456",
		"agent_type": "Explore"
	}`)

	event, err := ParseHookEvent(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ss, ok := event.(*SubagentStartInput)
	if !ok {
		t.Fatalf("expected *SubagentStartInput, got %T", event)
	}

	assertEqual(t, "SessionID", "abc123", ss.SessionID)
	assertEqual(t, "HookEventName", "SubagentStart", ss.HookEventName)
	assertEqual(t, "AgentID", "agent_456", ss.AgentID)
	assertEqual(t, "AgentType", "Explore", ss.AgentType)
}

func TestParseHookEvent_SubagentStop(t *testing.T) {
	input := []byte(`{
		"session_id": "abc123",
		"transcript_path": "/tmp/transcript.jsonl",
		"cwd": "/home/user/project",
		"permission_mode": "default",
		"hook_event_name": "SubagentStop",
		"stop_hook_active": false,
		"agent_id": "agent_456",
		"agent_type": "Explore",
		"agent_transcript_path": "/tmp/subagent_transcript.jsonl"
	}`)

	event, err := ParseHookEvent(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ss, ok := event.(*SubagentStopInput)
	if !ok {
		t.Fatalf("expected *SubagentStopInput, got %T", event)
	}

	assertEqual(t, "SessionID", "abc123", ss.SessionID)
	assertEqual(t, "HookEventName", "SubagentStop", ss.HookEventName)
	assertEqual(t, "StopHookActive", false, ss.StopHookActive)
	assertEqual(t, "AgentID", "agent_456", ss.AgentID)
	assertEqual(t, "AgentType", "Explore", ss.AgentType)
	assertEqual(t, "AgentTranscriptPath", "/tmp/subagent_transcript.jsonl", ss.AgentTranscriptPath)
}

func TestParseHookEvent_Unknown(t *testing.T) {
	input := []byte(`{
		"session_id": "abc123",
		"hook_event_name": "SomeFutureEvent"
	}`)

	_, err := ParseHookEvent(input)
	if err == nil {
		t.Fatal("expected error for unknown event type")
	}
}

func TestParseHookEvent_InvalidJSON(t *testing.T) {
	input := []byte(`not json`)

	_, err := ParseHookEvent(input)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestParseHookEvent_MissingEventName(t *testing.T) {
	input := []byte(`{
		"session_id": "abc123"
	}`)

	_, err := ParseHookEvent(input)
	if err == nil {
		t.Fatal("expected error for missing event name")
	}
}

func assertEqual[T comparable](t *testing.T, name string, expected, actual T) {
	t.Helper()
	if expected != actual {
		t.Errorf("%s: expected %v, got %v", name, expected, actual)
	}
}
