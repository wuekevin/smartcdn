package aliyun

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	alidns "github.com/alibabacloud-go/alidns-20150109/v4/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	"github.com/alibabacloud-go/tea/tea"

	"github.com/wuekevin/smartcdn/internal/domain/port"
	"github.com/wuekevin/smartcdn/internal/infra/ratelimit"
)

const (
	aliyunEndpoint    = "alidns.aliyuncs.com"
	aliyunRateLimit   = 300 // calls per minute (阿里云 DNS API 限制)
	defaultPageSize   = 500
)

// Credentials holds the JSON structure stored in dns_accounts.credentials_enc.
type Credentials struct {
	AccessKeyID     string `json:"access_key_id"`
	AccessKeySecret string `json:"access_key_secret"`
	// PlanLevel: "free" | "individual" | "enterprise_basic" | "enterprise_standard" | "enterprise_ultimate"
	// Optional; if set, used to override capability detection.
	PlanLevel string `json:"plan_level,omitempty"`
}

// AliyunDNSProvider implements port.DNSProvider for 阿里云 DNS.
type AliyunDNSProvider struct {
	client  *alidns.Client
	creds   Credentials
	limiter *ratelimit.TokenBucket
}

// NewAliyunDNSProvider creates a provider from raw credential JSON.
// This is the factory function registered with the ProviderRegistry.
func NewAliyunDNSProvider(credentialsJSON []byte) (port.DNSProvider, error) {
	var creds Credentials
	if err := json.Unmarshal(credentialsJSON, &creds); err != nil {
		return nil, fmt.Errorf("aliyun: invalid credentials JSON: %w", err)
	}
	if creds.AccessKeyID == "" || creds.AccessKeySecret == "" {
		return nil, fmt.Errorf("aliyun: access_key_id and access_key_secret are required")
	}

	cfg := &openapi.Config{
		AccessKeyId:     tea.String(creds.AccessKeyID),
		AccessKeySecret: tea.String(creds.AccessKeySecret),
		Endpoint:        tea.String(aliyunEndpoint),
	}
	client, err := alidns.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("aliyun: create client: %w", err)
	}

	return &AliyunDNSProvider{
		client:  client,
		creds:   creds,
		limiter: ratelimit.NewTokenBucket("aliyun-dns", aliyunRateLimit),
	}, nil
}

func (p *AliyunDNSProvider) ProviderType() port.ProviderType {
	return port.ProviderAliyun
}

func (p *AliyunDNSProvider) ValidateCredentials(ctx context.Context) error {
	if err := p.limiter.Wait(ctx); err != nil {
		return err
	}
	req := &alidns.DescribeDomainInfoRequest{DomainName: tea.String("example.com")}
	_, err := p.client.DescribeDomainInfo(req)
	// 凭证有效时，对不存在域名的 404 错误也算验证通过
	if err != nil {
		if isNotFoundOrForbidden(err) {
			return nil // credentials are valid, domain just doesn't exist in account
		}
		if isAuthError(err) {
			return fmt.Errorf("invalid credentials: %w", err)
		}
		// 对于其他错误，认为凭证有效（网络超时等）
		return nil
	}
	return nil
}

func (p *AliyunDNSProvider) DetectCapabilities(ctx context.Context) (*port.Capabilities, error) {
	// 能力基于套餐档次，通过 DescribeUserInfo 获取套餐信息
	if err := p.limiter.Wait(ctx); err != nil {
		return nil, err
	}

	// 阿里云默认能力（免费版）
	caps := &port.Capabilities{
		ISPRouting:       false,
		Lines:            []string{"default"},
		MinTTL:           600,
		MaxRecordsPerSet: 10,
		RawPlanName:      "free",
	}

	// 若凭证中指定了套餐级别，使用指定值
	if p.creds.PlanLevel != "" {
		fillCapsByPlan(caps, p.creds.PlanLevel)
		return caps, nil
	}

	// 尝试通过 API 检测当前套餐
	req := &alidns.DescribeSupportLinesRequest{DomainName: tea.String("example.com")}
	resp, err := p.client.DescribeSupportLines(req)
	if err == nil && resp.Body != nil && resp.Body.RecordLines != nil {
		lines := make([]string, 0, len(resp.Body.RecordLines.RecordLine))
		for _, l := range resp.Body.RecordLines.RecordLine {
			if l.LineCode != nil {
				lines = append(lines, *l.LineCode)
			}
		}
		if len(lines) > 1 {
			caps.ISPRouting = true
			caps.Lines = lines
			caps.MinTTL = 1
			caps.RawPlanName = "paid"
		}
	}

	return caps, nil
}

