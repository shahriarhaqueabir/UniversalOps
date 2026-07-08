package aiops

import (
	"fmt"
	"strings"
	"time"
)

// EnhancedReport is a comprehensive diagnostics report.
type EnhancedReport struct {
	Title       string
	GeneratedAt time.Time
	Sections    []ReportSection
}

// GenerateEnhancedReport creates a comprehensive report with all available data.
// It populates sections from provided data strings keyed by section title.
func GenerateEnhancedReport(title string, data map[string]string) *EnhancedReport {
	report := &EnhancedReport{
		Title:       title,
		GeneratedAt: time.Now(),
	}

	if title == "" {
		report.Title = "Hawkward System Report"
	}

	for sectionTitle, content := range data {
		report.Sections = append(report.Sections, ReportSection{
			Title:   sectionTitle,
			Content: content,
		})
	}

	return report
}

// String returns a plain-text version of the enhanced report.
func (r *EnhancedReport) String() string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("=== %s ===\n", r.Title))
	b.WriteString(fmt.Sprintf("Generated: %s\n\n", r.GeneratedAt.Format("2006-01-02 15:04:05")))

	b.WriteString(GenerateReport(r.Sections))

	return b.String()
}

// Markdown returns a markdown-formatted enhanced report.
func (r *EnhancedReport) Markdown() string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("# %s\n\n", r.Title))
	b.WriteString(fmt.Sprintf("_Generated: %s_\n\n", r.GeneratedAt.Format("2006-01-02 15:04:05")))

	b.WriteString(ReportToMarkdown(r.Sections))

	return b.String()
}

// Save saves the report in the specified format (text or markdown) to the given path.
func (r *EnhancedReport) Save(path, format string) error {
	var content string
	switch format {
	case "markdown", "md":
		content = r.Markdown()
	default:
		content = r.String()
	}
	return SaveReport(content, path)
}

// OllamaReport sends sections to Ollama for AI-powered analysis and returns the response.
func (r *EnhancedReport) OllamaReport() (string, error) {
	if !checkOllamaAvailable() {
		return "", fmt.Errorf("ollama is not available")
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Analyze this system report: %s\n\n", r.Title))
	for _, section := range r.Sections {
		b.WriteString(fmt.Sprintf("=== %s ===\n%s\n", section.Title, section.Content))
	}

	messages := []ChatMessage{
		{Role: "user", Content: b.String()},
	}

	return Chat(messages)
}

// checkOllamaAvailable checks if Ollama is reachable (non-model helper).
func checkOllamaAvailable() bool {
	status, err := CheckOllama()
	if err != nil {
		return false
	}
	return status.Available
}
