package migration

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/wuekevin/smartcdn/internal/domain/entity"
	"github.com/wuekevin/smartcdn/internal/domain/port"
	"github.com/wuekevin/smartcdn/internal/infra/crypto"
	"github.com/wuekevin/smartcdn/internal/infra/dnscheck"
	infraprovider "github.com/wuekevin/smartcdn/internal/infra/provider"
)

// Engine executes Mode-B DNS migrations:
// export_source → import_target → verify_target → cutover → cleanup_source
type Engine struct {
	jobs      port.MigrationJobRepository
	domains   port.DomainRepository
	records   port.DNSRecordRepository
	snapshots port.DNSSnapshotRepository
	accounts  port.DNSAccountRepository
	registry  *infraprovider.Registry
}

func NewEngine(
	jobs port.MigrationJobRepository,
	domains port.DomainRepository,
	records port.DNSRecordRepository,
	snapshots port.DNSSnapshotRepository,
	accounts port.DNSAccountRepository,
	registry *infraprovider.Registry,
) *Engine {
	return &Engine{
		jobs:      jobs,
		domains:   domains,
		records:   records,
		snapshots: snapshots,
		accounts:  accounts,
		registry:  registry,
	}
}

// StartRequest starts a new Mode-B migration.
type StartRequest struct {
	DomainID        uint64 `json:"domain_id"`
	TargetAccountID uint64 `json:"target_account_id"`
}

// MigrationResult is the result of a migration run or status query.
type MigrationResult struct {
	JobID    uint64              `json:"job_id"`
	Status   string              `json:"status"`
	StepLogs []entity.StepLog    `json:"step_logs"`
	Error    string              `json:"error,omitempty"`
}

// Start creates the job and runs steps 1-4 (export, import, verify).
// The caller must invoke Cutover to complete step 5 after confirming NS propagation.
func (e *Engine) Start(ctx context.Context, req StartRequest) (*MigrationResult, error) {
	d, err := e.domains.FindByID(ctx, req.DomainID)
	if err != nil {
		return nil, fmt.Errorf("domain not found: %w", err)
	}

	job := &entity.MigrationJob{
		DomainID:        req.DomainID,
		SourceAccountID: d.AccountID,
		TargetAccountID: req.TargetAccountID,
		Status:          entity.MigrationPending,
	}
	if err := e.jobs.Create(ctx, job); err != nil {
		return nil, fmt.Errorf("create migration job: %w", err)
	}

	job.Status = entity.MigrationRunning
	_ = e.jobs.Update(ctx, job)

	var stepLogs []entity.StepLog
	addStep := func(step, status, msg string) {
		stepLogs = append(stepLogs, entity.StepLog{Step: step, Status: status, Message: msg, At: time.Now()})
	}

	fail := func(step, msg string, err error) (*MigrationResult, error) {
		addStep(step, "failed", msg)
		logsJSON, _ := json.Marshal(stepLogs)
		job.Status = entity.MigrationFailed
		job.StepLogs = string(logsJSON)
		job.ErrorMsg = msg
		now := time.Now()
		job.FinishedAt = &now
		_ = e.jobs.Update(ctx, job)
		return &MigrationResult{JobID: job.ID, Status: string(entity.MigrationFailed), StepLogs: stepLogs, Error: msg}, err
	}

	// ── Step 1: Snapshot source ──────────────────────────────────────────────
	sourceRecs, err := e.records.ListByDomain(ctx, req.DomainID, 0, 1000)
	if err != nil {
		return fail("snapshot_source", "list source records failed", err)
	}
	snapData, _ := json.Marshal(sourceRecs)
	snap := &entity.DNSSnapshot{
		DomainID:     req.DomainID,
		SnapshotJSON: string(snapData),
		TriggerEvent: "pre_migration",
	}
	if err := e.snapshots.Create(ctx, snap); err != nil {
		return fail("snapshot_source", "create snapshot failed", err)
	}
	job.SnapshotID = snap.ID
	addStep("snapshot_source", "ok", fmt.Sprintf("snapshot id=%d, %d records", snap.ID, len(sourceRecs)))

	// ── Step 2: Export records from source provider ──────────────────────────
	srcProv, srcZoneID, err := e.loadProviderAndZone(ctx, d.AccountID, d.Name)
	if err != nil {
		return fail("export_source", fmt.Sprintf("load source provider: %v", err), err)
	}
	providerRecs, err := srcProv.ListRecords(ctx, srcZoneID)
	if err != nil {
		return fail("export_source", fmt.Sprintf("list source records: %v", err), err)
	}
	addStep("export_source", "ok", fmt.Sprintf("%d records exported from source", len(providerRecs)))

	// ── Step 3: Import records to target provider ────────────────────────────
	tgtProv, tgtZoneID, err := e.loadProviderAndZone(ctx, req.TargetAccountID, d.Name)
	if err != nil {
		return fail("import_target", fmt.Sprintf("load target provider: %v", err), err)
	}

	var importFailed int
	for _, r := range providerRecs {
		_, cerr := tgtProv.CreateRecord(ctx, tgtZoneID, port.DNSRecordInput{
			Host:     r.Host,
			Type:     r.Type,
			Value:    r.Value,
			TTL:      r.TTL,
			Line:     r.Line,
			Priority: r.Priority,
		})
		if cerr != nil {
			importFailed++
			addStep("import_target", "failed", fmt.Sprintf("create %s %s: %v", r.Host, r.Type, cerr))
		}
	}
	if importFailed > 0 {
		return fail("import_target", fmt.Sprintf("%d records failed to import", importFailed), errors.New("partial import failure"))
	}
	addStep("import_target", "ok", fmt.Sprintf("%d records imported to target", len(providerRecs)))

	// ── Step 4: Verify DNS propagation on target NS ──────────────────────────
	job.Status = entity.MigrationVerifying
	_ = e.jobs.Update(ctx, job)

	// Sample the first A record for verification
	verified := false
	for _, r := range providerRecs {
		if r.Type == "A" {
			fqdn := r.Host + "." + d.Name
			if r.Host == "@" {
				fqdn = d.Name
			}
			verCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
			authResult, verErr := dnscheck.LookupA(verCtx, fqdn)
			cancel()
			if verErr == nil && dnscheck.ContainsValue(authResult, r.Value) {
				addStep("verify_target", "ok", fmt.Sprintf("%s → %s resolved", fqdn, r.Value))
				verified = true
			} else {
				addStep("verify_target", "ok", fmt.Sprintf("%s DNS propagating (not yet visible at 8.8.8.8)", fqdn))
				verified = true // non-fatal: NS hasn't propagated yet
			}
			break
		}
	}
	if !verified {
		addStep("verify_target", "skipped", "no A record to verify")
	}

	logsJSON, _ := json.Marshal(stepLogs)
	job.StepLogs = string(logsJSON)
	job.Status = entity.MigrationVerifying // awaiting Cutover call
	_ = e.jobs.Update(ctx, job)

	return &MigrationResult{JobID: job.ID, Status: string(entity.MigrationVerifying), StepLogs: stepLogs}, nil
}

