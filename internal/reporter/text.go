package reporter

import (
	"fmt"
	"io"
	"strings"

	"github.com/evilmerge-dev/evil-merge-detector/internal/models"
	"github.com/fatih/color"
)

// TextReporter outputs human-readable colored table output.
type TextReporter struct{}

func NewText() *TextReporter {
	return &TextReporter{}
}

func (r *TextReporter) Report(w io.Writer, result *models.ScanResult) error {
	if _, err := fmt.Fprintf(w, "Scanning repository: %s", result.RepoPath); err != nil {
		return err
	}
	if result.Branch != "" {
		if _, err := fmt.Fprintf(w, " (branch: %s)", result.Branch); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(w, "Analyzed %d merge commits, found %d evil merges (in %s)\n\n",
		result.TotalMerges, result.EvilMerges, result.ScanDuration.Round(1e6)); err != nil {
		return err
	}

	if len(result.Reports) == 0 {
		green := color.New(color.FgGreen).SprintFunc()
		if _, err := fmt.Fprintln(w, green("No evil merges detected!")); err != nil {
			return err
		}
		return nil
	}

	header := fmt.Sprintf("%-10s  %-50s  %-25s  %s", "SEVERITY", "COMMIT", "AUTHOR", "FILES")
	if _, err := fmt.Fprintln(w, header); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, strings.Repeat("-", len(header)+10)); err != nil {
		return err
	}

	for _, report := range result.Reports {
		msg := strings.Split(report.Message, "\n")[0]
		if len(msg) > 40 {
			msg = msg[:37] + "..."
		}

		commitInfo := fmt.Sprintf("%s %s", report.ShortHash, msg)
		if len(commitInfo) > 50 {
			commitInfo = commitInfo[:47] + "..."
		}

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

		if _, err := fmt.Fprintf(w, "%-10s  %-50s  %-25s  %s\n", sevStr, commitInfo, author, fileStr); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintf(w, "\nRe-run with --format=json for full details on each merge.\n"); err != nil {
		return err
	}
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
