package transcript

import (
	"bufio"
	"claude-watcher/internal/tracker/domain"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

// Parser implements the TranscriptParser port for JSONL transcripts
type Parser struct{}

// NewParser creates a new transcript parser
func NewParser() *Parser {
	return &Parser{}
}

// Parse parses a Claude Code transcript file and extracts statistics
func (p *Parser) Parse(transcriptPath string) (domain.Statistics, error) {
	stats := domain.NewStatistics()

	if transcriptPath == "" || !fileExists(transcriptPath) {
		return stats, nil
	}

	file, err := os.Open(transcriptPath)
	if err != nil {
		return stats, fmt.Errorf("open transcript: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	const maxCapacity = 1024 * 1024 // 1MB buffer for long lines
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	filesAccessedSet := make(map[string]bool)
	filesModifiedSet := make(map[string]bool)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var entry Entry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue // Skip malformed entries
		}

		p.processEntry(&entry, &stats, filesAccessedSet, filesModifiedSet)
	}

	if err := scanner.Err(); err != nil {
		return stats, fmt.Errorf("scan transcript: %w", err)
	}

	// Convert sets to slices
	for file := range filesAccessedSet {
		stats.FilesAccessed = append(stats.FilesAccessed, file)
	}
	for file := range filesModifiedSet {
		stats.FilesModified = append(stats.FilesModified, file)
	}

	return stats, nil
}

func (p *Parser) processEntry(
	entry *Entry,
	stats *domain.Statistics,
	filesAccessed, filesModified map[string]bool,
) {
	// Track timestamps
	p.updateTimestamps(entry, stats)

	// Extract metadata
	p.extractMetadata(entry, stats)

	// Process by entry type
	switch entry.Type {
	case "human", "user":
		p.processUserMessage(entry, stats)
	case "assistant":
		p.processAssistantMessage(entry, stats, filesAccessed, filesModified)
	case "tool_use":
		p.processToolUse(entry, stats)
	case "tool_result":
		p.processToolResult(entry, stats)
	case "system":
		p.processSystemMessage(entry, stats)
	}
}

func (p *Parser) updateTimestamps(entry *Entry, stats *domain.Statistics) {
	if entry.Timestamp == "" {
		return
	}

	t, err := time.Parse(time.RFC3339, entry.Timestamp)
	if err != nil {
		return
	}

	if stats.StartTime == nil {
		stats.StartTime = &t
	}
	stats.EndTime = &t
}

func (p *Parser) extractMetadata(entry *Entry, stats *domain.Statistics) {
	if stats.GitBranch == "" && entry.GitBranch != "" {
		stats.GitBranch = entry.GitBranch
	}
	if stats.ClaudeVersion == "" && entry.Version != "" {
		stats.ClaudeVersion = entry.Version
	}
}

func (p *Parser) processUserMessage(entry *Entry, stats *domain.Statistics) {
	stats.UserPrompts++

	// Capture first prompt as summary
	if stats.Summary != "" {
		return
	}

	var msg Message
	if err := json.Unmarshal(entry.Message, &msg); err != nil {
		return
	}

	var content string
	if err := json.Unmarshal(msg.Content, &content); err != nil {
		return
	}

	if content != "" {
		if len(content) > 200 {
			stats.Summary = content[:200]
		} else {
			stats.Summary = content
		}
	}
}

func (p *Parser) processAssistantMessage(
	entry *Entry,
	stats *domain.Statistics,
	filesAccessed, filesModified map[string]bool,
) {
	stats.AssistantResponses++

	var msg Message
	if err := json.Unmarshal(entry.Message, &msg); err != nil {
		return
	}

	// Extract token usage
	stats.InputTokens += msg.Usage.InputTokens
	stats.OutputTokens += msg.Usage.OutputTokens
	stats.CacheReadTokens += msg.Usage.CacheReadInputTokens
	stats.CacheWriteTokens += msg.Usage.CacheCreationInputTokens
	stats.ThinkingTokens += msg.Usage.ThinkingTokens

	// Get model
	if stats.Model == "" {
		if msg.Model != "" {
			stats.Model = msg.Model
		} else if entry.Model != "" {
			stats.Model = entry.Model
		}
	}

	// Process tool uses in content
	var contentItems []ContentItem
	if err := json.Unmarshal(msg.Content, &contentItems); err != nil {
		return
	}

	for _, item := range contentItems {
		if item.Type == "tool_use" {
			p.recordToolUse(item.Name, item.Input, stats, filesAccessed, filesModified)
		}
	}
}

func (p *Parser) processToolUse(entry *Entry, stats *domain.Statistics) {
	toolName := entry.Name
	if toolName == "" {
		toolName = "unknown"
	}
	stats.ToolCalls++
	stats.ToolsBreakdown[toolName]++
}

func (p *Parser) processToolResult(entry *Entry, stats *domain.Statistics) {
	if entry.IsError {
		stats.ErrorsCount++
		return
	}

	var content string
	if err := json.Unmarshal(entry.Content, &content); err != nil {
		return
	}

	// Check first 100 chars for error indicators
	if len(content) > 100 {
		content = content[:100]
	}
	if strings.Contains(strings.ToLower(content), "error") {
		stats.ErrorsCount++
	}
}

func (p *Parser) recordToolUse(
	toolName string,
	inputRaw json.RawMessage,
	stats *domain.Statistics,
	filesAccessed, filesModified map[string]bool,
) {
	if toolName == "" {
		toolName = "unknown"
	}

	stats.ToolCalls++
	stats.ToolsBreakdown[toolName]++

	// Track file access for file-related tools
	if !isFileTool(toolName) {
		return
	}

	var input ToolInput
	if err := json.Unmarshal(inputRaw, &input); err != nil {
		return
	}

	filePath := input.FilePath
	if filePath == "" {
		filePath = input.Path
	}
	if filePath == "" {
		filePath = input.NotebookPath
	}

	if filePath == "" {
		return
	}

	filesAccessed[filePath] = true
	if isModifyingTool(toolName) {
		filesModified[filePath] = true
	}
}

func (p *Parser) processSystemMessage(entry *Entry, stats *domain.Statistics) {
	// Look for rate limit or usage limit messages
	if entry.Subtype == "api_error" || entry.Level == "error" {
		// Try to extract content as string
		var content string
		if err := json.Unmarshal(entry.Content, &content); err == nil {
			if strings.Contains(strings.ToLower(content), "limit") ||
				strings.Contains(strings.ToLower(content), "rate") {
				stats.LimitMessage = content
			}
		}
	}

	// Also check the content field directly for limit messages
	if entry.Content != nil {
		var content string
		if err := json.Unmarshal(entry.Content, &content); err == nil {
			if strings.Contains(content, "hit your limit") ||
				strings.Contains(content, "resets") {
				stats.LimitMessage = content
			}
		}
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func isFileTool(toolName string) bool {
	fileTools := map[string]bool{
		"Read":         true,
		"Edit":         true,
		"Write":        true,
		"Glob":         true,
		"Grep":         true,
		"LSP":          true,
		"NotebookEdit": true,
	}
	return fileTools[toolName]
}

func isModifyingTool(toolName string) bool {
	modifyingTools := map[string]bool{
		"Edit":         true,
		"Write":        true,
		"NotebookEdit": true,
	}
	return modifyingTools[toolName]
}
