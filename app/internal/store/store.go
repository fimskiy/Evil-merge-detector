package store

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Store provides access to the database.
type Store struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, databaseURL string) (*Store, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("connecting to database: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("pinging database: %w", err)
	}
	return &Store{pool: pool}, nil
}

func (s *Store) Close() {
	s.pool.Close()
}

// Migrate creates tables if they don't exist.
func (s *Store) Migrate(ctx context.Context) error {
	_, err := s.pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS installations (
			installation_id BIGINT PRIMARY KEY,
			account_login   TEXT NOT NULL,
			account_type    TEXT NOT NULL,
			plan            TEXT NOT NULL DEFAULT 'free',
			installed_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS scans (
			id              BIGSERIAL PRIMARY KEY,
			installation_id BIGINT NOT NULL,
			owner           TEXT NOT NULL,
			repo            TEXT NOT NULL,
			pr_number       INT,
			head_sha        TEXT NOT NULL,
			evil_merges     INT NOT NULL DEFAULT 0,
			total_merges    INT NOT NULL DEFAULT 0,
			duration_ms     BIGINT NOT NULL DEFAULT 0,
			scanned_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE INDEX IF NOT EXISTS scans_installation_id_idx ON scans (installation_id);
		CREATE INDEX IF NOT EXISTS scans_owner_repo_idx ON scans (owner, repo);

		ALTER TABLE installations
			ADD COLUMN IF NOT EXISTS last_full_scan_at TIMESTAMPTZ;

		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint WHERE conname = 'scans_installation_id_fk'
			) THEN
				DELETE FROM scans
					WHERE installation_id NOT IN (SELECT installation_id FROM installations);
				ALTER TABLE scans
					ADD CONSTRAINT scans_installation_id_fk
					FOREIGN KEY (installation_id) REFERENCES installations(installation_id)
					ON DELETE CASCADE;
			END IF;
		END
		$$;
	`)
	return err
}
