package reporter

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/evilmerge-dev/evil-merge-detector/internal/models"
)

// SARIF 2.1.0 — https://docs.oasis-open.org/sarif/sarif/v2.1.0/sarif-v2.1.0.html
// GitHub Code Scanning accepts this format via upload-sarif action.

const sarifSchema = "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json"

// ruleSpec holds the static metadata for each rule.
type ruleSpec struct {
	ID               string
	ShortDescription string
	FullDescription  string
	HelpURI          string
	Level            string
}

var ruleSpecs = []ruleSpec{
	{
		ID:               "EMD001",
		ShortDescription: "File silently modified in merge commit",
		FullDescription:  "A file that was not changed in either parent branch was modified in the merge commit. This is a strong indicator of an evil merge — code injected during the merge that bypasses code review.",
		HelpURI:          "https://github.com/evilmerge-dev/Evil-merge-detector#severity-levels",
		Level:            "error",
	},
	{
		ID:               "EMD002",
		ShortDescription: "New file introduced only in merge commit",
		FullDescription:  "A file was added in the merge commit that did not exist in either parent branch. Files introduced this way are invisible to code reviewers.",
		HelpURI:          "https://github.com/evilmerge-dev/Evil-merge-detector#severity-levels",
		Level:            "error",
	},
	{
		ID:               "EMD003",
		ShortDescription: "File deleted only in merge commit",
		FullDescription:  "A file was deleted in the merge commit but existed in both parent branches. Deletions introduced this way bypass code review.",
		HelpURI:          "https://github.com/evilmerge-dev/Evil-merge-detector#severity-levels",
		Level:            "error",
	},
	{
		ID:               "EMD004",
		ShortDescription: "Sensitive file modified in merge commit",
		FullDescription:  "A file matching a sensitive pattern (credentials, keys, auth, crypto) was modified in the merge commit in a way not explained by either parent branch.",
		HelpURI:          "https://github.com/evilmerge-dev/Evil-merge-detector#severity-levels",
		Level:            "error",
	},
	{
		ID:               "EMD005",
		ShortDescription: "Merge result differs unexpectedly from both parents",
		FullDescription:  "A file was changed in one parent branch, but the merge result differs from both the changed parent and the base. This may indicate undisclosed modifications introduced during conflict resolution.",
		HelpURI:          "https://github.com/evilmerge-dev/Evil-merge-detector#severity-levels",
		Level:            "warning",
	},
	{
		ID:               "EMD006",
		ShortDescription: "Conflict zone contains extra changes",
		FullDescription:  "A file was modified in both parent branches (conflict zone), and the merge result contains content not present in either parent. This warrants review, though it may be legitimate conflict resolution.",
		HelpURI:          "https://github.com/evilmerge-dev/Evil-merge-detector#severity-levels",
		Level:            "note",
	},
}

// changeTypeToRuleID maps a ChangeType+Severity pair to a rule ID.
func changeTypeToRuleID(ec models.EvilChange) string {
	switch ec.ChangeType {
	case models.ChangeAdded:
		return "EMD002"
	case models.ChangeDeleted:
		return "EMD003"
	case models.ChangeSensitive:
		return "EMD004"
	case models.ChangeModified:
		switch ec.Severity {
		case models.SeverityCritical:
			return "EMD001"
		case models.SeverityWarning:
			return "EMD005"
		default:
			return "EMD006"
		}
	default:
		return "EMD001"
	}
}

func severityToSarifLevel(s models.Severity) string {
	switch s {
	case models.SeverityCritical:
		return "error"
	case models.SeverityWarning:
		return "warning"
	default:
		return "note"
	}
}

type sarifOutput struct {
	Schema  string     `json:"$schema"`
	Version string     `json:"version"`
	Runs    []sarifRun `json:"runs"`
}

type sarifRun struct {
	Tool    sarifTool     `json:"tool"`
	Results []sarifResult `json:"results"`
}

type sarifTool struct {
	Driver sarifDriver `json:"driver"`
}

type sarifDriver struct {
	Name           string          `json:"name"`
	Version        string          `json:"version"`
	InformationURI string          `json:"informationUri"`
	Rules          []sarifRuleJSON `json:"rules"`
}

// sarifRuleJSON is the JSON-serializable form of a rule.
type sarifRuleJSON struct {
	ID               string   `json:"id"`
	ShortDescription sarifMsg `json:"shortDescription"`
	FullDescription  sarifMsg `json:"fullDescription,omitempty"`
	HelpURI          string   `json:"helpUri,omitempty"`
}

type sarifMsg struct {
	Text string `json:"text"`
}

type sarifResult struct {
	RuleID    string          `json:"ruleId"`
	Level     string          `json:"level"`
	Message   sarifMsg        `json:"message"`
	Locations []sarifLocation `json:"locations"`
	PartialFingerprints map[string]string `json:"partialFingerprints,omitempty"`
}

type sarifLocation struct {
	PhysicalLocation sarifPhysicalLocation `json:"physicalLocation"`
}

type sarifPhysicalLocation struct {
	ArtifactLocation sarifArtifactLocation `json:"artifactLocation"`
}

type sarifArtifactLocation struct {
	URI       string `json:"uri"`
	URIBaseID string `json:"uriBaseId,omitempty"`
}

// SARIFReporter outputs SARIF 2.1.0 for GitHub Code Scanning.
type SARIFReporter struct {
	// ToolVersion is injected by the caller (matches binary version).
	ToolVersion string
}

func NewSARIF(version string) *SARIFReporter {
	return &SARIFReporter{ToolVersion: version}
}

func (r *SARIFReporter) Report(w io.Writer, result *models.ScanResult) error {
	rules := make([]sarifRuleJSON, len(ruleSpecs))
	for i, spec := range ruleSpecs {
		rules[i] = sarifRuleJSON{
			ID:               spec.ID,
			ShortDescription: sarifMsg{Text: spec.ShortDescription},
			FullDescription:  sarifMsg{Text: spec.FullDescription},
			HelpURI:          spec.HelpURI,
		}
	}

	var results []sarifResult
	for _, report := range result.Reports {
		for _, ec := range report.EvilChanges {
			ruleID := changeTypeToRuleID(ec)
			res := sarifResult{
				RuleID: ruleID,
				Level:  severityToSarifLevel(ec.Severity),
				Message: sarifMsg{
					Text: fmt.Sprintf("Evil merge detected in commit %s (%s): %s — %s",
						report.ShortHash,
						report.Author,
						ec.FilePath,
						ec.Detail,
					),
				},
				Locations: []sarifLocation{
					{
						PhysicalLocation: sarifPhysicalLocation{
							ArtifactLocation: sarifArtifactLocation{
								URI:       ec.FilePath,
								URIBaseID: "%SRCROOT%",
							},
						},
					},
				},
				PartialFingerprints: map[string]string{
					"commitHash/v1": report.CommitHash,
				},
			}
			results = append(results, res)
		}
	}

	if results == nil {
		results = []sarifResult{}
	}

	out := sarifOutput{
		Schema:  sarifSchema,
		Version: "2.1.0",
		Runs: []sarifRun{
			{
				Tool: sarifTool{
					Driver: sarifDriver{
						Name:           "Evil Merge Detector",
						Version:        r.ToolVersion,
						InformationURI: "https://github.com/evilmerge-dev/Evil-merge-detector",
						Rules:          rules,
					},
				},
				Results: results,
			},
		},
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		return fmt.Errorf("encoding SARIF: %w", err)
	}
	return nil
}
