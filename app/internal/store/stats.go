package store

import "context"

type PlatformStats struct {
	TotalInstallations     int
	TotalScans             int
	TotalEvilMerges        int
	ScansLast30Days        int
	ActiveInstallations30d int
	NewInstallations7d     int
	NewInstallations30d    int
	EvilMergeRate          float64
	AvgScanDurationMs      float64
	P95ScanDurationMs      float64
	OrgInstallations       int
	UserInstallations      int
	PRScanRate             float64
	InstallationsNoScans   int
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
	var totalMerges int
	if err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(evil_merges), 0), COALESCE(SUM(total_merges), 0) FROM (
			SELECT DISTINCT ON (installation_id, owner, repo) evil_merges, total_merges
			FROM scans
			ORDER BY installation_id, owner, repo, scanned_at DESC
		) latest
	`).Scan(&stats.TotalEvilMerges, &totalMerges); err != nil {
		return nil, err
	}
	if totalMerges > 0 {
		stats.EvilMergeRate = float64(stats.TotalEvilMerges) / float64(totalMerges) * 100
	}
	if err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM scans WHERE scanned_at >= NOW() - INTERVAL '30 days'
	`).Scan(&stats.ScansLast30Days); err != nil {
		return nil, err
	}
	if err := s.pool.QueryRow(ctx, `
		SELECT COUNT(DISTINCT installation_id) FROM scans WHERE scanned_at >= NOW() - INTERVAL '30 days'
	`).Scan(&stats.ActiveInstallations30d); err != nil {
		return nil, err
	}
	if err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FILTER (WHERE installed_at >= NOW() - INTERVAL '7 days'),
		       COUNT(*) FILTER (WHERE installed_at >= NOW() - INTERVAL '30 days')
		FROM installations
	`).Scan(&stats.NewInstallations7d, &stats.NewInstallations30d); err != nil {
		return nil, err
	}
	if err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(AVG(duration_ms), 0)::float8,
		       COALESCE(percentile_cont(0.95) WITHIN GROUP (ORDER BY duration_ms), 0)::float8
		FROM scans
	`).Scan(&stats.AvgScanDurationMs, &stats.P95ScanDurationMs); err != nil {
		return nil, err
	}
	if err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FILTER (WHERE account_type = 'Organization'),
		       COUNT(*) FILTER (WHERE account_type != 'Organization')
		FROM installations
	`).Scan(&stats.OrgInstallations, &stats.UserInstallations); err != nil {
		return nil, err
	}
	var prScans int
	if err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FILTER (WHERE pr_number IS NOT NULL) FROM scans
	`).Scan(&prScans); err != nil {
		return nil, err
	}
	if stats.TotalScans > 0 {
		stats.PRScanRate = float64(prScans) / float64(stats.TotalScans) * 100
	}
	if err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM installations i
		WHERE NOT EXISTS (SELECT 1 FROM scans s WHERE s.installation_id = i.installation_id)
	`).Scan(&stats.InstallationsNoScans); err != nil {
		return nil, err
	}
	return &stats, nil
}
