package repo

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/wuekevin/smartcdn/internal/domain/entity"
	"github.com/wuekevin/smartcdn/internal/domain/port"
)

// SSLCredentialRepo 实现 port.SSLCredentialRepository。
type SSLCredentialRepo struct{ db *gorm.DB }

func NewSSLCredentialRepo(db *gorm.DB) *SSLCredentialRepo { return &SSLCredentialRepo{db: db} }

func (r *SSLCredentialRepo) Create(ctx context.Context, c *entity.SSLCredential) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *SSLCredentialRepo) FindByID(ctx context.Context, id uint64) (*entity.SSLCredential, error) {
	var c entity.SSLCredential
	err := r.db.WithContext(ctx).Where("deleted_at IS NULL").First(&c, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &c, err
}

func (r *SSLCredentialRepo) List(ctx context.Context, category string) ([]entity.SSLCredential, error) {
	var items []entity.SSLCredential
	q := r.db.WithContext(ctx).Where("deleted_at IS NULL")
	if category != "" {
		q = q.Where("category = ?", category)
	}
	return items, q.Order("id DESC").Find(&items).Error
}

func (r *SSLCredentialRepo) Update(ctx context.Context, c *entity.SSLCredential) error {
	return r.db.WithContext(ctx).Save(c).Error
}

func (r *SSLCredentialRepo) Delete(ctx context.Context, id uint64) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&entity.SSLCredential{}).
		Where("id = ?", id).Update("deleted_at", now).Error
}

var _ port.SSLCredentialRepository = (*SSLCredentialRepo)(nil)

// SSLDeployJobRepo 实现 port.SSLDeployJobRepository。
type SSLDeployJobRepo struct{ db *gorm.DB }

func NewSSLDeployJobRepo(db *gorm.DB) *SSLDeployJobRepo { return &SSLDeployJobRepo{db: db} }

func (r *SSLDeployJobRepo) Create(ctx context.Context, j *entity.SSLDeployJob) error {
	return r.db.WithContext(ctx).Create(j).Error
}

func (r *SSLDeployJobRepo) FindByID(ctx context.Context, id uint64) (*entity.SSLDeployJob, error) {
	var j entity.SSLDeployJob
	err := r.db.WithContext(ctx).First(&j, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &j, err
}

func (r *SSLDeployJobRepo) ListByCert(ctx context.Context, certID uint64) ([]entity.SSLDeployJob, error) {
	var jobs []entity.SSLDeployJob
	return jobs, r.db.WithContext(ctx).Where("cert_id = ?", certID).Order("id DESC").Find(&jobs).Error
}

func (r *SSLDeployJobRepo) Update(ctx context.Context, j *entity.SSLDeployJob) error {
	return r.db.WithContext(ctx).Save(j).Error
}

func (r *SSLDeployJobRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&entity.SSLDeployJob{}, id).Error
}

var _ port.SSLDeployJobRepository = (*SSLDeployJobRepo)(nil)
