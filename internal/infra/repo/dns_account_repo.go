package repo

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/wuekevin/smartcdn/internal/domain/entity"
)

type DNSAccountRepo struct {
	db *gorm.DB
}

func NewDNSAccountRepo(db *gorm.DB) *DNSAccountRepo {
	return &DNSAccountRepo{db: db}
}

func (r *DNSAccountRepo) Create(ctx context.Context, acc *entity.DNSAccount) error {
	return r.db.WithContext(ctx).Create(acc).Error
}

func (r *DNSAccountRepo) FindByID(ctx context.Context, id uint64) (*entity.DNSAccount, error) {
	var acc entity.DNSAccount
	err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&acc).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("DNSAccountRepo.FindByID: %w", err)
	}
	return &acc, nil
}

func (r *DNSAccountRepo) List(ctx context.Context, afterID uint64, limit int) ([]entity.DNSAccount, error) {
	if limit <= 0 {
		limit = 20
	}
	var accs []entity.DNSAccount
	q := r.db.WithContext(ctx).Where("deleted_at IS NULL")
	if afterID > 0 {
		q = q.Where("id > ?", afterID)
	}
	err := q.Order("id ASC").Limit(limit).Find(&accs).Error
	return accs, err
}

func (r *DNSAccountRepo) UpdateCredentials(ctx context.Context, id uint64, enc string) error {
	return r.db.WithContext(ctx).Model(&entity.DNSAccount{}).Where("id = ?", id).
		Update("credentials_enc", enc).Error
}

func (r *DNSAccountRepo) UpdateCapabilities(ctx context.Context, id uint64, capJSON string) error {
	return r.db.WithContext(ctx).Model(&entity.DNSAccount{}).Where("id = ?", id).
		Update("capabilities_json", capJSON).Error
}

func (r *DNSAccountRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&entity.DNSAccount{}).Error
}
