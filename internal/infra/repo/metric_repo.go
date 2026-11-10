package repo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/wuekevin/smartcdn/internal/domain/entity"
	"github.com/wuekevin/smartcdn/internal/domain/port"
)

// MetricRepo ──────────────────────────────────────────────────────────────────

type MetricRepo struct{ db *gorm.DB }

func NewMetricRepo(db *gorm.DB) *MetricRepo { return &MetricRepo{db: db} }

func (r *MetricRepo) BatchInsert(ctx context.Context, metrics []entity.Metric) error {
	if len(metrics) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).CreateInBatches(metrics, 100).Error
}

func (r *MetricRepo) QueryWindow(ctx context.Context, sourceType string, metricType entity.MetricType, from, to time.Time) ([]port.MetricAgg, error) {
	type row struct {
		SourceID uint64
		AvgValue float64
		MinValue float64
		MaxValue float64
		Count    int64
	}
	var rows []row
	err := r.db.WithContext(ctx).Model(&entity.Metric{}).
		Select("source_id, AVG(value) as avg_value, MIN(value) as min_value, MAX(value) as max_value, COUNT(*) as count").
		Where("source_type = ? AND type = ? AND collected_at BETWEEN ? AND ?", sourceType, metricType, from, to).
		Group("source_id").
		Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("MetricRepo.QueryWindow: %w", err)
	}
	out := make([]port.MetricAgg, len(rows))
	for i, row := range rows {
		out[i] = port.MetricAgg{
			SourceID: row.SourceID,
			AvgValue: row.AvgValue,
			MinValue: row.MinValue,
			MaxValue: row.MaxValue,
			Count:    row.Count,
		}
	}
	return out, nil
}

func (r *MetricRepo) QueryBySource(ctx context.Context, sourceType string, sourceID uint64, metricType entity.MetricType, limit int) ([]entity.Metric, error) {
	if limit <= 0 {
		limit = 100
	}
	var ms []entity.Metric
	err := r.db.WithContext(ctx).
		Where("source_type = ? AND source_id = ? AND type = ?", sourceType, sourceID, metricType).
		Order("collected_at DESC").Limit(limit).Find(&ms).Error
	return ms, err
}

// AlertRuleRepo ───────────────────────────────────────────────────────────────

type AlertRuleRepo struct{ db *gorm.DB }

func NewAlertRuleRepo(db *gorm.DB) *AlertRuleRepo { return &AlertRuleRepo{db: db} }

func (r *AlertRuleRepo) Create(ctx context.Context, rule *entity.AlertRule) error {
	return r.db.WithContext(ctx).Create(rule).Error
}

func (r *AlertRuleRepo) FindByID(ctx context.Context, id uint64) (*entity.AlertRule, error) {
	var rule entity.AlertRule
	if err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&rule).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("AlertRuleRepo.FindByID: %w", err)
	}
	return &rule, nil
}

func (r *AlertRuleRepo) ListEnabled(ctx context.Context) ([]entity.AlertRule, error) {
	var rules []entity.AlertRule
	return rules, r.db.WithContext(ctx).Where("enabled = 1 AND deleted_at IS NULL").Find(&rules).Error
}

func (r *AlertRuleRepo) List(ctx context.Context, afterID uint64, limit int) ([]entity.AlertRule, error) {
	if limit <= 0 {
		limit = 20
	}
	var rules []entity.AlertRule
	q := r.db.WithContext(ctx).Where("deleted_at IS NULL")
	if afterID > 0 {
		q = q.Where("id > ?", afterID)
	}
	return rules, q.Order("id ASC").Limit(limit).Find(&rules).Error
}

func (r *AlertRuleRepo) Update(ctx context.Context, rule *entity.AlertRule) error {
	return r.db.WithContext(ctx).Save(rule).Error
}

func (r *AlertRuleRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&entity.AlertRule{}).Error
}

// AlertRepo ───────────────────────────────────────────────────────────────────

type AlertRepo struct{ db *gorm.DB }

func NewAlertRepo(db *gorm.DB) *AlertRepo { return &AlertRepo{db: db} }

func (r *AlertRepo) Create(ctx context.Context, a *entity.Alert) error {
	return r.db.WithContext(ctx).Create(a).Error
}

func (r *AlertRepo) FindByID(ctx context.Context, id uint64) (*entity.Alert, error) {
	var a entity.Alert
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&a).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("AlertRepo.FindByID: %w", err)
	}
	return &a, nil
}

func (r *AlertRepo) FindFiringByRule(ctx context.Context, ruleID, sourceID uint64) (*entity.Alert, error) {
	var a entity.Alert
	if err := r.db.WithContext(ctx).
		Where("rule_id = ? AND source_id = ? AND status = 'firing'", ruleID, sourceID).
		First(&a).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("AlertRepo.FindFiringByRule: %w", err)
	}
	return &a, nil
}

func (r *AlertRepo) List(ctx context.Context, status entity.AlertStatus, afterID uint64, limit int) ([]entity.Alert, error) {
	if limit <= 0 {
		limit = 20
	}
	var alerts []entity.Alert
	q := r.db.WithContext(ctx)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if afterID > 0 {
		q = q.Where("id > ?", afterID)
	}
	return alerts, q.Order("id DESC").Limit(limit).Find(&alerts).Error
}

func (r *AlertRepo) Update(ctx context.Context, a *entity.Alert) error {
	return r.db.WithContext(ctx).Save(a).Error
}

// AlertChannelRepo ─────────────────────────────────────────────────────────────

type AlertChannelRepo struct{ db *gorm.DB }

func NewAlertChannelRepo(db *gorm.DB) *AlertChannelRepo { return &AlertChannelRepo{db: db} }

func (r *AlertChannelRepo) Create(ctx context.Context, ch *entity.AlertChannel) error {
	return r.db.WithContext(ctx).Create(ch).Error
}

func (r *AlertChannelRepo) FindByID(ctx context.Context, id uint64) (*entity.AlertChannel, error) {
	var ch entity.AlertChannel
	if err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&ch).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("AlertChannelRepo.FindByID: %w", err)
	}
	return &ch, nil
}

func (r *AlertChannelRepo) List(ctx context.Context, afterID uint64, limit int) ([]entity.AlertChannel, error) {
	if limit <= 0 {
		limit = 20
	}
	var chs []entity.AlertChannel
	q := r.db.WithContext(ctx).Where("deleted_at IS NULL")
	if afterID > 0 {
		q = q.Where("id > ?", afterID)
	}
	return chs, q.Order("id ASC").Limit(limit).Find(&chs).Error
}

func (r *AlertChannelRepo) Update(ctx context.Context, ch *entity.AlertChannel) error {
	return r.db.WithContext(ctx).Save(ch).Error
}

func (r *AlertChannelRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&entity.AlertChannel{}).Error
}
