package store

import (
	"context"
	"time"
)

type Installation struct {
	InstallationID int64
	AccountLogin   string
	AccountType    string
	Plan           string
	InstalledAt    time.Time
	UpdatedAt      time.Time
}

func (s *Store) UpsertInstallation(ctx context.Context, inst Installation) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO installations (installation_id, account_login, account_type, plan)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (installation_id) DO UPDATE
		SET account_login = EXCLUDED.account_login,
		    plan          = EXCLUDED.plan,
		    updated_at    = NOW()
	`, inst.InstallationID, inst.AccountLogin, inst.AccountType, inst.Plan)
	return err
}

func (s *Store) GetInstallation(ctx context.Context, installationID int64) (*Installation, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT installation_id, account_login, account_type, plan, installed_at, updated_at
		FROM installations WHERE installation_id = $1
	`, installationID)

	var inst Installation
	err := row.Scan(&inst.InstallationID, &inst.AccountLogin, &inst.AccountType,
		&inst.Plan, &inst.InstalledAt, &inst.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &inst, nil
}

func (s *Store) GetInstallationByLogin(ctx context.Context, login string) (*Installation, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT installation_id, account_login, account_type, plan, installed_at, updated_at
		FROM installations WHERE account_login = $1
		ORDER BY installed_at DESC LIMIT 1
	`, login)

	var inst Installation
	err := row.Scan(&inst.InstallationID, &inst.AccountLogin, &inst.AccountType,
		&inst.Plan, &inst.InstalledAt, &inst.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &inst, nil
}

func (s *Store) UpdatePlan(ctx context.Context, installationID int64, plan string) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE installations SET plan = $1, updated_at = NOW()
		WHERE installation_id = $2
	`, plan, installationID)
	return err
}

func (s *Store) DeleteInstallation(ctx context.Context, installationID int64) error {
	_, err := s.pool.Exec(ctx, `
		DELETE FROM installations WHERE installation_id = $1
	`, installationID)
	return err
}
