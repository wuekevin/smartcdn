package repo

import "time"

type AuditFilter struct {
	OperatorID uint64
	StartTime  time.Time
	EndTime    time.Time
	AfterID    uint64 // cursor-based pagination: entries with id < AfterID
	Limit      int
}
