package monitor

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/wuekevin/smartcdn/internal/domain/entity"
	"github.com/wuekevin/smartcdn/internal/domain/port"
	"github.com/wuekevin/smartcdn/internal/infra/repo"
)

var (
	ErrRuleNotFound    = errors.New("alert rule not found")
	ErrChannelNotFound = errors.New("alert channel not found")
	ErrAlertNotFound   = errors.New("alert not found")
)

// Service handles metric ingestion, alert evaluation, and notification dispatch.
type Service struct {
	metrics  port.MetricRepository
	rules    port.AlertRuleRepository
	alerts   port.AlertRepository
	channels port.AlertChannelRepository
	http     *http.Client
}

func NewService(
	metrics port.MetricRepository,
	rules port.AlertRuleRepository,
	alerts port.AlertRepository,
	channels port.AlertChannelRepository,
) *Service {
	return &Service{
		metrics:  metrics,
		rules:    rules,
		alerts:   alerts,
		channels: channels,
		http:     &http.Client{Timeout: 10 * time.Second},
	}
}

// --- Metric ingestion ---

type IngestInput struct {
	SourceType  string
	SourceID    uint64
	MetricType  entity.MetricType
	Value       float64
	Tags        map[string]string
	CollectedAt time.Time
}

func (s *Service) IngestMetrics(ctx context.Context, inputs []IngestInput) error {
	if len(inputs) == 0 {
		return nil
	}
	ms := make([]entity.Metric, len(inputs))
	for i, in := range inputs {
		tagJ, _ := json.Marshal(in.Tags)
		at := in.CollectedAt
		if at.IsZero() {
			at = time.Now()
		}
		ms[i] = entity.Metric{
			SourceType:  in.SourceType,
			SourceID:    in.SourceID,
			Type:        in.MetricType,
			Value:       in.Value,
			Tags:        string(tagJ),
			CollectedAt: at,
		}
	}
	if err := s.metrics.BatchInsert(ctx, ms); err != nil {
		return fmt.Errorf("monitor: ingest metrics: %w", err)
	}
	// Evaluate rules asynchronously after ingestion.
	go s.EvaluateAll(context.Background())
	return nil
}

func (s *Service) QueryMetrics(ctx context.Context, sourceType string, sourceID uint64, metricType entity.MetricType, limit int) ([]entity.Metric, error) {
	return s.metrics.QueryBySource(ctx, sourceType, sourceID, metricType, limit)
}

// --- Alert rules ---

type CreateRuleInput struct {
	Name          string
	Description   string
	SourceType    string
	SourceID      *uint64
	MetricType    entity.MetricType
	Condition     entity.AlertCondition
	Threshold     float64
	WindowMinutes int
	ForMinutes    int
	Severity      entity.AlertSeverity
	ChannelIDs    []uint64
}

func (s *Service) CreateRule(ctx context.Context, in CreateRuleInput) (*entity.AlertRule, error) {
	chIDs, _ := json.Marshal(in.ChannelIDs)
	rule := &entity.AlertRule{
		Name:          in.Name,
		Description:   in.Description,
		SourceType:    in.SourceType,
		SourceID:      in.SourceID,
		MetricType:    in.MetricType,
		Condition:     in.Condition,
		Threshold:     in.Threshold,
		WindowMinutes: windowDefault(in.WindowMinutes, 5),
		ForMinutes:    windowDefault(in.ForMinutes, 1),
		Severity:      severityDefault(in.Severity),
		Enabled:       true,
		ChannelIDs:    string(chIDs),
	}
	if err := s.rules.Create(ctx, rule); err != nil {
		return nil, fmt.Errorf("monitor: create rule: %w", err)
	}
	return rule, nil
}

func (s *Service) GetRule(ctx context.Context, id uint64) (*entity.AlertRule, error) {
	r, err := s.rules.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrRuleNotFound
		}
		return nil, err
	}
	return r, nil
}

func (s *Service) ListRules(ctx context.Context, afterID uint64, limit int) ([]entity.AlertRule, error) {
	return s.rules.List(ctx, afterID, limit)
}