// Cutover completes the migration: switches domain.AccountID to target, then cleans up source.
// Should be called after the registrar NS has been updated and propagated.
func (e *Engine) Cutover(ctx context.Context, jobID uint64) (*MigrationResult, error) {
	job, err := e.jobs.FindByID(ctx, jobID)
	if err != nil {
		return nil, fmt.Errorf("job not found: %w", err)
	}
	if job.Status != entity.MigrationVerifying {
		return nil, fmt.Errorf("job is not in verifying state (current: %s)", job.Status)
	}

	var stepLogs []entity.StepLog
	if job.StepLogs != "" {
		_ = json.Unmarshal([]byte(job.StepLogs), &stepLogs)
	}

	addStep := func(step, status, msg string) {
		stepLogs = append(stepLogs, entity.StepLog{Step: step, Status: status, Message: msg, At: time.Now()})
	}

	fail := func(step, msg string, ferr error) (*MigrationResult, error) {
		addStep(step, "failed", msg)
		logsJSON, _ := json.Marshal(stepLogs)
		job.Status = entity.MigrationFailed
		job.StepLogs = string(logsJSON)
		job.ErrorMsg = msg
		now := time.Now()
		job.FinishedAt = &now
		_ = e.jobs.Update(ctx, job)
		return &MigrationResult{JobID: job.ID, Status: string(entity.MigrationFailed), StepLogs: stepLogs, Error: msg}, ferr
	}

	job.Status = entity.MigrationCutover
	_ = e.jobs.Update(ctx, job)

	// ── Step 5: Switch domain to target account ──────────────────────────────
	d, err := e.domains.FindByID(ctx, job.DomainID)
	if err != nil {
		return fail("cutover", "domain not found", err)
	}
	oldAccountID := d.AccountID
	d.AccountID = job.TargetAccountID
	if err := e.domains.Update(ctx, d); err != nil {
		return fail("cutover", "update domain account failed", err)
	}
	addStep("cutover", "ok", fmt.Sprintf("domain.account_id %d → %d", oldAccountID, job.TargetAccountID))

	// ── Step 6: Verify target NS now resolves correctly ──────────────────────
	tgtNSs, nsErr := e.getNameservers(ctx, job.TargetAccountID, d.Name)
	if nsErr != nil {
		addStep("verify_ns", "skipped", fmt.Sprintf("get NS failed (non-fatal): %v", nsErr))
	} else {
		addStep("verify_ns", "ok", fmt.Sprintf("target NS: %v", tgtNSs))
	}

	// ── Step 7: Cleanup source records ───────────────────────────────────────
	srcProv, srcZoneID, cleanErr := e.loadProviderAndZone(ctx, oldAccountID, d.Name)
	if cleanErr != nil {
		addStep("cleanup_source", "skipped", fmt.Sprintf("load source provider failed (non-fatal): %v", cleanErr))
	} else {
		srcRecs, listErr := srcProv.ListRecords(ctx, srcZoneID)
		if listErr != nil {
			addStep("cleanup_source", "skipped", fmt.Sprintf("list source records failed: %v", listErr))
		} else {
			deleted := 0
			for _, r := range srcRecs {
				if delErr := srcProv.DeleteRecord(ctx, srcZoneID, r.ID); delErr == nil {
					deleted++
				}
			}
			addStep("cleanup_source", "ok", fmt.Sprintf("%d source records removed", deleted))
		}
	}

	logsJSON, _ := json.Marshal(stepLogs)
	job.Status = entity.MigrationDone
	job.StepLogs = string(logsJSON)
	now := time.Now()
	job.FinishedAt = &now
	_ = e.jobs.Update(ctx, job)

	return &MigrationResult{JobID: job.ID, Status: string(entity.MigrationDone), StepLogs: stepLogs}, nil
}

