package port

import (
	"context"

	"github.com/wuekevin/smartcdn/internal/domain/entity"
)

// PublishJobRepository persists DNS publish jobs.
type PublishJobRepository interface {
	Create(ctx context.Context, job *entity.PublishJob) error
	FindByID(ctx context.Context, id uint64) (*entity.PublishJob, error)
	Update(ctx context.Context, job *entity.PublishJob) error
	List(ctx context.Context, afterID uint64, limit int) ([]entity.PublishJob, error)
	ListByDomain(ctx context.Context, domainID uint64, afterID uint64, limit int) ([]entity.PublishJob, error)
}