func (s *Service) UpdateRule(ctx context.Context, id uint64, in CreateRuleInput) (*entity.AlertRule, error) {
	rule, err := s.rules.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrRuleNotFound
		}
		return nil, err
	}
	chIDs, _ := json.Marshal(in.ChannelIDs)
	rule.Name = in.Name
	rule.Description = in.Description
	rule.SourceType = in.SourceType
	rule.SourceID = in.SourceID
	rule.MetricType = in.MetricType
	rule.Condition = in.Condition
	rule.Threshold = in.Threshold
	rule.WindowMinutes = windowDefault(in.WindowMinutes, 5)
	rule.ForMinutes = windowDefault(in.ForMinutes, 1)
	rule.Severity = severityDefault(in.Severity)
	rule.ChannelIDs = string(chIDs)
	if err := s.rules.Update(ctx, rule); err != nil {
		return nil, err
	}
	return rule, nil
}

func (s *Service) DeleteRule(ctx context.Context, id uint64) error {
	if _, err := s.rules.FindByID(ctx, id); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return ErrRuleNotFound
		}
		return err
	}
	return s.rules.Delete(ctx, id)
}

// --- Alert channel ---

type CreateChannelInput struct {
	Name   string
	Type   entity.AlertChannelType
	Config map[string]interface{}
}

func (s *Service) CreateChannel(ctx context.Context, in CreateChannelInput) (*entity.AlertChannel, error) {
	cfgJ, _ := json.Marshal(in.Config)
	ch := &entity.AlertChannel{
		Name:       in.Name,
		Type:       in.Type,
		ConfigJSON: string(cfgJ),
		Enabled:    true,
	}
	if err := s.channels.Create(ctx, ch); err != nil {
		return nil, fmt.Errorf("monitor: create channel: %w", err)
	}
	return ch, nil
}

func (s *Service) GetChannel(ctx context.Context, id uint64) (*entity.AlertChannel, error) {
	ch, err := s.channels.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrChannelNotFound
		}
		return nil, err
	}
	return ch, nil
}

func (s *Service) ListChannels(ctx context.Context, afterID uint64, limit int) ([]entity.AlertChannel, error) {
	return s.channels.List(ctx, afterID, limit)
}

func (s *Service) DeleteChannel(ctx context.Context, id uint64) error {
	if _, err := s.channels.FindByID(ctx, id); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return ErrChannelNotFound
		}
		return err
	}
	return s.channels.Delete(ctx, id)
}

// --- Alert queries ---

func (s *Service) ListAlerts(ctx context.Context, status entity.AlertStatus, afterID uint64, limit int) ([]entity.Alert, error) {
	return s.alerts.List(ctx, status, afterID, limit)
}

func (s *Service) ResolveAlert(ctx context.Context, id uint64) error {
	a, err := s.alerts.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return ErrAlertNotFound
		}
		return err
	}
	now := time.Now()
	a.Status = entity.AlertStatusResolved
	a.ResolvedAt = &now
	return s.alerts.Update(ctx, a)
}

// --- Evaluation engine ---

// EvaluateAll evaluates every enabled rule and fires/resolves alerts.
// Called periodically by a background goroutine or after metric ingestion.
func (s *Service) EvaluateAll(ctx context.Context) {
	rules, err := s.rules.ListEnabled(ctx)
	if err != nil {
		return
	}
	now := time.Now()
	for _, rule := range rules {
		from := now.Add(-time.Duration(rule.WindowMinutes) * time.Minute)
		aggs, err := s.metrics.QueryWindow(ctx, rule.SourceType, rule.MetricType, from, now)
		if err != nil {
			continue
		}
		for _, agg := range aggs {
			if rule.SourceID != nil && agg.SourceID != *rule.SourceID {
				continue
			}
			if s.conditionMet(agg.AvgValue, rule.Condition, rule.Threshold) {
				s.fireAlert(ctx, &rule, agg.SourceID, agg.AvgValue)
			} else {
				s.autoResolve(ctx, rule.ID, agg.SourceID)
			}
		}
	}
}

