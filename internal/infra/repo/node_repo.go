package repo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/wuekevin/smartcdn/internal/domain/entity"
)

// CDNNodeRepo ─────────────────────────────────────────────────────────────

type CDNNodeRepo struct{ db *gorm.DB }

func NewCDNNodeRepo(db *gorm.DB) *CDNNodeRepo { return &CDNNodeRepo{db: db} }

func (r *CDNNodeRepo) Create(ctx context.Context, n *entity.CDNNode) error {
	return r.db.WithContext(ctx).Create(n).Error
}

func (r *CDNNodeRepo) FindByID(ctx context.Context, id uint64) (*entity.CDNNode, error) {
	var n entity.CDNNode
	if err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&n).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("CDNNodeRepo.FindByID: %w", err)
	}
	return &n, nil
}

func (r *CDNNodeRepo) FindByIP(ctx context.Context, ip string) (*entity.CDNNode, error) {
	var n entity.CDNNode
	if err := r.db.WithContext(ctx).Where("ip = ? AND deleted_at IS NULL", ip).First(&n).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("CDNNodeRepo.FindByIP: %w", err)
	}
	return &n, nil
}

func (r *CDNNodeRepo) FindByToken(ctx context.Context, token string) (*entity.CDNNode, error) {
	var n entity.CDNNode
	if err := r.db.WithContext(ctx).Where("auth_token = ? AND deleted_at IS NULL", token).First(&n).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("CDNNodeRepo.FindByToken: %w", err)
	}
	return &n, nil
}

func (r *CDNNodeRepo) List(ctx context.Context, afterID uint64, limit int) ([]entity.CDNNode, error) {
	if limit <= 0 {
		limit = 20
	}
	var ns []entity.CDNNode
	q := r.db.WithContext(ctx).Where("deleted_at IS NULL")
	if afterID > 0 {
		q = q.Where("id > ?", afterID)
	}
	return ns, q.Order("id ASC").Limit(limit).Find(&ns).Error
}

func (r *CDNNodeRepo) Update(ctx context.Context, n *entity.CDNNode) error {
	return r.db.WithContext(ctx).Save(n).Error
}

func (r *CDNNodeRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&entity.CDNNode{}).Error
}

// NodeAgentRepo ───────────────────────────────────────────────────────────

type NodeAgentRepo struct{ db *gorm.DB }

func NewNodeAgentRepo(db *gorm.DB) *NodeAgentRepo { return &NodeAgentRepo{db: db} }

func (r *NodeAgentRepo) Upsert(ctx context.Context, a *entity.NodeAgent) error {
	return r.db.WithContext(ctx).Save(a).Error
}

func (r *NodeAgentRepo) FindByNodeID(ctx context.Context, nodeID uint64) (*entity.NodeAgent, error) {
	var a entity.NodeAgent
	if err := r.db.WithContext(ctx).Where("node_id = ?", nodeID).First(&a).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("NodeAgentRepo.FindByNodeID: %w", err)
	}
	return &a, nil
}

func (r *NodeAgentRepo) FindStaleBefore(ctx context.Context, threshold time.Time, limit int) ([]entity.NodeAgent, error) {
	var agents []entity.NodeAgent
	return agents, r.db.WithContext(ctx).
		Where("last_seen_at < ? OR last_seen_at IS NULL", threshold).
		Limit(limit).Find(&agents).Error
}

// NodeTaskRepo ────────────────────────────────────────────────────────────

type NodeTaskRepo struct{ db *gorm.DB }

func NewNodeTaskRepo(db *gorm.DB) *NodeTaskRepo { return &NodeTaskRepo{db: db} }

func (r *NodeTaskRepo) Create(ctx context.Context, t *entity.NodeTask) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *NodeTaskRepo) FindByID(ctx context.Context, id uint64) (*entity.NodeTask, error) {
	var t entity.NodeTask
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&t).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("NodeTaskRepo.FindByID: %w", err)
	}
	return &t, nil
}

func (r *NodeTaskRepo) ListPendingForNode(ctx context.Context, nodeID uint64, limit int) ([]entity.NodeTask, error) {
	if limit <= 0 {
		limit = 10
	}
	var tasks []entity.NodeTask
	return tasks, r.db.WithContext(ctx).
		Where("node_id = ? AND status = 'pending'", nodeID).
		Order("id ASC").Limit(limit).Find(&tasks).Error
}

func (r *NodeTaskRepo) Update(ctx context.Context, t *entity.NodeTask) error {
	return r.db.WithContext(ctx).Save(t).Error
}

// NodePoolRepo ────────────────────────────────────────────────────────────

type NodePoolRepo struct{ db *gorm.DB }

func NewNodePoolRepo(db *gorm.DB) *NodePoolRepo { return &NodePoolRepo{db: db} }

func (r *NodePoolRepo) Create(ctx context.Context, p *entity.NodePool) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *NodePoolRepo) FindByID(ctx context.Context, id uint64) (*entity.NodePool, error) {
	var p entity.NodePool
	if err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("NodePoolRepo.FindByID: %w", err)
	}
	return &p, nil
}

func (r *NodePoolRepo) List(ctx context.Context, afterID uint64, limit int) ([]entity.NodePool, error) {
	if limit <= 0 {
		limit = 20
	}
	var pools []entity.NodePool
	q := r.db.WithContext(ctx).Where("deleted_at IS NULL")
	if afterID > 0 {
		q = q.Where("id > ?", afterID)
	}
	return pools, q.Order("id ASC").Limit(limit).Find(&pools).Error
}

func (r *NodePoolRepo) Update(ctx context.Context, p *entity.NodePool) error {
	return r.db.WithContext(ctx).Save(p).Error
}

func (r *NodePoolRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&entity.NodePool{}).Error
}

func (r *NodePoolRepo) AddMember(ctx context.Context, m *entity.NodePoolMember) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *NodePoolRepo) RemoveMember(ctx context.Context, poolID, nodeID uint64) error {
	return r.db.WithContext(ctx).
		Where("pool_id = ? AND node_id = ?", poolID, nodeID).
		Delete(&entity.NodePoolMember{}).Error
}

func (r *NodePoolRepo) ListMembers(ctx context.Context, poolID uint64) ([]entity.NodePoolMember, error) {
	var members []entity.NodePoolMember
	return members, r.db.WithContext(ctx).
		Where("pool_id = ?", poolID).Order("priority DESC, id ASC").Find(&members).Error
}
