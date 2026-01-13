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
	// Extensive realistic tags matching real development workflows
	tags := []domain.Tag{
		// Task Type - What kind of work was done
		{Name: "bug_fix", Category: "task_type", Color: "#EF4444"},
		{Name: "new_feature", Category: "task_type", Color: "#22C55E"},
		{Name: "refactoring", Category: "task_type", Color: "#3B82F6"},
		{Name: "code_review", Category: "task_type", Color: "#8B5CF6"},
		{Name: "documentation", Category: "task_type", Color: "#F97316"},
		{Name: "testing", Category: "task_type", Color: "#EC4899"},
		{Name: "performance", Category: "task_type", Color: "#14B8A6"},
		{Name: "security", Category: "task_type", Color: "#EF4444"},
		{Name: "dependency_update", Category: "task_type", Color: "#6366F1"},
		{Name: "debugging", Category: "task_type", Color: "#F59E0B"},

		// Architecture - Which part of the system
		{Name: "frontend", Category: "architecture", Color: "#06B6D4"},
		{Name: "backend", Category: "architecture", Color: "#8B5CF6"},
		{Name: "database", Category: "architecture", Color: "#EC4899"},
		{Name: "api", Category: "architecture", Color: "#22C55E"},
		{Name: "cli", Category: "architecture", Color: "#F97316"},
		{Name: "infrastructure", Category: "architecture", Color: "#84CC16"},
		{Name: "devops", Category: "architecture", Color: "#EAB308"},
		{Name: "full_stack", Category: "architecture", Color: "#A855F7"},

		// Prompt Style - How you interacted with Claude
		{Name: "detailed_spec", Category: "prompt_style", Color: "#6366F1"},
		{Name: "minimal_prompt", Category: "prompt_style", Color: "#14B8A6"},
		{Name: "iterative", Category: "prompt_style", Color: "#F59E0B"},
		{Name: "pair_programming", Category: "prompt_style", Color: "#22C55E"},
		{Name: "code_generation", Category: "prompt_style", Color: "#3B82F6"},
		{Name: "explanation", Category: "prompt_style", Color: "#8B5CF6"},
		{Name: "troubleshooting", Category: "prompt_style", Color: "#EF4444"},

		// Outcome - How did the session go
		{Name: "success", Category: "outcome", Color: "#22C55E"},
		{Name: "partial_success", Category: "outcome", Color: "#EAB308"},
		{Name: "needs_revision", Category: "outcome", Color: "#F97316"},
		{Name: "blocked", Category: "outcome", Color: "#EF4444"},
		{Name: "learning", Category: "outcome", Color: "#6366F1"},
		{Name: "exploration", Category: "outcome", Color: "#14B8A6"},
	}

	p := prompter.NewBubbleTeaPrompter(testLogger{})

	fmt.Println()
	fmt.Println(strings.Repeat("─", 50))
	fmt.Println("  Claude Watcher - TUI Prompter Test")
	fmt.Println(strings.Repeat("─", 50))
	fmt.Println()
	fmt.Println("  This test launches the quality feedback TUI")
	fmt.Println("  with realistic tag categories:")
	fmt.Println()
	fmt.Printf("    • Task Type:    %d options\n", countByCategory(tags, "task_type"))
	fmt.Printf("    • Architecture: %d options\n", countByCategory(tags, "architecture"))
	fmt.Printf("    • Prompt Style: %d options\n", countByCategory(tags, "prompt_style"))
	fmt.Printf("    • Outcome:      %d options\n", countByCategory(tags, "outcome"))
	fmt.Println()
	fmt.Println("  Vim-style Controls:")
	fmt.Println("    j/k           Navigate up/down")
	fmt.Println("    h/l           Previous/next step")
	fmt.Println("    Space         Toggle selection")
	fmt.Println("    Enter         Confirm and proceed")
	fmt.Println("    1-5           Quick rating select")
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
		fmt.Println("  Selected Tags:")
		for _, tag := range data.Tags {
			fmt.Printf("    • %s\n", tag)
		}
	} else {
		fmt.Println("  Selected Tags: (none)")
	}

	fmt.Println()
	if data.Rating != nil {
		fmt.Printf("  Rating: %s\n", renderRating(*data.Rating))
	} else {
		fmt.Println("  Rating: (skipped)")
	}

	fmt.Println()
	if data.Notes != "" {
		fmt.Printf("  Notes: %s\n", data.Notes)
	} else {
		fmt.Println("  Notes: (empty)")
	}

	fmt.Println()
	fmt.Println(strings.Repeat("─", 50))
}

func countByCategory(tags []domain.Tag, category string) int {
	count := 0
	for _, t := range tags {
		if t.Category == category {
			count++
		}
	}
	return count
}

func renderRating(rating int) string {
	return fmt.Sprintf("%d/5", rating)
}
