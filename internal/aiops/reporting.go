package aiops

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// ReportSection holds a single section of a report.
type ReportSection struct {
	Title   string
	Content string
}

// GenerateReport combines sections into a formatted text report.
func GenerateReport(sections []ReportSection) string {
	var b strings.Builder
	for i, section := range sections {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(fmt.Sprintf("=== %s ===\n", section.Title))
		b.WriteString(section.Content)
		if !strings.HasSuffix(section.Content, "\n") {
			b.WriteString("\n")
		}
	}
	return b.String()
}

// ReportToMarkdown generates a markdown report from the given sections.
func ReportToMarkdown(sections []ReportSection) string {
	var b strings.Builder
	for i, section := range sections {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(fmt.Sprintf("## %s\n\n", section.Title))
		b.WriteString(section.Content)
		if !strings.HasSuffix(section.Content, "\n") {
			b.WriteString("\n")
		}
	}
	return b.String()
}

// ReportToJSON generates a pretty-printed JSON report from the given sections.
func ReportToJSON(sections []ReportSection) (string, error) {
	data, err := json.MarshalIndent(sections, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal report json: %w", err)
	}
	return string(data) + "\n", nil
}

// ReportToCSV generates a CSV report from the given sections.
func ReportToCSV(sections []ReportSection) (string, error) {
	var b bytes.Buffer
	writer := csv.NewWriter(&b)

	if err := writer.Write([]string{"title", "content"}); err != nil {
		return "", fmt.Errorf("write csv header: %w", err)
	}
	for _, section := range sections {
		if err := writer.Write([]string{section.Title, section.Content}); err != nil {
			return "", fmt.Errorf("write csv row: %w", err)
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", fmt.Errorf("flush csv report: %w", err)
	}

	return b.String(), nil
}

// SaveReport writes report content to a file.
func SaveReport(content, path string) error {
	return os.WriteFile(path, []byte(content), 0644)
}

// addSection adds a report section from user input.
// Input format: "Title|Content" or "Title" for a section with just a title.
func addSection(sections []ReportSection, input string) []ReportSection {
	parts := strings.SplitN(input, "|", 2)
	title := strings.TrimSpace(parts[0])
	content := ""
	if len(parts) == 2 {
		content = strings.TrimSpace(parts[1])
	}
	if title == "" {
		return sections
	}
	return append(sections, ReportSection{Title: title, Content: content})
}
