package store

import (
	"context"
	"time"
)

type ScanRecord struct {
	InstallationID int64
	Owner          string
	Repo           string
	PRNumber       int
	HeadSHA        string
	EvilMerges     int
	TotalMerges    int
	DurationMs     int64
	ScannedAt      time.Time
}

func (s *Store) SaveScan(ctx context.Context, rec ScanRecord) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO scans
			(installation_id, owner, repo, pr_number, head_sha, evil_merges, total_merges, duration_ms)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, rec.InstallationID, rec.Owner, rec.Repo, rec.PRNumber,
		rec.HeadSHA, rec.EvilMerges, rec.TotalMerges, rec.DurationMs)
	return err
}

func (s *Store) MonthlyScansCount(ctx context.Context, installationID int64) (int, error) {
	var count int
	err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM scans
		WHERE installation_id = $1
		  AND scanned_at >= date_trunc('month', NOW())
	`, installationID).Scan(&count)
	return count, err
}

// LastScan returns the most recent scan record for the given repository, or nil if none exists.
func (s *Store) LastScan(ctx context.Context, owner, repo string) (*ScanRecord, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT installation_id, owner, repo, pr_number, head_sha,
		       evil_merges, total_merges, duration_ms, scanned_at
		FROM scans
		WHERE owner = $1 AND repo = $2
		ORDER BY scanned_at DESC
		LIMIT 1
	`, owner, repo)

	var rec ScanRecord
	err := row.Scan(&rec.InstallationID, &rec.Owner, &rec.Repo, &rec.PRNumber,
		&rec.HeadSHA, &rec.EvilMerges, &rec.TotalMerges, &rec.DurationMs, &rec.ScannedAt)
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

// ClaimRepoScan atomically claims the right to scan owner/repo, succeeding
// at most once per window. Used to rate-limit repos that push far more often
// than a full history scan can keep up with. The INSERT ... ON CONFLICT ...
// WHERE is a single statement, so Postgres serializes concurrent callers on
// the same row instead of racing a separate check-then-act.
func (s *Store) ClaimRepoScan(ctx context.Context, owner, repo string, window time.Duration) (bool, error) {
	tag, err := s.pool.Exec(ctx, `
		INSERT INTO repo_scan_claims (owner, repo, claimed_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (owner, repo) DO UPDATE
			SET claimed_at = NOW()
			WHERE repo_scan_claims.claimed_at < NOW() - $3::interval
	`, owner, repo, window)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

// DeleteRepoScanClaim removes a repo's scan claim. Used to clean up after tests
// and after the temporary /internal/verify-repo-scan-claim diagnostic.
func (s *Store) DeleteRepoScanClaim(ctx context.Context, owner, repo string) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM repo_scan_claims WHERE owner = $1 AND repo = $2`, owner, repo)
	return err
}

func (s *Store) RecentScans(ctx context.Context, installationID int64, limit int) ([]ScanRecord, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT installation_id, owner, repo, pr_number, head_sha,
		       evil_merges, total_merges, duration_ms, scanned_at
		FROM scans
		WHERE installation_id = $1
		ORDER BY scanned_at DESC
		LIMIT $2
	`, installationID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var scans []ScanRecord
	for rows.Next() {
		var rec ScanRecord
		if err := rows.Scan(&rec.InstallationID, &rec.Owner, &rec.Repo, &rec.PRNumber,
			&rec.HeadSHA, &rec.EvilMerges, &rec.TotalMerges, &rec.DurationMs, &rec.ScannedAt); err != nil {
			return nil, err
		}
		scans = append(scans, rec)
	}
	return scans, rows.Err()
}
