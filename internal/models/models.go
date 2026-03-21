package models

import "time"

// Severity represents the severity level of an evil merge finding.
type Severity int

const (
	SeverityInfo     Severity = iota // Changes in conflicted files (likely legitimate)
	SeverityWarning                  // Unexpected changes in non-conflicted files
	SeverityCritical                 // New files or changes in sensitive patterns
)

func (s Severity) String() string {
	switch s {
	case SeverityInfo:
		return "INFO"
	case SeverityWarning:
		return "WARNING"
	case SeverityCritical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

// ChangeType describes how a file was modified in an evil merge.
type ChangeType int

const (
	ChangeModified    ChangeType = iota // File content differs from expected
	ChangeAdded                         // File added only in merge commit
	ChangeDeleted                       // File deleted only in merge commit
	ChangeSensitive                     // Change in a sensitive file pattern
)

func (c ChangeType) String() string {
	switch c {
	case ChangeModified:
		return "modified"
	case ChangeAdded:
		return "added"
	case ChangeDeleted:
		return "deleted"
	case ChangeSensitive:
		return "sensitive"
	default:
		return "unknown"
	}
}

// EvilChange represents a single suspicious change within a merge commit.
type EvilChange struct {
	FilePath   string     `json:"file_path"`
	ChangeType ChangeType `json:"change_type"`
	Severity   Severity   `json:"severity"`
	Detail     string     `json:"detail"`
	Diff       string     `json:"diff,omitempty"` // populated only with --commit flag
}

// MergeReport represents analysis results for a single merge commit.
type MergeReport struct {
	CommitHash    string       `json:"commit_hash"`
	ShortHash     string       `json:"short_hash"`
	Message       string       `json:"message"`
	Author        string       `json:"author"`
	AuthorEmail   string       `json:"author_email"`
	Date          time.Time    `json:"date"`
	ParentHashes  []string     `json:"parent_hashes"`
	EvilChanges   []EvilChange `json:"evil_changes"`
	MaxSeverity   Severity     `json:"max_severity"`
}

// ScanResult holds the complete scan output.
type ScanResult struct {
	RepoPath       string        `json:"repo_path"`
	Branch         string        `json:"branch"`
	TotalMerges    int           `json:"total_merges"`
	EvilMerges     int           `json:"evil_merges"`
	Reports        []MergeReport `json:"reports"`
	ScanDuration   time.Duration `json:"scan_duration"`
}

// ScanOptions configures the scanning behavior.
type ScanOptions struct {
	RepoPath    string
	Branch      string
	Since       *time.Time
	Until       *time.Time
	Limit       int
	MinSeverity Severity
}
