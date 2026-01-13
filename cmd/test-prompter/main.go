package main

import (
	"claude-watcher/internal/tracker/adapters/prompter"
	"claude-watcher/internal/tracker/domain"
	"fmt"
	"log"
	"strings"
)

// Simple logger for testing
type testLogger struct{}

func (l testLogger) Debug(msg string) {}
func (l testLogger) Error(msg string) { fmt.Println("[ERROR]", msg) }

func main() {
	// Task type tags only (other categories replaced by scales)
	tags := []domain.Tag{
		{Name: "feature", Category: "task_type", Color: "#22C55E"},
		{Name: "bugfix", Category: "task_type", Color: "#EF4444"},
		{Name: "refactor", Category: "task_type", Color: "#3B82F6"},
		{Name: "exploration", Category: "task_type", Color: "#8B5CF6"},
		{Name: "docs", Category: "task_type", Color: "#F59E0B"},
		{Name: "test", Category: "task_type", Color: "#06B6D4"},
		{Name: "config", Category: "task_type", Color: "#64748B"},
	}

	p := prompter.NewBubbleTeaPrompter(testLogger{})

	fmt.Println()
	fmt.Println(strings.Repeat("─", 50))
	fmt.Println("  Claude Watcher - TUI Prompter Test")
	fmt.Println(strings.Repeat("─", 50))
	fmt.Println()
	fmt.Println("  This test launches the quality feedback TUI")
	fmt.Println()
	fmt.Println("  Steps:")
	fmt.Printf("    1. Task Type (multi-select, %d options)\n", len(tags))
	fmt.Println("    2. Prompt Specificity (1-5 scale)")
	fmt.Println("    3. Task Completion (1-5 scale)")
	fmt.Println("    4. Code Confidence (1-5 scale)")
	fmt.Println("    5. Session Satisfaction (1-5 scale)")
	fmt.Println("    6. Notes (optional)")
	fmt.Println()
	fmt.Println("  Vim-style Controls:")
	fmt.Println("    j/k           Navigate tags")
	fmt.Println("    h/l           Scale value / prev-next step")
	fmt.Println("    Space         Toggle tag / confirm scale")
	fmt.Println("    Enter         Next step")
	fmt.Println("    1-5           Quick scale select")
	fmt.Println("    i             Insert mode (notes)")
	fmt.Println("    Esc           Normal mode / quit")
	fmt.Println("    q             Quit")
	fmt.Println()
	fmt.Println(strings.Repeat("─", 50))
	fmt.Println()
	fmt.Println("  Press Enter to start the TUI...")
	fmt.Scanln()

	data, err := p.CollectQualityData(tags)
	if err != nil {
		log.Fatal(err)
	}

	// Display results
	fmt.Println()
	fmt.Println(strings.Repeat("─", 50))
	fmt.Println("  Results")
	fmt.Println(strings.Repeat("─", 50))
	fmt.Println()

	if len(data.Tags) > 0 {
		fmt.Println("  Task Types:")
		for _, tag := range data.Tags {
			fmt.Printf("    • %s\n", tag)
		}
	} else {
		fmt.Println("  Task Types: (none)")
	}

	fmt.Println()
	fmt.Println("  Scales:")
	fmt.Printf("    Prompt Specificity:   %s\n", formatScale(data.PromptSpecificity))
	fmt.Printf("    Task Completion:      %s\n", formatScale(data.TaskCompletion))
	fmt.Printf("    Code Confidence:      %s\n", formatScale(data.CodeConfidence))
	fmt.Printf("    Session Satisfaction: %s\n", formatScale(data.Rating))

	fmt.Println()
	if data.Notes != "" {
		fmt.Printf("  Notes: %s\n", data.Notes)
	} else {
		fmt.Println("  Notes: (empty)")
	}

	fmt.Println()
	fmt.Println(strings.Repeat("─", 50))
}

func formatScale(val *int) string {
	if val == nil {
		return "(skipped)"
	}
	return fmt.Sprintf("%d/5", *val)
}