func (p *AliyunDNSProvider) ListZones(ctx context.Context) ([]port.DNSZone, error) {
	var zones []port.DNSZone
	pageNum := int64(1)
	for {
		if err := p.limiter.Wait(ctx); err != nil {
			return nil, err
		}
		req := &alidns.DescribeDomainsRequest{
			PageNumber: tea.Int64(pageNum),
			PageSize:   tea.Int64(defaultPageSize),
		}
		resp, err := p.client.DescribeDomains(req)
		if err != nil {
			return nil, fmt.Errorf("aliyun ListZones: %w", err)
		}
		if resp.Body == nil || resp.Body.Domains == nil {
			break
		}
		for _, d := range resp.Body.Domains.Domain {
			zones = append(zones, port.DNSZone{
				ID:     tea.StringValue(d.DomainId),
				Domain: tea.StringValue(d.DomainName),
			})
		}
		total := tea.Int64Value(resp.Body.TotalCount)
		if int64(len(zones)) >= total {
			break
		}
		pageNum++
	}
	return zones, nil
}

func (p *AliyunDNSProvider) ListRecords(ctx context.Context, zoneID string) ([]port.DNSRecord, error) {
	// 阿里云 API 使用域名而非 zoneID 查询记录
	domain := zoneID
	var records []port.DNSRecord
	pageNum := int64(1)
	for {
		if err := p.limiter.Wait(ctx); err != nil {
			return nil, err
		}
		req := &alidns.DescribeDomainRecordsRequest{
			DomainName: tea.String(domain),
			PageNumber: tea.Int64(pageNum),
			PageSize:   tea.Int64(defaultPageSize),
		}
		resp, err := p.client.DescribeDomainRecords(req)
		if err != nil {
			return nil, fmt.Errorf("aliyun ListRecords: %w", err)
		}
		if resp.Body == nil || resp.Body.DomainRecords == nil {
			break
		}
		for _, r := range resp.Body.DomainRecords.Record {
			records = append(records, recordFromAliyun(r))
		}
		total := tea.Int64Value(resp.Body.TotalCount)
		if int64(len(records)) >= total {
			break
		}
		pageNum++
	}
	return records, nil
}

func (p *AliyunDNSProvider) CreateRecord(ctx context.Context, zoneID string, input port.DNSRecordInput) (string, error) {
	if err := p.limiter.Wait(ctx); err != nil {
		return "", err
	}
	req := &alidns.AddDomainRecordRequest{
		DomainName: tea.String(zoneID),
		RR:         tea.String(hostToRR(input.Host)),
		Type:       tea.String(input.Type),
		Value:      tea.String(input.Value),
		TTL:        tea.Int64(int64(input.TTL)),
		Line:       tea.String(lineOrDefault(input.Line)),
	}
	if input.Priority > 0 {
		req.Priority = tea.Int64(int64(input.Priority))
	}
	resp, err := p.client.AddDomainRecord(req)
	if err != nil {
		return "", fmt.Errorf("aliyun CreateRecord: %w", err)
	}
	return tea.StringValue(resp.Body.RecordId), nil
}

func (p *AliyunDNSProvider) UpdateRecord(ctx context.Context, _, recordID string, input port.DNSRecordInput) error {
	if err := p.limiter.Wait(ctx); err != nil {
		return err
	}
	req := &alidns.UpdateDomainRecordRequest{
		RecordId: tea.String(recordID),
		RR:       tea.String(hostToRR(input.Host)),
		Type:     tea.String(input.Type),
		Value:    tea.String(input.Value),
		TTL:      tea.Int64(int64(input.TTL)),
		Line:     tea.String(lineOrDefault(input.Line)),
	}
	if input.Priority > 0 {
		req.Priority = tea.Int64(int64(input.Priority))
	}
	_, err := p.client.UpdateDomainRecord(req)
	return err
}

func (p *AliyunDNSProvider) DeleteRecord(ctx context.Context, _, recordID string) error {
	if err := p.limiter.Wait(ctx); err != nil {
		return err
	}
	_, err := p.client.DeleteDomainRecord(&alidns.DeleteDomainRecordRequest{RecordId: tea.String(recordID)})
	return err
}

func (p *AliyunDNSProvider) CreateBatch(ctx context.Context, zoneID string, inputs []port.DNSRecordInput) ([]port.BatchResult, error) {
	results := make([]port.BatchResult, len(inputs))
	for i, input := range inputs {
		id, err := p.CreateRecord(ctx, zoneID, input)
		results[i] = port.BatchResult{RecordID: id, Err: err}
	}
	return results, nil
}

