package dnspod

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	dnspodSDK "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dnspod/v20210323"

	"github.com/wuekevin/smartcdn/internal/domain/port"
	"github.com/wuekevin/smartcdn/internal/infra/ratelimit"
)

const (
	dnspodEndpoint  = "dnspod.tencentcloudapi.com"
	dnspodRateLimit = 300
)

// Credentials holds the JSON structure for DNSPod accounts.
type Credentials struct {
	SecretID  string `json:"secret_id"`
	SecretKey string `json:"secret_key"`
}

// DNSPodProvider implements port.DNSProvider for 腾讯云 DNSPod.
type DNSPodProvider struct {
	client  *dnspodSDK.Client
	limiter *ratelimit.TokenBucket
}

// NewDNSPodProvider creates a provider from raw credential JSON.
func NewDNSPodProvider(credentialsJSON []byte) (port.DNSProvider, error) {
	var creds Credentials
	if err := json.Unmarshal(credentialsJSON, &creds); err != nil {
		return nil, fmt.Errorf("dnspod: invalid credentials JSON: %w", err)
	}
	if creds.SecretID == "" || creds.SecretKey == "" {
		return nil, fmt.Errorf("dnspod: secret_id and secret_key are required")
	}

	credential := common.NewCredential(creds.SecretID, creds.SecretKey)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = dnspodEndpoint
	client, err := dnspodSDK.NewClient(credential, "", cpf)
	if err != nil {
		return nil, fmt.Errorf("dnspod: create client: %w", err)
	}

	return &DNSPodProvider{
		client:  client,
		limiter: ratelimit.NewTokenBucket("dnspod", dnspodRateLimit),
	}, nil
}

func (p *DNSPodProvider) ProviderType() port.ProviderType {
	return port.ProviderDNSPod
}

func (p *DNSPodProvider) ValidateCredentials(ctx context.Context) error {
	if err := p.limiter.Wait(ctx); err != nil {
		return err
	}
	req := dnspodSDK.NewDescribeDomainListRequest()
	offset := int64(0)
	limit := int64(1)
	req.Offset = &offset
	req.Limit = &limit
	_, err := p.client.DescribeDomainList(req)
	if err != nil {
		if isAuthError(err) {
			return fmt.Errorf("invalid credentials: %w", err)
		}
	}
	return nil
}

func (p *DNSPodProvider) DetectCapabilities(ctx context.Context) (*port.Capabilities, error) {
	caps := &port.Capabilities{
		ISPRouting:       false,
		Lines:            []string{"默认"},
		MinTTL:           600,
		MaxRecordsPerSet: 2,
		RawPlanName:      "free",
	}

	if err := p.limiter.Wait(ctx); err != nil {
		return nil, err
	}

	// 检测第一个域名的套餐
	req := dnspodSDK.NewDescribeDomainListRequest()
	offset := int64(0)
	limit := int64(1)
	req.Offset = &offset
	req.Limit = &limit
	resp, err := p.client.DescribeDomainList(req)
	if err == nil && resp.Response != nil && len(resp.Response.DomainList) > 0 {
		domain := resp.Response.DomainList[0]
		if domain.GradeTitle != nil && isPaidPlan(*domain.GradeTitle) {
			caps.ISPRouting = true
			caps.Lines = []string{"默认", "电信", "联通", "移动", "境外"}
			caps.MinTTL = 1
			caps.MaxRecordsPerSet = 20
			caps.RawPlanName = *domain.GradeTitle
		}
	}

	return caps, nil
}

