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
	if err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*), COALESCE(SUM(evil_merges), 0) FROM scans
	`).Scan(&stats.TotalScans, &stats.TotalEvilMerges); err != nil {
		return nil, err
	}
	if err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM scans WHERE scanned_at >= NOW() - INTERVAL '30 days'
	`).Scan(&stats.ScansLast30Days); err != nil {
		return nil, err
	}
	return &stats, nil
}
