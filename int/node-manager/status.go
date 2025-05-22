package nodeManager

import "sync"

// NodeStatus represents the current NodeStatus of the node
type NodeStatus string

// node Status constants
const (
	NodeStatusOn            NodeStatus = "on"
	NodeStatusOff           NodeStatus = "off"
	NodeStatusBootstrapping NodeStatus = "bootstrapping"
	NodeStatusStopping      NodeStatus = "stopping"
	NodeStatusError         NodeStatus = "error"
	NodeManagerErrorStatus  NodeStatus = "nodeManagerError"
)

type NodeStatusHandler struct {
	status chan NodeStatus
	hasNew bool
	mu     sync.Mutex
}

func NewNodeStatusHandler() *NodeStatusHandler {
	return &NodeStatusHandler{
		status: make(chan NodeStatus, 1), // Buffered channel of size 1
	}
}

func (n *NodeStatusHandler) SetStatus(status NodeStatus) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if len(n.status) > 0 {
		<-n.status
	}
	n.status <- status
	n.hasNew = true
}