func (p *DNSPodProvider) ListZones(ctx context.Context) ([]port.DNSZone, error) {
	var zones []port.DNSZone
	offset := int64(0)
	limit := int64(3000)
	for {
		if err := p.limiter.Wait(ctx); err != nil {
			return nil, err
		}
		req := dnspodSDK.NewDescribeDomainListRequest()
		req.Offset = &offset
		req.Limit = &limit
		resp, err := p.client.DescribeDomainList(req)
		if err != nil {
			return nil, fmt.Errorf("dnspod ListZones: %w", err)
		}
		if resp.Response == nil || len(resp.Response.DomainList) == 0 {
			break
		}
		for _, d := range resp.Response.DomainList {
			zone := port.DNSZone{}
			if d.DomainId != nil {
				zone.ID = fmt.Sprintf("%d", *d.DomainId)
			}
			if d.Name != nil {
				zone.Domain = *d.Name
			}
			zones = append(zones, zone)
		}
		if resp.Response.DomainCountInfo == nil || resp.Response.DomainCountInfo.DomainTotal == nil {
			break
		}
		total := int64(*resp.Response.DomainCountInfo.DomainTotal)
		offset += int64(len(resp.Response.DomainList))
		if offset >= total {
			break
		}
	}
	return zones, nil
}

func (p *DNSPodProvider) ListRecords(ctx context.Context, zoneID string) ([]port.DNSRecord, error) {
	var records []port.DNSRecord
	offset := uint64(0)
	limit := uint64(3000)
	for {
		if err := p.limiter.Wait(ctx); err != nil {
			return nil, err
		}
		req := dnspodSDK.NewDescribeRecordListRequest()
		req.Domain = ptrString(zoneID)
		req.Offset = &offset
		req.Limit = &limit
		resp, err := p.client.DescribeRecordList(req)
		if err != nil {
			return nil, fmt.Errorf("dnspod ListRecords: %w", err)
		}
		if resp.Response == nil || len(resp.Response.RecordList) == 0 {
			break
		}
		for _, r := range resp.Response.RecordList {
			records = append(records, recordFromDNSPod(r))
		}
		if resp.Response.RecordCountInfo == nil || resp.Response.RecordCountInfo.TotalCount == nil {
			break
		}
		total := *resp.Response.RecordCountInfo.TotalCount
		offset += uint64(len(resp.Response.RecordList))
		if offset >= total {
			break
		}
	}
	return records, nil
}

func (p *DNSPodProvider) CreateRecord(ctx context.Context, zoneID string, input port.DNSRecordInput) (string, error) {
	if err := p.limiter.Wait(ctx); err != nil {
		return "", err
	}
	req := dnspodSDK.NewCreateRecordRequest()
	req.Domain = ptrString(zoneID)
	req.SubDomain = ptrString(input.Host)
	req.RecordType = ptrString(input.Type)
	req.Value = ptrString(input.Value)
	ttl := uint64(input.TTL)
	req.TTL = &ttl
	req.RecordLine = ptrString(lineOrDefault(input.Line))
	if input.Priority > 0 {
		mx := uint64(input.Priority)
		req.MX = &mx
	}
	resp, err := p.client.CreateRecord(req)
	if err != nil {
		return "", fmt.Errorf("dnspod CreateRecord: %w", err)
	}
	if resp.Response == nil || resp.Response.RecordId == nil {
		return "", fmt.Errorf("dnspod CreateRecord: empty response")
	}
	return fmt.Sprintf("%d", *resp.Response.RecordId), nil
}

func (p *DNSPodProvider) UpdateRecord(ctx context.Context, zoneID, recordID string, input port.DNSRecordInput) error {
	if err := p.limiter.Wait(ctx); err != nil {
		return err
	}
	id, err := parseUint64(recordID)
	if err != nil {
		return err
	}
	req := dnspodSDK.NewModifyRecordRequest()
	req.Domain = ptrString(zoneID)
	req.RecordId = &id
	req.SubDomain = ptrString(input.Host)
	req.RecordType = ptrString(input.Type)
	req.Value = ptrString(input.Value)
	ttl := uint64(input.TTL)
	req.TTL = &ttl
	req.RecordLine = ptrString(lineOrDefault(input.Line))
	_, err = p.client.ModifyRecord(req)
	return err
}

func (p *DNSPodProvider) DeleteRecord(ctx context.Context, zoneID, recordID string) error {
	if err := p.limiter.Wait(ctx); err != nil {
		return err
	}
	id, err := parseUint64(recordID)
	if err != nil {
		return err
	}
	req := dnspodSDK.NewDeleteRecordRequest()
	req.Domain = ptrString(zoneID)
	req.RecordId = &id
	_, err = p.client.DeleteRecord(req)
	return err
}

