package port

import (
	"context"
	"time"

	"github.com/wuekevin/smartcdn/internal/domain/entity"
)

// MetricRepository persists time-series metric data points.
type MetricRepository interface {
	BatchInsert(ctx context.Context, metrics []entity.Metric) error
	// QueryWindow returns aggregated (avg) metric values in [from, to] grouped by sourceID.
	QueryWindow(ctx context.Context, sourceType string, metricType entity.MetricType, from, to time.Time) ([]MetricAgg, error)
	// QueryBySource returns raw data points for one source, most-recent first.
	QueryBySource(ctx context.Context, sourceType string, sourceID uint64, metricType entity.MetricType, limit int) ([]entity.Metric, error)
}

// MetricAgg is a single aggregated result row from QueryWindow.
type MetricAgg struct {
	SourceID uint64
	AvgValue float64
	MinValue float64
	MaxValue float64
	Count    int64
}

// AlertRuleRepository persists alert rule definitions.
type AlertRuleRepository interface {
	Create(ctx context.Context, r *entity.AlertRule) error
	FindByID(ctx context.Context, id uint64) (*entity.AlertRule, error)
	ListEnabled(ctx context.Context) ([]entity.AlertRule, error)
	List(ctx context.Context, afterID uint64, limit int) ([]entity.AlertRule, error)
	Update(ctx context.Context, r *entity.AlertRule) error
	Delete(ctx context.Context, id uint64) error
}

// AlertRepository persists fired alert instances.
type AlertRepository interface {
	Create(ctx context.Context, a *entity.Alert) error
	FindByID(ctx context.Context, id uint64) (*entity.Alert, error)
	FindFiringByRule(ctx context.Context, ruleID, sourceID uint64) (*entity.Alert, error)
	List(ctx context.Context, status entity.AlertStatus, afterID uint64, limit int) ([]entity.Alert, error)
	Update(ctx context.Context, a *entity.Alert) error
}

// AlertChannelRepository persists notification channel configs.
type AlertChannelRepository interface {
	Create(ctx context.Context, ch *entity.AlertChannel) error
	FindByID(ctx context.Context, id uint64) (*entity.AlertChannel, error)
	List(ctx context.Context, afterID uint64, limit int) ([]entity.AlertChannel, error)
	Update(ctx context.Context, ch *entity.AlertChannel) error
	Delete(ctx context.Context, id uint64) error
}