func (p *AliyunDNSProvider) GetRecord(ctx context.Context, _, recordID string) (*port.DNSRecord, error) {
	if err := p.limiter.Wait(ctx); err != nil {
		return nil, err
	}
	resp, err := p.client.DescribeDomainRecordInfo(&alidns.DescribeDomainRecordInfoRequest{RecordId: tea.String(recordID)})
	if err != nil {
		return nil, fmt.Errorf("aliyun GetRecord: %w", err)
	}
	if resp.Body == nil {
		return nil, fmt.Errorf("aliyun GetRecord: empty response")
	}
	r := port.DNSRecord{
		ID:    tea.StringValue(resp.Body.RecordId),
		Host:  tea.StringValue(resp.Body.RR),
		Type:  tea.StringValue(resp.Body.Type),
		Value: tea.StringValue(resp.Body.Value),
		TTL:   int(tea.Int64Value(resp.Body.TTL)),
		Line:  tea.StringValue(resp.Body.Line),
	}
	return &r, nil
}

func (p *AliyunDNSProvider) FindRecords(ctx context.Context, zoneID, host, recordType, line string) ([]port.DNSRecord, error) {
	all, err := p.ListRecords(ctx, zoneID)
	if err != nil {
		return nil, err
	}
	var out []port.DNSRecord
	for _, r := range all {
		if host != "" && r.Host != host {
			continue
		}
		if recordType != "" && !strings.EqualFold(r.Type, recordType) {
			continue
		}
		if line != "" && r.Line != line {
			continue
		}
		out = append(out, r)
	}
	return out, nil
}

func (p *AliyunDNSProvider) PauseRecord(ctx context.Context, _, recordID string) error {
	if err := p.limiter.Wait(ctx); err != nil {
		return err
	}
	_, err := p.client.SetDomainRecordStatus(&alidns.SetDomainRecordStatusRequest{
		RecordId: tea.String(recordID),
		Status:   tea.String("Disable"),
	})
	return err
}

func (p *AliyunDNSProvider) ResumeRecord(ctx context.Context, _, recordID string) error {
	if err := p.limiter.Wait(ctx); err != nil {
		return err
	}
	_, err := p.client.SetDomainRecordStatus(&alidns.SetDomainRecordStatusRequest{
		RecordId: tea.String(recordID),
		Status:   tea.String("Enable"),
	})
	return err
}

func (p *AliyunDNSProvider) GetNameservers(ctx context.Context, zoneID string) ([]string, error) {
	if err := p.limiter.Wait(ctx); err != nil {
		return nil, err
	}
	resp, err := p.client.DescribeDomainInfo(&alidns.DescribeDomainInfoRequest{DomainName: tea.String(zoneID)})
	if err != nil {
		return nil, err
	}
	if resp.Body == nil || resp.Body.DnsServers == nil {
		return nil, nil
	}
	var ns []string
	for _, s := range resp.Body.DnsServers.DnsServer {
		if s != nil {
			ns = append(ns, *s)
		}
	}
	return ns, nil
}

// --- helpers ---

func recordFromAliyun(r *alidns.DescribeDomainRecordsResponseBodyDomainRecordsRecord) port.DNSRecord {
	rec := port.DNSRecord{
		ID:    tea.StringValue(r.RecordId),
		Host:  tea.StringValue(r.RR),
		Type:  tea.StringValue(r.Type),
		Value: tea.StringValue(r.Value),
		TTL:   int(tea.Int64Value(r.TTL)),
		Line:  tea.StringValue(r.Line),
	}
	if r.Priority != nil {
		rec.Priority = int(*r.Priority)
	}
	if r.UpdateTimestamp != nil {
		ts := tea.Int64Value(r.UpdateTimestamp)
		rec.UpdatedAt = time.Unix(ts/1000, (ts%1000)*1e6)
	}
	return rec
}

func hostToRR(host string) string {
	if host == "@" || host == "" {
		return "@"
	}
	return host
}

func lineOrDefault(line string) string {
	if line == "" {
		return "default"
	}
	return line
}

func isAuthError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "InvalidAccessKeyId") ||
		strings.Contains(msg, "SignatureDoesNotMatch") ||
		strings.Contains(msg, "Forbidden")
}

func isNotFoundOrForbidden(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "InvalidDomainName.NoExist") ||
		strings.Contains(msg, "DomainNotFound")
}

func fillCapsByPlan(caps *port.Capabilities, plan string) {
	switch strings.ToLower(plan) {
	case "individual", "enterprise_basic", "enterprise_standard", "enterprise_ultimate":
		caps.ISPRouting = true
		caps.Lines = []string{"default", "telecom", "unicom", "mobile", "overseas"}
		caps.MinTTL = 1
		caps.MaxRecordsPerSet = 50
		caps.RawPlanName = plan
	}
}
