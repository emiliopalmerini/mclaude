package parser

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/emiliopalmerini/mclaude/internal/domain"
)

type ParsedTranscript struct {
	StartedAt *time.Time
	EndedAt   *time.Time
	ModelID   *string // Model used in the session (e.g., "claude-opus-4-5-20251101")
	Metrics   *domain.SessionMetrics
	Tools     []*domain.SessionTool
	Files     []*domain.SessionFile
	Commands  []*domain.SessionCommand
	Subagents []*domain.SessionSubagent
}

type TranscriptEntry struct {
	Type              string          `json:"type"`
	Timestamp         string          `json:"timestamp,omitempty"`
	Model             string          `json:"model,omitempty"`
	Message           *Message        `json:"message,omitempty"`
	Usage             *Usage          `json:"usage,omitempty"`
	Result            json.RawMessage `json:"result,omitempty"`
	ToolUseResultData json.RawMessage `json:"toolUseResult,omitempty"`
}

type Message struct {
	Role    string    `json:"role"`
	Content []Content `json:"content"`
	Usage   *Usage    `json:"usage,omitempty"`
}

type Content struct {
	Type         string          `json:"type"`
	Text         string          `json:"text,omitempty"`
	ToolUseID    string          `json:"id,omitempty"`
	ToolUseIDRef string          `json:"tool_use_id,omitempty"` // In tool_result entries
	Name         string          `json:"name,omitempty"`
	Input        json.RawMessage `json:"input,omitempty"`
}

type Usage struct {
	InputTokens              int64 `json:"input_tokens"`
	OutputTokens             int64 `json:"output_tokens"`
	CacheReadInputTokens     int64 `json:"cache_read_input_tokens"`
	CacheCreationInputTokens int64 `json:"cache_creation_input_tokens"`
}

type ToolInput struct {
	FilePath string `json:"file_path,omitempty"`
	Command  string `json:"command,omitempty"`
}

type ToolResult struct {
	ExitCode *int `json:"exit_code,omitempty"`
}

type SubagentToolInput struct {
	Description  string `json:"description,omitempty"`
	SubagentType string `json:"subagent_type,omitempty"`
	Model        string `json:"model,omitempty"`
	Prompt       string `json:"prompt,omitempty"`
	Skill        string `json:"skill,omitempty"`
	Args         string `json:"args,omitempty"`
}

type ToolUseResult struct {
	Status            string `json:"status,omitempty"`
	TotalDurationMs   *int64 `json:"totalDurationMs,omitempty"`
	TotalTokens       int64  `json:"totalTokens,omitempty"`
	TotalToolUseCount int64  `json:"totalToolUseCount,omitempty"`
	Usage             *Usage `json:"usage,omitempty"`
}

type pendingSubagent struct {
	agentType   string
	agentKind   string // "task" or "skill"
	description *string
	model       *string
}

