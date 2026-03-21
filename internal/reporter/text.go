package reporter

import (
	"fmt"
	"io"
	"strings"

	"github.com/fimskiy/evil-merge-detector/internal/models"
	"github.com/fatih/color"
)

// TextReporter outputs human-readable colored table output.
type TextReporter struct{}

// NewText creates a new TextReporter.
func NewText() *TextReporter {
	return &TextReporter{}
}

func (r *TextReporter) Report(w io.Writer, result *models.ScanResult) error {
	fmt.Fprintf(w, "Scanning repository: %s", result.RepoPath)
	if result.Branch != "" {
		fmt.Fprintf(w, " (branch: %s)", result.Branch)
	}
	fmt.Fprintln(w)

	fmt.Fprintf(w, "Analyzed %d merge commits, found %d evil merges (in %s)\n\n",
		result.TotalMerges, result.EvilMerges, result.ScanDuration.Round(1e6))

	if len(result.Reports) == 0 {
		green := color.New(color.FgGreen).SprintFunc()
		fmt.Fprintln(w, green("No evil merges detected!"))
		return nil
	}

	// Print table header
	header := fmt.Sprintf("%-10s  %-50s  %-25s  %s", "SEVERITY", "COMMIT", "AUTHOR", "FILES")
	fmt.Fprintln(w, header)
	fmt.Fprintln(w, strings.Repeat("-", len(header)+10))

	for _, report := range result.Reports {
		// Truncate message to first line, max 40 chars
		msg := strings.Split(report.Message, "\n")[0]
		if len(msg) > 40 {
			msg = msg[:37] + "..."
		}

		commitInfo := fmt.Sprintf("%s %s", report.ShortHash, msg)
		if len(commitInfo) > 50 {
			commitInfo = commitInfo[:47] + "..."
		}

		// Collect affected file names
		var files []string
		for _, ec := range report.EvilChanges {
			files = append(files, ec.FilePath)
		}
		fileStr := strings.Join(files, ", ")
		if len(fileStr) > 30 {
			fileStr = fileStr[:27] + "..."
		}

		author := report.AuthorEmail
		if len(author) > 25 {
			author = author[:22] + "..."
		}

		sevStr := colorSeverity(report.MaxSeverity)

		fmt.Fprintf(w, "%-10s  %-50s  %-25s  %s\n", sevStr, commitInfo, author, fileStr)
	}

	fmt.Fprintf(w, "\nRe-run with --format=json for full details on each merge.\n")
	return nil
}

func colorSeverity(s models.Severity) string {
	switch s {
	case models.SeverityCritical:
		return color.New(color.FgRed, color.Bold).Sprint("CRITICAL")
	case models.SeverityWarning:
		return color.New(color.FgYellow).Sprint("WARNING")
	case models.SeverityInfo:
		return color.New(color.FgCyan).Sprint("INFO")
	default:
		return s.String()
	}
}
