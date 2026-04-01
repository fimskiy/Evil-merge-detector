package ghclient

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v84/github"

	"github.com/fimskiy/evil-merge-detector/internal/models"
)

const checkName = "Evil Merge Detector"

func CreateCheckRun(ctx context.Context, client *github.Client, owner, repo, headSHA string) (int64, error) {
	run, _, err := client.Checks.CreateCheckRun(ctx, owner, repo, github.CreateCheckRunOptions{
		Name:    checkName,
		HeadSHA: headSHA,
		Status:  github.Ptr("in_progress"),
	})
	if err != nil {
		return 0, fmt.Errorf("creating check run: %w", err)
	}
	return run.GetID(), nil
}

func UpdateCheckRun(ctx context.Context, client *github.Client, owner, repo string, runID int64, result *models.ScanResult) error {
	conclusion := "success"
	title := "No evil merges detected"
	summary := "All merge commits look clean."

	if result.EvilMerges > 0 {
		conclusion = "failure"
		title = fmt.Sprintf("Found %d evil merge(s)", result.EvilMerges)
		summary = buildSummary(result)
	}

	_, _, err := client.Checks.UpdateCheckRun(ctx, owner, repo, runID, github.UpdateCheckRunOptions{
		Name:       checkName,
		Status:     github.Ptr("completed"),
		Conclusion: github.Ptr(conclusion),
		Output: &github.CheckRunOutput{
			Title:   github.Ptr(title),
			Summary: github.Ptr(summary),
		},
	})
	if err != nil {
		return fmt.Errorf("updating check run: %w", err)
	}
	return nil
}

func FailCheckRun(ctx context.Context, client *github.Client, owner, repo string, runID int64, scanErr error) error {
	_, _, err := client.Checks.UpdateCheckRun(ctx, owner, repo, runID, github.UpdateCheckRunOptions{
		Name:       checkName,
		Status:     github.Ptr("completed"),
		Conclusion: github.Ptr("failure"),
		Output: &github.CheckRunOutput{
			Title:   github.Ptr("Scan failed"),
			Summary: github.Ptr(fmt.Sprintf("Evil Merge Detector encountered an error: %v", scanErr)),
		},
	})
	return err
}

func buildSummary(result *models.ScanResult) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Found **%d** evil merge commit(s) with **%d** suspicious change(s).\n\n",
		result.EvilMerges, countChanges(result))
	fmt.Fprintf(&sb, "| Severity | Commit | File | Detail |\n")
	fmt.Fprintf(&sb, "|----------|--------|------|--------|\n")
	for _, r := range result.Reports {
		for _, ec := range r.EvilChanges {
			sev := ec.Severity.String()
			fmt.Fprintf(&sb, "| %s | `%s` | `%s` | %s |\n",
				sev, r.ShortHash, ec.FilePath, ec.Detail)
		}
	}
	return sb.String()
}

func countChanges(result *models.ScanResult) int {
	n := 0
	for _, r := range result.Reports {
		n += len(r.EvilChanges)
	}
	return n
}