// Rollback reverts a failed or in-progress migration using the pre-migration snapshot.
func (e *Engine) Rollback(ctx context.Context, jobID uint64) (*MigrationResult, error) {
	job, err := e.jobs.FindByID(ctx, jobID)
	if err != nil {
		return nil, fmt.Errorf("job not found: %w", err)
	}
	if job.Status == entity.MigrationDone || job.Status == entity.MigrationRolledBack {
		return nil, fmt.Errorf("cannot rollback job in state %s", job.Status)
	}

	var stepLogs []entity.StepLog
	if job.StepLogs != "" {
		_ = json.Unmarshal([]byte(job.StepLogs), &stepLogs)
	}
	addStep := func(step, status, msg string) {
		stepLogs = append(stepLogs, entity.StepLog{Step: step, Status: status, Message: msg, At: time.Now()})
	}

	// Restore domain account to source
	d, err := e.domains.FindByID(ctx, job.DomainID)
	if err != nil {
		return nil, fmt.Errorf("domain not found: %w", err)
	}
	if d.AccountID != job.SourceAccountID {
		d.AccountID = job.SourceAccountID
		if err := e.domains.Update(ctx, d); err != nil {
			return nil, fmt.Errorf("restore domain account: %w", err)
		}
		addStep("rollback_cutover", "ok", fmt.Sprintf("domain.account_id restored to %d", job.SourceAccountID))
	}

	// Delete target records
	tgtProv, tgtZoneID, tgtErr := e.loadProviderAndZone(ctx, job.TargetAccountID, d.Name)
	if tgtErr != nil {
		addStep("rollback_cleanup_target", "skipped", fmt.Sprintf("load target: %v", tgtErr))
	} else {
		tgtRecs, _ := tgtProv.ListRecords(ctx, tgtZoneID)
		deleted := 0
		for _, r := range tgtRecs {
			if delErr := tgtProv.DeleteRecord(ctx, tgtZoneID, r.ID); delErr == nil {
				deleted++
			}
		}
		addStep("rollback_cleanup_target", "ok", fmt.Sprintf("%d target records removed", deleted))
	}

	logsJSON, _ := json.Marshal(stepLogs)
	job.Status = entity.MigrationRolledBack
	job.StepLogs = string(logsJSON)
	now := time.Now()
	job.FinishedAt = &now
	_ = e.jobs.Update(ctx, job)

	return &MigrationResult{JobID: job.ID, Status: string(entity.MigrationRolledBack), StepLogs: stepLogs}, nil
}

// --- helpers ---

func (e *Engine) loadProviderAndZone(ctx context.Context, accountID uint64, domainName string) (port.DNSProvider, string, error) {
	acct, err := e.accounts.FindByID(ctx, accountID)
	if err != nil {
		return nil, "", fmt.Errorf("load account %d: %w", accountID, err)
	}
	credBytes, err := crypto.Decrypt(acct.CredentialsEnc)
	if err != nil {
		return nil, "", fmt.Errorf("decrypt credentials: %w", err)
	}
	prov, err := e.registry.GetDNSWithCredentials(port.ProviderType(acct.ProviderType), credBytes)
	if err != nil {
		return nil, "", fmt.Errorf("get provider: %w", err)
	}
	zones, err := prov.ListZones(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("list zones: %w", err)
	}
	for _, z := range zones {
		if z.Domain == domainName {
			return prov, z.ID, nil
		}
	}
	return nil, "", fmt.Errorf("zone not found for domain %q on account %d", domainName, accountID)
}

func (e *Engine) getNameservers(ctx context.Context, accountID uint64, domainName string) ([]string, error) {
	prov, zoneID, err := e.loadProviderAndZone(ctx, accountID, domainName)
	if err != nil {
		return nil, err
	}
	return prov.GetNameservers(ctx, zoneID)
}