func ParseTranscript(sessionID, path string) (*ParsedTranscript, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open transcript: %w", err)
	}
	defer func() { _ = file.Close() }()

	result := &ParsedTranscript{
		Metrics: &domain.SessionMetrics{
			SessionID: sessionID,
		},
		Tools:     make([]*domain.SessionTool, 0),
		Files:     make([]*domain.SessionFile, 0),
		Commands:  make([]*domain.SessionCommand, 0),
		Subagents: make([]*domain.SessionSubagent, 0),
	}

	toolCounts := make(map[string]*domain.SessionTool)
	fileCounts := make(map[string]*domain.SessionFile) // key: filepath:operation
	pendingSubagents := make(map[string]*pendingSubagent)

	scanner := bufio.NewScanner(file)
	// Increase buffer size for large lines
	buf := make([]byte, 0, 1024*1024)
	scanner.Buffer(buf, 10*1024*1024)

	var firstTimestamp, lastTimestamp *time.Time
	var modelID *string

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var entry TranscriptEntry
		if err := json.Unmarshal(line, &entry); err != nil {
			// Skip malformed lines
			continue
		}

		// Track timestamps
		if entry.Timestamp != "" {
			t, err := time.Parse(time.RFC3339Nano, entry.Timestamp)
			if err == nil {
				if firstTimestamp == nil {
					firstTimestamp = &t
				}
				lastTimestamp = &t
			}
		}

		// Process based on entry type
		switch entry.Type {
		case "user", "human":
			result.Metrics.MessageCountUser++
			// Check for toolUseResult (sub-agent completion data)
			if len(entry.ToolUseResultData) > 0 && entry.Message != nil {
				processSubagentResult(entry, sessionID, pendingSubagents, result)
			}
		case "assistant":
			result.Metrics.MessageCountAssistant++
			// Capture model ID from assistant messages (use first occurrence)
			if modelID == nil && entry.Model != "" {
				m := entry.Model
				modelID = &m
			}
			if entry.Message != nil {
				processAssistantMessage(entry.Message, sessionID, toolCounts, fileCounts, pendingSubagents, result)
			}
		case "result":
			// Tool results - check for errors
			processToolResult(entry, result)
		}

		// Accumulate token usage (check both top-level and message-level usage)
		usage := entry.Usage
		if usage == nil && entry.Message != nil {
			usage = entry.Message.Usage
		}
		if usage != nil {
			result.Metrics.TokenInput += usage.InputTokens
			result.Metrics.TokenOutput += usage.OutputTokens
			result.Metrics.TokenCacheRead += usage.CacheReadInputTokens
			result.Metrics.TokenCacheWrite += usage.CacheCreationInputTokens
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading transcript: %w", err)
	}

	// Calculate turn count (pairs of user-assistant messages)
	result.Metrics.TurnCount = min(result.Metrics.MessageCountUser, result.Metrics.MessageCountAssistant)

	// Set timestamps and model
	result.StartedAt = firstTimestamp
	result.EndedAt = lastTimestamp
	result.ModelID = modelID

	// Convert maps to slices
	for _, tool := range toolCounts {
		result.Tools = append(result.Tools, tool)
	}
	for _, file := range fileCounts {
		result.Files = append(result.Files, file)
	}

	return result, nil
}

func processAssistantMessage(msg *Message, sessionID string, toolCounts map[string]*domain.SessionTool, fileCounts map[string]*domain.SessionFile, pendingSubs map[string]*pendingSubagent, result *ParsedTranscript) {
	for _, content := range msg.Content {
		if content.Type != "tool_use" {
			continue
		}

		toolName := content.Name
		if toolName == "" {
			continue
		}

		// Track tool usage
		if existing, ok := toolCounts[toolName]; ok {
			existing.InvocationCount++
		} else {
			toolCounts[toolName] = &domain.SessionTool{
				SessionID:       sessionID,
				ToolName:        toolName,
				InvocationCount: 1,
			}
		}

		// Detect sub-agent invocations (Task or Skill tool_use)
		if (toolName == "Task" || toolName == "Skill") && len(content.Input) > 0 && content.ToolUseID != "" {
			var subInput SubagentToolInput
			if err := json.Unmarshal(content.Input, &subInput); err == nil {
				pending := &pendingSubagent{}
				if toolName == "Task" {
					pending.agentKind = "task"
					pending.agentType = subInput.SubagentType
					if pending.agentType == "" {
						pending.agentType = "unknown"
					}
					if subInput.Description != "" {
						desc := subInput.Description
						pending.description = &desc
					}
					if subInput.Model != "" {
						m := subInput.Model
						pending.model = &m
					}
				} else { // Skill
					pending.agentKind = "skill"
					pending.agentType = subInput.Skill
					if pending.agentType == "" {
						pending.agentType = "unknown"
					}
				}
				pendingSubs[content.ToolUseID] = pending
			}
		}

		// Parse tool input for file operations and bash commands
		if len(content.Input) > 0 {
			var input ToolInput
			if err := json.Unmarshal(content.Input, &input); err == nil {
				// Track file operations
				if input.FilePath != "" {
					operation := getFileOperation(toolName)
					if operation != "" {
						key := fmt.Sprintf("%s:%s", input.FilePath, operation)
						if existing, ok := fileCounts[key]; ok {
							existing.OperationCount++
						} else {
							fileCounts[key] = &domain.SessionFile{
								SessionID:      sessionID,
								FilePath:       input.FilePath,
								Operation:      operation,
								OperationCount: 1,
							}
						}
					}
				}

				// Track bash commands
				if input.Command != "" && toolName == "Bash" {
					result.Commands = append(result.Commands, &domain.SessionCommand{
						SessionID: sessionID,
						Command:   input.Command,
					})
				}
			}
		}
	}
}

