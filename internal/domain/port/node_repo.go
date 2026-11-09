package port

import (
	"context"
	"time"

	"github.com/wuekevin/smartcdn/internal/domain/entity"
)

// CDNNodeRepository abstracts persistence for self-hosted CDN nodes.
type CDNNodeRepository interface {
	Create(ctx context.Context, n *entity.CDNNode) error
	FindByID(ctx context.Context, id uint64) (*entity.CDNNode, error)
	FindByIP(ctx context.Context, ip string) (*entity.CDNNode, error)
	FindByToken(ctx context.Context, token string) (*entity.CDNNode, error)
	List(ctx context.Context, afterID uint64, limit int) ([]entity.CDNNode, error)
	Update(ctx context.Context, n *entity.CDNNode) error
	Delete(ctx context.Context, id uint64) error
}

// NodeAgentRepository abstracts persistence for Node Agent registrations.
type NodeAgentRepository interface {
	Upsert(ctx context.Context, a *entity.NodeAgent) error
	FindByNodeID(ctx context.Context, nodeID uint64) (*entity.NodeAgent, error)
	FindStaleBefore(ctx context.Context, threshold time.Time, limit int) ([]entity.NodeAgent, error)
}

// NodeTaskRepository abstracts persistence for management tasks.
type NodeTaskRepository interface {
	Create(ctx context.Context, t *entity.NodeTask) error
	FindByID(ctx context.Context, id uint64) (*entity.NodeTask, error)
	ListPendingForNode(ctx context.Context, nodeID uint64, limit int) ([]entity.NodeTask, error)
	Update(ctx context.Context, t *entity.NodeTask) error
}

// NodePoolRepository abstracts persistence for node pools.
type NodePoolRepository interface {
	Create(ctx context.Context, p *entity.NodePool) error
	FindByID(ctx context.Context, id uint64) (*entity.NodePool, error)
	List(ctx context.Context, afterID uint64, limit int) ([]entity.NodePool, error)
	Update(ctx context.Context, p *entity.NodePool) error
	Delete(ctx context.Context, id uint64) error

	AddMember(ctx context.Context, m *entity.NodePoolMember) error
	RemoveMember(ctx context.Context, poolID, nodeID uint64) error
	ListMembers(ctx context.Context, poolID uint64) ([]entity.NodePoolMember, error)
}
