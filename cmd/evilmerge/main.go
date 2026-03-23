package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fimskiy/evil-merge-detector/internal/models"
	"github.com/fimskiy/evil-merge-detector/internal/reporter"
	"github.com/fimskiy/evil-merge-detector/internal/scanner"
	"github.com/spf13/cobra"
)

var version = "dev"

func main() {
	var rootCmd = &cobra.Command{
		Use:   "evilmerge",
		Short: "Detect evil merges in Git repositories",
		Long: `Evil Merge Detector finds merge commits that contain changes
beyond conflict resolution — changes that weren't in either parent branch.

These "evil merges" bypass code review and can hide bugs or malicious code.`,
	}

	var (
		scanBranch   string
		scanSince    string
		scanUntil    string
		scanFormat   string
		scanSeverity string
		scanLimit    int
		scanFailOn   string
		scanCommit   string
		scanTimeout  time.Duration
	)

	var scanCmd = &cobra.Command{
		Use:   "scan [path]",
		Short: "Scan a repository for evil merges",
		Long:  `Scan analyzes merge commits in the repository and reports any that contain unexpected changes.`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoPath := ""
			if len(args) > 0 {
				repoPath = args[0]
			}

			// Build context: always respect Ctrl+C, optionally add timeout
			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer stop()

			if scanTimeout > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, scanTimeout)
				defer cancel()
			}

			// Print header (skip for machine-readable formats)
			if scanFormat != "json" && scanFormat != "sarif" {
				if _, err := fmt.Fprintf(os.Stdout, "Evil Merge Detector %s\n", version); err != nil {
					return err
				}
			}

			s := scanner.New()

			// --commit mode: detailed single-commit inspection
			if scanCommit != "" {
				report, err := s.InspectCommit(ctx, repoPath, scanCommit)
				if err != nil {
					return err
				}
				return reporter.PrintDetail(os.Stdout, report)
			}

			opts := models.ScanOptions{
				RepoPath:    repoPath,
				Branch:      scanBranch,
				Limit:       scanLimit,
				MinSeverity: parseSeverity(scanSeverity),
			}

			if scanSince != "" {
				t, err := time.Parse("2006-01-02", scanSince)
				if err != nil {
					return fmt.Errorf("invalid --since date (use YYYY-MM-DD): %w", err)
				}
				opts.Since = &t
			}

			if scanUntil != "" {
				t, err := time.Parse("2006-01-02", scanUntil)
				if err != nil {
					return fmt.Errorf("invalid --until date (use YYYY-MM-DD): %w", err)
				}
				opts.Until = &t
			}

			failOnSeverity := parseSeverity(scanFailOn)

			result, err := s.Scan(ctx, opts)
			if err != nil {
				return err
			}

			var rep reporter.Reporter
			switch scanFormat {
			case "json":
				rep = reporter.NewJSON(true)
			case "sarif":
				rep = reporter.NewSARIF(version)
			default:
				rep = reporter.NewText()
			}

			if err := rep.Report(os.Stdout, result); err != nil {
				return err
			}

			// Exit with non-zero code if evil merges found above threshold
			if scanFailOn != "" {
				for _, r := range result.Reports {
					if r.MaxSeverity >= failOnSeverity {
						os.Exit(1)
					}
				}
			}

			return nil
		},
	}

	scanCmd.Flags().StringVar(&scanBranch, "branch", "", "Branch to scan (default: current HEAD)")
	scanCmd.Flags().StringVar(&scanSince, "since", "", "Scan commits after this date (YYYY-MM-DD)")
	scanCmd.Flags().StringVar(&scanUntil, "until", "", "Scan commits before this date (YYYY-MM-DD)")
	scanCmd.Flags().StringVar(&scanFormat, "format", "text", "Output format: text, json, sarif")
	scanCmd.Flags().StringVar(&scanSeverity, "severity", "", "Minimum severity to report: info, warning, critical")
	scanCmd.Flags().IntVar(&scanLimit, "limit", 0, "Maximum number of merge commits to analyze (0 = unlimited)")
	scanCmd.Flags().StringVar(&scanFailOn, "fail-on", "", "Exit with code 1 if evil merges found at or above this severity")
	scanCmd.Flags().StringVar(&scanCommit, "commit", "", "Inspect a specific merge commit in detail (by full or short hash)")
	scanCmd.Flags().DurationVar(&scanTimeout, "timeout", 0, "Maximum scan duration, e.g. 30s, 5m (0 = unlimited)")

	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Evil Merge Detector %s\n", version)
		},
	}

	rootCmd.AddCommand(scanCmd, versionCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func parseSeverity(s string) models.Severity {
	switch s {
	case "critical":
		return models.SeverityCritical
	case "warning":
		return models.SeverityWarning
	case "info":
		return models.SeverityInfo
	default:
		return models.SeverityInfo
	}
}
