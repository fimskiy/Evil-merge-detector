package store_test

import (
	"context"
	"os"
	"testing"

	"github.com/fimskiy/evil-merge-detector/app/internal/store"
)

func testDB(t *testing.T) *store.Store {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	db, err := store.New(context.Background(), url)
	if err != nil {
		t.Fatalf("store.New: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	if err := db.Migrate(context.Background()); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	return db
}

// --- installations ---

func TestUpsertAndGetInstallation(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()

	const id = int64(8_000_001)
	t.Cleanup(func() { db.DeleteInstallation(ctx, id) })

	inst := store.Installation{
		InstallationID: id,
		AccountLogin:   "testorg",
		AccountType:    "Organization",
		Plan:           "free",
	}
	if err := db.UpsertInstallation(ctx, inst); err != nil {
		t.Fatalf("UpsertInstallation: %v", err)
	}

	got, err := db.GetInstallation(ctx, id)
	if err != nil {
		t.Fatalf("GetInstallation: %v", err)
	}
	if got.AccountLogin != "testorg" {
		t.Errorf("login %q, want testorg", got.AccountLogin)
	}
	if got.Plan != "free" {
		t.Errorf("plan %q, want free", got.Plan)
	}
}

func TestUpsertInstallation_UpdatesOnConflict(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()

	const id = int64(8_000_002)
	t.Cleanup(func() { db.DeleteInstallation(ctx, id) })

	base := store.Installation{InstallationID: id, AccountLogin: "old", AccountType: "User", Plan: "free"}
	if err := db.UpsertInstallation(ctx, base); err != nil {
		t.Fatalf("first upsert: %v", err)
	}

	updated := store.Installation{InstallationID: id, AccountLogin: "new", AccountType: "User", Plan: "pro"}
	if err := db.UpsertInstallation(ctx, updated); err != nil {
		t.Fatalf("second upsert: %v", err)
	}

	got, err := db.GetInstallation(ctx, id)
	if err != nil {
		t.Fatalf("GetInstallation: %v", err)
	}
	if got.AccountLogin != "new" {
		t.Errorf("login %q, want new", got.AccountLogin)
	}
	if got.Plan != "pro" {
		t.Errorf("plan %q, want pro", got.Plan)
	}
}

func TestGetInstallationByLogin(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()

	const id = int64(8_000_003)
	t.Cleanup(func() { db.DeleteInstallation(ctx, id) })

	if err := db.UpsertInstallation(ctx, store.Installation{
		InstallationID: id,
		AccountLogin:   "uniquelogin",
		AccountType:    "User",
		Plan:           "free",
	}); err != nil {
		t.Fatalf("UpsertInstallation: %v", err)
	}

	got, err := db.GetInstallationByLogin(ctx, "uniquelogin")
	if err != nil {
		t.Fatalf("GetInstallationByLogin: %v", err)
	}
	if got.InstallationID != id {
		t.Errorf("id %d, want %d", got.InstallationID, id)
	}
}

func TestUpdatePlan(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()

	const id = int64(8_000_004)
	t.Cleanup(func() { db.DeleteInstallation(ctx, id) })

	if err := db.UpsertInstallation(ctx, store.Installation{
		InstallationID: id, AccountLogin: "u", AccountType: "User", Plan: "free",
	}); err != nil {
		t.Fatalf("UpsertInstallation: %v", err)
	}
	if err := db.UpdatePlan(ctx, id, "pro"); err != nil {
		t.Fatalf("UpdatePlan: %v", err)
	}
	got, _ := db.GetInstallation(ctx, id)
	if got.Plan != "pro" {
		t.Errorf("plan %q, want pro", got.Plan)
	}
}

func TestDeleteInstallation(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()

	const id = int64(8_000_005)
	if err := db.UpsertInstallation(ctx, store.Installation{
		InstallationID: id, AccountLogin: "u", AccountType: "User", Plan: "free",
	}); err != nil {
		t.Fatalf("UpsertInstallation: %v", err)
	}
	if err := db.DeleteInstallation(ctx, id); err != nil {
		t.Fatalf("DeleteInstallation: %v", err)
	}
	_, err := db.GetInstallation(ctx, id)
	if err == nil {
		t.Error("expected error after deletion, got nil")
	}
}

func TestMarkFullScannedAndListPending(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()

	const id = int64(8_000_006)
	t.Cleanup(func() { db.DeleteInstallation(ctx, id) })

	if err := db.UpsertInstallation(ctx, store.Installation{
		InstallationID: id, AccountLogin: "u", AccountType: "User", Plan: "free",
	}); err != nil {
		t.Fatalf("UpsertInstallation: %v", err)
	}

	// Should appear in pending list (last_full_scan_at IS NULL).
	list, err := db.ListPendingFullScans(ctx)
	if err != nil {
		t.Fatalf("ListPendingFullScans: %v", err)
	}
	found := false
	for _, inst := range list {
		if inst.InstallationID == id {
			found = true
		}
	}
	if !found {
		t.Error("installation should appear in pending list before MarkFullScanned")
	}

	if err := db.MarkFullScanned(ctx, id); err != nil {
		t.Fatalf("MarkFullScanned: %v", err)
	}

	// Should no longer appear in pending list.
	list, err = db.ListPendingFullScans(ctx)
	if err != nil {
		t.Fatalf("ListPendingFullScans after mark: %v", err)
	}
	for _, inst := range list {
		if inst.InstallationID == id {
			t.Error("installation should not appear in pending list after MarkFullScanned")
		}
	}
}

// --- scans ---

func TestSaveAndLastScan(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()

	const instID = int64(8_000_007)
	t.Cleanup(func() { db.DeleteInstallation(ctx, instID) })
	if err := db.UpsertInstallation(ctx, store.Installation{
		InstallationID: instID, AccountLogin: "u", AccountType: "User", Plan: "free",
	}); err != nil {
		t.Fatalf("UpsertInstallation: %v", err)
	}

	rec := store.ScanRecord{
		InstallationID: instID,
		Owner:          "acme",
		Repo:           "testrepo",
		HeadSHA:        "abc123",
		EvilMerges:     2,
		TotalMerges:    5,
		DurationMs:     300,
	}
	if err := db.SaveScan(ctx, rec); err != nil {
		t.Fatalf("SaveScan: %v", err)
	}

	got, err := db.LastScan(ctx, "acme", "testrepo")
	if err != nil {
		t.Fatalf("LastScan: %v", err)
	}
	if got.EvilMerges != 2 {
		t.Errorf("evil_merges %d, want 2", got.EvilMerges)
	}
	if got.HeadSHA != "abc123" {
		t.Errorf("head_sha %q, want abc123", got.HeadSHA)
	}
}

func TestLastScan_NoRows(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()

	got, err := db.LastScan(ctx, "nonexistent", "repo")
	if err == nil {
		t.Error("expected error for missing scan, got nil")
	}
	if got != nil {
		t.Error("expected nil record")
	}
}

func TestMonthlyScansCount(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()

	const instID = int64(8_000_008)
	t.Cleanup(func() { db.DeleteInstallation(ctx, instID) })
	if err := db.UpsertInstallation(ctx, store.Installation{
		InstallationID: instID, AccountLogin: "u2", AccountType: "User", Plan: "free",
	}); err != nil {
		t.Fatalf("UpsertInstallation: %v", err)
	}

	for range 3 {
		if err := db.SaveScan(ctx, store.ScanRecord{
			InstallationID: instID, Owner: "o", Repo: "r", HeadSHA: "x",
		}); err != nil {
			t.Fatalf("SaveScan: %v", err)
		}
	}

	count, err := db.MonthlyScansCount(ctx, instID)
	if err != nil {
		t.Fatalf("MonthlyScansCount: %v", err)
	}
	if count != 3 {
		t.Errorf("count %d, want 3", count)
	}
}

func TestRecentScans(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()

	const instID = int64(8_000_009)
	t.Cleanup(func() { db.DeleteInstallation(ctx, instID) })
	if err := db.UpsertInstallation(ctx, store.Installation{
		InstallationID: instID, AccountLogin: "u3", AccountType: "User", Plan: "free",
	}); err != nil {
		t.Fatalf("UpsertInstallation: %v", err)
	}

	for i := range 5 {
		if err := db.SaveScan(ctx, store.ScanRecord{
			InstallationID: instID, Owner: "o", Repo: "r",
			HeadSHA: string(rune('a' + i)),
		}); err != nil {
			t.Fatalf("SaveScan %d: %v", i, err)
		}
	}

	scans, err := db.RecentScans(ctx, instID, 3)
	if err != nil {
		t.Fatalf("RecentScans: %v", err)
	}
	if len(scans) != 3 {
		t.Errorf("got %d scans, want 3", len(scans))
	}
}