func (p *DNSPodProvider) CreateBatch(ctx context.Context, zoneID string, inputs []port.DNSRecordInput) ([]port.BatchResult, error) {
	results := make([]port.BatchResult, len(inputs))
	for i, input := range inputs {
		id, err := p.CreateRecord(ctx, zoneID, input)
		results[i] = port.BatchResult{RecordID: id, Err: err}
	}
	return results, nil
}

func (p *DNSPodProvider) GetRecord(ctx context.Context, zoneID, recordID string) (*port.DNSRecord, error) {
	if err := p.limiter.Wait(ctx); err != nil {
		return nil, err
	}
	id, err := parseUint64(recordID)
	if err != nil {
		return nil, err
	}
	req := dnspodSDK.NewDescribeRecordRequest()
	req.Domain = ptrString(zoneID)
	req.RecordId = &id
	resp, err := p.client.DescribeRecord(req)
	if err != nil {
		return nil, err
	}
	if resp.Response == nil || resp.Response.RecordInfo == nil {
		return nil, fmt.Errorf("dnspod GetRecord: empty response")
	}
	ri := resp.Response.RecordInfo
	r := port.DNSRecord{
		ID:    recordID,
		Type:  derefString(ri.RecordType),
		Value: derefString(ri.Value),
		Line:  derefString(ri.RecordLine),
	}
	if ri.SubDomain != nil {
		r.Host = *ri.SubDomain
	}
	if ri.TTL != nil {
		r.TTL = int(*ri.TTL)
	}
	return &r, nil
}

func (p *DNSPodProvider) FindRecords(ctx context.Context, zoneID, host, recordType, line string) ([]port.DNSRecord, error) {
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

func (p *DNSPodProvider) PauseRecord(ctx context.Context, zoneID, recordID string) error {
	if err := p.limiter.Wait(ctx); err != nil {
		return err
	}
	id, err := parseUint64(recordID)
	if err != nil {
		return err
	}
	req := dnspodSDK.NewModifyRecordStatusRequest()
	req.Domain = ptrString(zoneID)
	req.RecordId = &id
	req.Status = ptrString("DISABLE")
	_, err = p.client.ModifyRecordStatus(req)
	return err
}

func (p *DNSPodProvider) ResumeRecord(ctx context.Context, zoneID, recordID string) error {
	if err := p.limiter.Wait(ctx); err != nil {
		return err
	}
	id, err := parseUint64(recordID)
	if err != nil {
		return err
	}
	req := dnspodSDK.NewModifyRecordStatusRequest()
	req.Domain = ptrString(zoneID)
	req.RecordId = &id
	req.Status = ptrString("ENABLE")
	_, err = p.client.ModifyRecordStatus(req)
	return err
}

func (p *DNSPodProvider) GetNameservers(_ context.Context, _ string) ([]string, error) {
	return nil, &port.UnsupportedCapabilityError{
		Operation: "GetNameservers",
		Reason:    "not supported by DNSPod API",
	}
}

// --- helpers ---

func recordFromDNSPod(r *dnspodSDK.RecordListItem) port.DNSRecord {
	rec := port.DNSRecord{
		Line: derefString(r.Line),
		Type: derefString(r.Type),
	}
	if r.RecordId != nil {
		rec.ID = fmt.Sprintf("%d", *r.RecordId)
	}
	if r.Name != nil {
		rec.Host = *r.Name
	}
	if r.Value != nil {
		rec.Value = *r.Value
	}
	if r.TTL != nil {
		rec.TTL = int(*r.TTL)
	}
	if r.MX != nil {
		rec.Priority = int(*r.MX)
	}
	return rec
}

func lineOrDefault(line string) string {
	if line == "" {
		return "默认"
	}
	return line
}

func isPaidPlan(grade string) bool {
	g := strings.ToLower(grade)
	return strings.Contains(g, "plus") || strings.Contains(g, "专业") || strings.Contains(g, "企业")
}

func isAuthError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "AuthFailure") || strings.Contains(msg, "InvalidSecretId")
}

func ptrString(s string) *string { return &s }

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func parseUint64(s string) (uint64, error) {
	var n uint64
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}
