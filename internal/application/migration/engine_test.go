package migration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/wuekevin/smartcdn/internal/domain/entity"
	"github.com/wuekevin/smartcdn/internal/domain/port"
)

// --- stubs ---

type stubMigJobRepo struct{ jobs []*entity.MigrationJob }

func (r *stubMigJobRepo) Create(_ context.Context, j *entity.MigrationJob) error {
	j.ID = uint64(len(r.jobs) + 1)
	r.jobs = append(r.jobs, j)
	return nil
}
func (r *stubMigJobRepo) FindByID(_ context.Context, id uint64) (*entity.MigrationJob, error) {
	for _, j := range r.jobs {
		if j.ID == id {
			return j, nil
		}
	}
	return nil, nil
}
func (r *stubMigJobRepo) Update(_ context.Context, _ *entity.MigrationJob) error { return nil }
func (r *stubMigJobRepo) ListByDomain(_ context.Context, _ uint64, _ uint64, _ int) ([]entity.MigrationJob, error) {
	return nil, nil
}

type stubMigDomainRepo struct{}

func (r *stubMigDomainRepo) Create(_ context.Context, _ *entity.Domain) error { return nil }
func (r *stubMigDomainRepo) FindByID(_ context.Context, id uint64) (*entity.Domain, error) {
	return &entity.Domain{ID: id, Name: "migrate.example", AccountID: 1}, nil
}
func (r *stubMigDomainRepo) FindByName(_ context.Context, _ string) (*entity.Domain, error) {
	return nil, nil
}
func (r *stubMigDomainRepo) List(_ context.Context, _ uint64, _ int) ([]entity.Domain, error) {
	return nil, nil
}
func (r *stubMigDomainRepo) Update(_ context.Context, _ *entity.Domain) error { return nil }
func (r *stubMigDomainRepo) Delete(_ context.Context, _ uint64) error         { return nil }

type stubMigRecordRepo struct{}

func (r *stubMigRecordRepo) Create(_ context.Context, _ *entity.DNSRecord) error { return nil }
func (r *stubMigRecordRepo) FindByID(_ context.Context, _ uint64) (*entity.DNSRecord, error) {
	return nil, nil
}
func (r *stubMigRecordRepo) ListByDomain(_ context.Context, _ uint64, _ uint64, _ int) ([]entity.DNSRecord, error) {
	return nil, nil
}
func (r *stubMigRecordRepo) Update(_ context.Context, _ *entity.DNSRecord) error  { return nil }
func (r *stubMigRecordRepo) Delete(_ context.Context, _ uint64) error             { return nil }
func (r *stubMigRecordRepo) DeleteByDomain(_ context.Context, _ uint64) error     { return nil }
func (r *stubMigRecordRepo) FindByProviderID(_ context.Context, _ uint64, _ string) (*entity.DNSRecord, error) {
	return nil, nil
}

type stubMigSnapshotRepo struct{}

func (r *stubMigSnapshotRepo) Create(_ context.Context, s *entity.DNSSnapshot) error {
	s.ID = 1; s.CreatedAt = time.Now(); return nil
}
func (r *stubMigSnapshotRepo) FindByID(_ context.Context, _ uint64) (*entity.DNSSnapshot, error) {
	return nil, nil
}
func (r *stubMigSnapshotRepo) ListByDomain(_ context.Context, _ uint64, _ uint64, _ int) ([]entity.DNSSnapshot, error) {
	return nil, nil
}

type stubMigAccountRepo struct{}

func (r *stubMigAccountRepo) Create(_ context.Context, _ *entity.DNSAccount) error { return nil }
func (r *stubMigAccountRepo) FindByID(_ context.Context, id uint64) (*entity.DNSAccount, error) {
	return nil, fmt.Errorf("account %d not found", id)
}
func (r *stubMigAccountRepo) List(_ context.Context, _ uint64, _ int) ([]entity.DNSAccount, error) {
	return nil, nil
}
func (r *stubMigAccountRepo) UpdateCredentials(_ context.Context, _ uint64, _ string) error {
	return nil
}
func (r *stubMigAccountRepo) UpdateCapabilities(_ context.Context, _ uint64, _ string) error {
	return nil
}
func (r *stubMigAccountRepo) Delete(_ context.Context, _ uint64) error { return nil }

// TestStartMigration_FailsWhenAccountNotFound verifies the engine fails gracefully
// when the source/target account doesn't exist (no credentials → no provider).
func TestStartMigration_FailsWhenAccountNotFound(t *testing.T) {
	engine := &Engine{
		jobs:      &stubMigJobRepo{},
		domains:   &stubMigDomainRepo{},
		records:   &stubMigRecordRepo{},
		snapshots: &stubMigSnapshotRepo{},
		accounts:  &stubMigAccountRepo{},
		registry:  nil,
	}

	result, err := engine.Start(context.Background(), StartRequest{
		DomainID:        1,
		TargetAccountID: 2,
	})
	// Should fail at export_source (can't load account credentials), not panic
	if err != nil && result == nil {
		t.Fatal("expected non-nil result even on failure")
	}
	if result != nil {
		if len(result.StepLogs) == 0 {
			t.Error("expected at least one step log")
		}
		hasSnapshotOK := false
		for _, sl := range result.StepLogs {
			if sl.Step == "snapshot_source" && sl.Status == "ok" {
				hasSnapshotOK = true
			}
		}
		if !hasSnapshotOK {
			t.Error("expected snapshot_source step to succeed before export fails")
		}
	}
}

// TestRollback_CannotRollbackDoneJob verifies rollback is rejected for completed jobs.
func TestRollback_CannotRollbackDoneJob(t *testing.T) {
	repo := &stubMigJobRepo{}
	doneJob := &entity.MigrationJob{Status: entity.MigrationDone}
	_ = repo.Create(context.Background(), doneJob)

	engine := &Engine{
		jobs:    repo,
		domains: &stubMigDomainRepo{},
		records: &stubMigRecordRepo{},
	}

	_, err := engine.Rollback(context.Background(), doneJob.ID)
	if err == nil {
		t.Error("expected error when rolling back a done job")
	}

	_ = port.ErrProviderNotFound // ensure port package linkage
}
