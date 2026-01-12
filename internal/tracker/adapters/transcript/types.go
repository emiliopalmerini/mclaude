package transcript

import "encoding/json"

// Entry represents a single line in the JSONL transcript
type Entry struct {
	Type      string          `json:"type"`
	Timestamp string          `json:"timestamp"`
	GitBranch string          `json:"gitBranch"`
	Version   string          `json:"version"`
	Model     string          `json:"model"`
	Message   json.RawMessage `json:"message"`
	Name      string          `json:"name"`
	IsError   bool            `json:"is_error"`
	Content   json.RawMessage `json:"content"`
}

// Message represents the message field in transcript entries
type Message struct {
	Usage   Usage           `json:"usage"`
	Model   string          `json:"model"`
	Content json.RawMessage `json:"content"`
}

// Usage represents token usage in a message
type Usage struct {
	InputTokens              int `json:"input_tokens"`
	OutputTokens             int `json:"output_tokens"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
	ThinkingTokens           int `json:"thinking_tokens"`
}

// ContentItem represents a content item in assistant messages
type ContentItem struct {
	Type  string          `json:"type"`
	Name  string          `json:"name"`
	Input json.RawMessage `json:"input"`
}

// ToolInput represents generic tool input with file paths
type ToolInput struct {
	FilePath     string `json:"file_path"`
	Path         string `json:"path"`
	NotebookPath string `json:"notebook_path"`
}