func processSubagentResult(entry TranscriptEntry, sessionID string, pendingSubs map[string]*pendingSubagent, result *ParsedTranscript) {
	// Parse the toolUseResult
	var toolUseResult ToolUseResult
	if err := json.Unmarshal(entry.ToolUseResultData, &toolUseResult); err != nil {
		return
	}

	// Find matching tool_use_id from the user message content
	var matchedToolUseID string
	if entry.Message != nil {
		for _, content := range entry.Message.Content {
			if content.Type == "tool_result" && content.ToolUseIDRef != "" {
				if _, ok := pendingSubs[content.ToolUseIDRef]; ok {
					matchedToolUseID = content.ToolUseIDRef
					break
				}
			}
		}
	}

	// If no match found via content, try to match with the most recent pending sub-agent
	if matchedToolUseID == "" {
		for id := range pendingSubs {
			matchedToolUseID = id
			break
		}
	}

	if matchedToolUseID == "" {
		return
	}

	pending, ok := pendingSubs[matchedToolUseID]
	if !ok {
		return
	}

	subagent := &domain.SessionSubagent{
		SessionID:   sessionID,
		AgentType:   pending.agentType,
		AgentKind:   pending.agentKind,
		Description: pending.description,
		Model:       pending.model,
		TotalTokens: toolUseResult.TotalTokens,
		ToolUseCount: toolUseResult.TotalToolUseCount,
		TotalDurationMs: toolUseResult.TotalDurationMs,
	}

	if toolUseResult.Usage != nil {
		subagent.TokenInput = toolUseResult.Usage.InputTokens
		subagent.TokenOutput = toolUseResult.Usage.OutputTokens
		subagent.TokenCacheRead = toolUseResult.Usage.CacheReadInputTokens
		subagent.TokenCacheWrite = toolUseResult.Usage.CacheCreationInputTokens

		// Accumulate sub-agent tokens into session totals
		result.Metrics.TokenInput += toolUseResult.Usage.InputTokens
		result.Metrics.TokenOutput += toolUseResult.Usage.OutputTokens
		result.Metrics.TokenCacheRead += toolUseResult.Usage.CacheReadInputTokens
		result.Metrics.TokenCacheWrite += toolUseResult.Usage.CacheCreationInputTokens
	}

	result.Subagents = append(result.Subagents, subagent)

	// Remove from pending
	delete(pendingSubs, matchedToolUseID)
}

func processToolResult(entry TranscriptEntry, result *ParsedTranscript) {
	if len(entry.Result) == 0 {
		return
	}

	// Try to parse as an error result
	var resultData struct {
		IsError  bool `json:"is_error"`
		ExitCode *int `json:"exit_code,omitempty"`
	}
	if err := json.Unmarshal(entry.Result, &resultData); err == nil {
		if resultData.IsError {
			result.Metrics.ErrorCount++
		}
		// Update last bash command with exit code
		if resultData.ExitCode != nil && len(result.Commands) > 0 {
			lastCmd := result.Commands[len(result.Commands)-1]
			if lastCmd.ExitCode == nil {
				lastCmd.ExitCode = resultData.ExitCode
			}
		}
	}
}

func getFileOperation(toolName string) string {
	switch toolName {
	case "Read":
		return "read"
	case "Write":
		return "write"
	case "Edit":
		return "edit"
	default:
		return ""
	}
}
