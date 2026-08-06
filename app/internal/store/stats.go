package store

import "context"

type PlatformStats struct {
	TotalInstallations int
	TotalScans         int
	TotalEvilMerges    int
	ScansLast30Days    int
}

func (s *Store) PlatformStats(ctx context.Context) (*PlatformStats, error) {
	var stats PlatformStats
	if err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM installations`).Scan(&stats.TotalInstallations); err != nil {
		return nil, err
	}
	if err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM scans`).Scan(&stats.TotalScans); err != nil {
		return nil, err
	}
	// Each scan walks the whole branch history, so evil_merges is a cumulative
	// count as of that commit, not an incremental one — summing every scan row
	// would multiply-count the same repo on every re-scan. Only the latest scan
	// per repo reflects its current findings.
	if err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(evil_merges), 0) FROM (
			SELECT DISTINCT ON (installation_id, owner, repo) evil_merges
			FROM scans
			ORDER BY installation_id, owner, repo, scanned_at DESC
		) latest
	`).Scan(&stats.TotalEvilMerges); err != nil {
		return nil, err
	}
	if err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM scans WHERE scanned_at >= NOW() - INTERVAL '30 days'
	`).Scan(&stats.ScansLast30Days); err != nil {
		return nil, err
	}
	return &stats, nil
}
