package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	nodeapp "github.com/wuekevin/smartcdn/internal/application/node"
	"github.com/wuekevin/smartcdn/internal/domain/entity"
)

// NodeHandler handles operator-facing node and pool management APIs.
type NodeHandler struct {
	svc *nodeapp.Service
}

func NewNodeHandler(svc *nodeapp.Service) *NodeHandler {
	return &NodeHandler{svc: svc}
}

func (h *NodeHandler) RegisterRoutes(r *gin.RouterGroup) {
	nodes := r.Group("/nodes")
	nodes.POST("", h.CreateNode)
	nodes.GET("", h.ListNodes)
	nodes.GET("/:id", h.GetNode)
	nodes.DELETE("/:id", h.DeleteNode)
	nodes.POST("/:id/https-check", h.RetriggerHTTPSCheck)
	nodes.GET("/:id/install-cmd", h.InstallCommand)
	nodes.POST("/:id/tasks", h.DispatchTask)

	pools := r.Group("/node-pools")
	pools.POST("", h.CreatePool)
	pools.GET("", h.ListPools)
	pools.GET("/:pid", h.GetPool)
	pools.DELETE("/:pid", h.DeletePool)
	pools.POST("/:pid/members", h.AddMember)
	pools.DELETE("/:pid/members/:nid", h.RemoveMember)
}

// --- Node ---

type createNodeReq struct {
	Name          string `json:"name" binding:"required,max=128"`
	IP            string `json:"ip" binding:"required"`
	ISP           string `json:"isp"`
	Region        string `json:"region"`
	OriginAddr    string `json:"origin_addr"`
	OriginPort    int    `json:"origin_port"`
	AuthConfirmed bool   `json:"auth_confirmed"`
}

func (h *NodeHandler) CreateNode(c *gin.Context) {
	var req createNodeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	v, err := h.svc.CreateNode(c.Request.Context(), nodeapp.CreateNodeInput{
		Name:          req.Name,
		IP:            req.IP,
		ISP:           req.ISP,
		Region:        req.Region,
		OriginAddr:    req.OriginAddr,
		OriginPort:    req.OriginPort,
		AuthConfirmed: req.AuthConfirmed,
	})
	if err != nil {
		if errors.Is(err, nodeapp.ErrNodeIPTaken) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, v)
}

func (h *NodeHandler) ListNodes(c *gin.Context) {
	afterID, _ := parseUint64Query(c, "after_id")
	limit, _ := parseIntQuery(c, "limit", 20)
	views, err := h.svc.ListNodes(c.Request.Context(), afterID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": views})
}

func (h *NodeHandler) GetNode(c *gin.Context) {
	id, err := parseUint64Param(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	v, err := h.svc.GetNode(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, nodeapp.ErrNodeNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, v)
}

func (h *NodeHandler) DeleteNode(c *gin.Context) {
	id, err := parseUint64Param(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.DeleteNode(c.Request.Context(), id); err != nil {
		if errors.Is(err, nodeapp.ErrNodeNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *NodeHandler) RetriggerHTTPSCheck(c *gin.Context) {
	id, err := parseUint64Param(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.RetriggerHTTPSCheck(c.Request.Context(), id); err != nil {
		if errors.Is(err, nodeapp.ErrNodeNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"message": "HTTPS check triggered"})
}

func (h *NodeHandler) InstallCommand(c *gin.Context) {
	id, err := parseUint64Param(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	addr := c.Query("management_addr")
	if addr == "" {
		addr = "https://smartcdn.example.com"
	}
	cmd, err := h.svc.InstallCommand(c.Request.Context(), id, addr)
	if err != nil {
		if errors.Is(err, nodeapp.ErrNodeNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"command": cmd})
}

type dispatchTaskReq struct {
	Type    string `json:"type" binding:"required"`
	Payload string `json:"payload"`
}

func (h *NodeHandler) DispatchTask(c *gin.Context) {
	id, err := parseUint64Param(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req dispatchTaskReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	v, err := h.svc.DispatchTask(c.Request.Context(), nodeapp.DispatchTaskInput{
		NodeID:  id,
		Type:    req.Type,
		Payload: req.Payload,
	})
	if err != nil {
		if errors.Is(err, nodeapp.ErrNodeNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, v)
}

// --- Pool ---

type createPoolReq struct {
	Name    string   `json:"name" binding:"required,max=128"`
	NodeIDs []uint64 `json:"node_ids"`
}

func (h *NodeHandler) CreatePool(c *gin.Context) {
	var req createPoolReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	v, err := h.svc.CreatePool(c.Request.Context(), nodeapp.CreatePoolInput{
		Name:    req.Name,
		NodeIDs: req.NodeIDs,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, v)
}

func (h *NodeHandler) ListPools(c *gin.Context) {
	afterID, _ := parseUint64Query(c, "after_id")
	limit, _ := parseIntQuery(c, "limit", 20)
	views, err := h.svc.ListPools(c.Request.Context(), afterID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": views})
}

func (h *NodeHandler) GetPool(c *gin.Context) {
	pid, err := parseUint64Param(c, "pid")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pool id"})
		return
	}
	v, err := h.svc.GetPool(c.Request.Context(), pid)
	if err != nil {
		if errors.Is(err, nodeapp.ErrPoolNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, v)
}

func (h *NodeHandler) DeletePool(c *gin.Context) {
	pid, err := parseUint64Param(c, "pid")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pool id"})
		return
	}
	if err := h.svc.DeletePool(c.Request.Context(), pid); err != nil {
		if errors.Is(err, nodeapp.ErrPoolNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

type addMemberReq struct {
	NodeID uint64 `json:"node_id" binding:"required"`
}

func (h *NodeHandler) AddMember(c *gin.Context) {
	pid, err := parseUint64Param(c, "pid")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pool id"})
		return
	}
	var req addMemberReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.AddPoolMember(c.Request.Context(), pid, req.NodeID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *NodeHandler) RemoveMember(c *gin.Context) {
	pid, err := parseUint64Param(c, "pid")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pool id"})
		return
	}
	nid, err := parseUint64Param(c, "nid")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid node id"})
		return
	}
	if err := h.svc.RemovePoolMember(c.Request.Context(), pid, nid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// --- Helpers reused from dns_account_handler ---
// parseUint64Query defined in cdn_handler.go (same package)

// taskStatusFromString converts a string to NodeTaskStatus.
func taskStatusFromString(s string) entity.NodeTaskStatus {
	switch entity.NodeTaskStatus(s) {
	case entity.NodeTaskSucceeded, entity.NodeTaskFailed:
		return entity.NodeTaskStatus(s)
	default:
		return entity.NodeTaskFailed
	}
}
