package parser

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/emiliopalmerini/claude-watcher/internal/domain"
)

type ParsedTranscript struct {
	StartedAt *time.Time
	EndedAt   *time.Time
	Metrics   *domain.SessionMetrics
	Tools     []*domain.SessionTool
	Files     []*domain.SessionFile
	Commands  []*domain.SessionCommand
}

type TranscriptEntry struct {
	Type      string          `json:"type"`
	Timestamp string          `json:"timestamp,omitempty"`
	Message   *Message        `json:"message,omitempty"`
	Usage     *Usage          `json:"usage,omitempty"`
	Result    json.RawMessage `json:"result,omitempty"`
}

type Message struct {
	Role    string    `json:"role"`
	Content []Content `json:"content"`
}

type Content struct {
	Type      string          `json:"type"`
	Text      string          `json:"text,omitempty"`
	ToolUseID string          `json:"id,omitempty"`
	Name      string          `json:"name,omitempty"`
	Input     json.RawMessage `json:"input,omitempty"`
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

func ParseTranscript(sessionID, path string) (*ParsedTranscript, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open transcript: %w", err)
	}
	defer file.Close()

	result := &ParsedTranscript{
		Metrics: &domain.SessionMetrics{
			SessionID: sessionID,
		},
		Tools:    make([]*domain.SessionTool, 0),
		Files:    make([]*domain.SessionFile, 0),
		Commands: make([]*domain.SessionCommand, 0),
	}

	toolCounts := make(map[string]*domain.SessionTool)
	fileCounts := make(map[string]*domain.SessionFile) // key: filepath:operation

	scanner := bufio.NewScanner(file)
	// Increase buffer size for large lines
	buf := make([]byte, 0, 1024*1024)
	scanner.Buffer(buf, 10*1024*1024)

	var firstTimestamp, lastTimestamp *time.Time

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
		case "assistant":
			result.Metrics.MessageCountAssistant++
			if entry.Message != nil {
				processAssistantMessage(entry.Message, sessionID, toolCounts, fileCounts, result)
			}
		case "result":
			// Tool results - check for errors
			processToolResult(entry, result)
		}

		// Accumulate token usage
		if entry.Usage != nil {
			result.Metrics.TokenInput += entry.Usage.InputTokens
			result.Metrics.TokenOutput += entry.Usage.OutputTokens
			result.Metrics.TokenCacheRead += entry.Usage.CacheReadInputTokens
			result.Metrics.TokenCacheWrite += entry.Usage.CacheCreationInputTokens
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading transcript: %w", err)
	}

	// Calculate turn count (pairs of user-assistant messages)
	result.Metrics.TurnCount = min(result.Metrics.MessageCountUser, result.Metrics.MessageCountAssistant)

	// Set timestamps
	result.StartedAt = firstTimestamp
	result.EndedAt = lastTimestamp

	// Convert maps to slices
	for _, tool := range toolCounts {
		result.Tools = append(result.Tools, tool)
	}
	for _, file := range fileCounts {
		result.Files = append(result.Files, file)
	}

	return result, nil
}

func processAssistantMessage(msg *Message, sessionID string, toolCounts map[string]*domain.SessionTool, fileCounts map[string]*domain.SessionFile, result *ParsedTranscript) {
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
