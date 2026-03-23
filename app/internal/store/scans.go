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
