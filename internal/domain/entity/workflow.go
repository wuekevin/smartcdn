package entity

import "time"

// WorkflowNodeType is the kind of a DAG workflow node.
type WorkflowNodeType string

const (
	WFNodeAcmeIssue     WorkflowNodeType = "acme_issue"
	WFNodeDeploy        WorkflowNodeType = "deploy_cert"
	WFNodeDNSValidate   WorkflowNodeType = "dns_validate"
	WFNodeNotify        WorkflowNodeType = "notify"
	WFNodeCondition     WorkflowNodeType = "condition"
	WFNodeWait          WorkflowNodeType = "wait"
)

// WorkflowStatus is the lifecycle state of a workflow run.
type WorkflowStatus string

const (
	WFStatusPending   WorkflowStatus = "pending"
	WFStatusRunning   WorkflowStatus = "running"
	WFStatusSucceeded WorkflowStatus = "succeeded"
	WFStatusFailed    WorkflowStatus = "failed"
	WFStatusCancelled WorkflowStatus = "cancelled"
)

// WorkflowNodeStatus is the state of a single node execution.
type WorkflowNodeStatus string

const (
	WFNodePending   WorkflowNodeStatus = "pending"
	WFNodeRunning   WorkflowNodeStatus = "running"
	WFNodeSucceeded WorkflowNodeStatus = "succeeded"
	WFNodeFailed    WorkflowNodeStatus = "failed"
	WFNodeSkipped   WorkflowNodeStatus = "skipped"
)

// WorkflowTemplate is a reusable DAG definition (stored as JSON).
type WorkflowTemplate struct {
	ID          uint64 `gorm:"primarykey;autoIncrement"`
	Name        string `gorm:"size:128;not null;uniqueIndex"`
	Description string `gorm:"size:512"`
	// GraphJSON is the DAG definition: JSON array of WorkflowNodeDef.
	GraphJSON   string    `gorm:"type:longtext;not null"`
	IsActive    bool      `gorm:"not null;default:true"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time `gorm:"index"`
}

// WorkflowRun is one execution of a WorkflowTemplate.
type WorkflowRun struct {
	ID              uint64         `gorm:"primarykey;autoIncrement"`
	TemplateID      uint64         `gorm:"not null;index"`
	TemplateName    string         `gorm:"size:128;not null"` // snapshot at start
	TriggerType     string         `gorm:"size:64;not null"`  // manual / auto_renew / cert_deploy
	TriggerRef      string         `gorm:"size:255"`          // e.g. certificate ID
	Status          WorkflowStatus `gorm:"size:32;not null;default:'pending'"`
	ContextJSON     string         `gorm:"type:text"` // key-value pairs passed through nodes
	StartedAt       *time.Time
	FinishedAt      *time.Time
	LastError       string    `gorm:"type:text"`
	RetryCount      int       `gorm:"default:0"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// WorkflowNodeRun is the execution record for one node in a WorkflowRun.
type WorkflowNodeRun struct {
	ID         uint64             `gorm:"primarykey;autoIncrement"`
	RunID      uint64             `gorm:"not null;index"`
	NodeID     string             `gorm:"size:128;not null"` // matches WorkflowNodeDef.ID
	NodeType   WorkflowNodeType   `gorm:"size:64;not null"`
	Status     WorkflowNodeStatus `gorm:"size:32;not null;default:'pending'"`
	InputJSON  string             `gorm:"type:text"`
	OutputJSON string             `gorm:"type:text"`
	StartedAt  *time.Time
	FinishedAt *time.Time
	Error      string    `gorm:"type:text"`
	RetryCount int       `gorm:"default:0"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
