package reporter

import (
	"fmt"
	"io"
	"strings"

	"github.com/fatih/color"
	"github.com/fimskiy/evil-merge-detector/internal/models"
)

// PrintDetail writes a detailed report for a single merge commit to w.
// Used by the --commit flag.
func PrintDetail(w io.Writer, report *models.MergeReport) error {
	sep := strings.Repeat("─", 60)

	if _, err := fmt.Fprintf(w, "\nCommit:   %s\n", report.CommitHash); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Author:   %s <%s>\n", report.Author, report.AuthorEmail); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Date:     %s\n", report.Date.Format("2006-01-02 15:04:05")); err != nil {
		return err
	}

	msg := strings.Split(strings.TrimSpace(report.Message), "\n")[0]
	if _, err := fmt.Fprintf(w, "Message:  %s\n", msg); err != nil {
		return err
	}

	if len(report.ParentHashes) >= 2 {
		if _, err := fmt.Fprintf(w, "Parents:  %s  %s\n",
			report.ParentHashes[0][:7], report.ParentHashes[1][:7]); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}

	if len(report.EvilChanges) == 0 {
		green := color.New(color.FgGreen).SprintFunc()
		if _, err := fmt.Fprintln(w, green("No evil changes detected — this looks like a clean merge.")); err != nil {
			return err
		}
		return nil
	}

	count := len(report.EvilChanges)
	noun := "change"
	if count != 1 {
		noun = "changes"
	}
	if _, err := fmt.Fprintf(w, "Found %d evil %s:\n\n", count, noun); err != nil {
		return err
	}

	for _, ec := range report.EvilChanges {
		sevStr := colorSeverity(ec.Severity)
		if _, err := fmt.Fprintf(w, "%s\n", sep); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "[%s] %s\n", sevStr, ec.FilePath); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "  %s (%s)\n", ec.Detail, ec.ChangeType); err != nil {
			return err
		}

		if ec.Diff != "" {
			if _, err := fmt.Fprintln(w); err != nil {
				return err
			}
			for _, line := range strings.Split(strings.TrimSuffix(ec.Diff, "\n"), "\n") {
				colored := colorizeDiffLine(line)
				if _, err := fmt.Fprintln(w, colored); err != nil {
					return err
				}
			}
		}

		if _, err := fmt.Fprintln(w); err != nil {
			return err
		}
	}

	return nil
}

// colorizeDiffLine applies color to a diff line based on its prefix.
func colorizeDiffLine(line string) string {
	if strings.HasPrefix(line, "+") {
		return color.New(color.FgGreen).Sprint(line)
	}
	if strings.HasPrefix(line, "-") {
		return color.New(color.FgRed).Sprint(line)
	}
	return line
}