func (s *Service) conditionMet(value float64, cond entity.AlertCondition, threshold float64) bool {
	switch cond {
	case entity.AlertCondGT:
		return value > threshold
	case entity.AlertCondGTE:
		return value >= threshold
	case entity.AlertCondLT:
		return value < threshold
	case entity.AlertCondLTE:
		return value <= threshold
	case entity.AlertCondEQ:
		return value == threshold
	}
	return false
}

func (s *Service) fireAlert(ctx context.Context, rule *entity.AlertRule, sourceID uint64, value float64) {
	// Idempotent: skip if already firing.
	existing, err := s.alerts.FindFiringByRule(ctx, rule.ID, sourceID)
	if err == nil && existing != nil {
		return
	}
	msg := fmt.Sprintf("%s: %s %.2f %s %.2f",
		rule.Name, rule.MetricType, value, rule.Condition, rule.Threshold)
	a := &entity.Alert{
		RuleID:       rule.ID,
		RuleName:     rule.Name,
		SourceType:   rule.SourceType,
		SourceID:     sourceID,
		MetricType:   rule.MetricType,
		TriggerValue: value,
		Threshold:    rule.Threshold,
		Severity:     rule.Severity,
		Status:       entity.AlertStatusFiring,
		FiredAt:      time.Now(),
		Message:      msg,
	}
	if err := s.alerts.Create(ctx, a); err != nil {
		return
	}
	s.notifyChannels(ctx, rule, a)
}

func (s *Service) autoResolve(ctx context.Context, ruleID, sourceID uint64) {
	a, err := s.alerts.FindFiringByRule(ctx, ruleID, sourceID)
	if err != nil {
		return
	}
	now := time.Now()
	a.Status = entity.AlertStatusResolved
	a.ResolvedAt = &now
	_ = s.alerts.Update(ctx, a)
}

// --- Notification dispatch ---

func (s *Service) notifyChannels(ctx context.Context, rule *entity.AlertRule, a *entity.Alert) {
	var channelIDs []uint64
	_ = json.Unmarshal([]byte(rule.ChannelIDs), &channelIDs)
	for _, cid := range channelIDs {
		ch, err := s.channels.FindByID(ctx, cid)
		if err != nil || !ch.Enabled {
			continue
		}
		go s.sendNotification(ch, a)
	}
}

func (s *Service) sendNotification(ch *entity.AlertChannel, a *entity.Alert) {
	var cfg map[string]interface{}
	_ = json.Unmarshal([]byte(ch.ConfigJSON), &cfg)

	switch ch.Type {
	case entity.AlertChannelWebhook, entity.AlertChannelDingTalk, entity.AlertChannelSlack:
		s.sendWebhook(cfg, ch.Type, a)
	case entity.AlertChannelEmail:
		// Email requires SMTP integration; log for now.
	}
}

func (s *Service) sendWebhook(cfg map[string]interface{}, chType entity.AlertChannelType, a *entity.Alert) {
	url, _ := cfg["url"].(string)
	if url == "" {
		return
	}

	var payload map[string]interface{}
	switch chType {
	case entity.AlertChannelDingTalk:
		payload = map[string]interface{}{
			"msgtype": "text",
			"text": map[string]string{
				"content": fmt.Sprintf("[%s] %s", strings.ToUpper(string(a.Severity)), a.Message),
			},
		}
	case entity.AlertChannelSlack:
		payload = map[string]interface{}{
			"text": fmt.Sprintf(":warning: *[%s]* %s", strings.ToUpper(string(a.Severity)), a.Message),
		}
	default:
		payload = map[string]interface{}{
			"alert_id":   a.ID,
			"rule_name":  a.RuleName,
			"severity":   a.Severity,
			"status":     a.Status,
			"message":    a.Message,
			"fired_at":   a.FiredAt,
		}
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.http.Do(req)
	if err != nil {
		return
	}
	resp.Body.Close()
}

// --- helpers ---

func windowDefault(v, def int) int {
	if v <= 0 {
		return def
	}
	return v
}

func severityDefault(s entity.AlertSeverity) entity.AlertSeverity {
	if s == "" {
		return entity.AlertSeverityWarning
	}
	return s
}
