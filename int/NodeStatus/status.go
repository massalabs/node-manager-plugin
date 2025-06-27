package nodeStatus

// NodeStatus represents the current NodeStatus of the node
type NodeStatus string

// node Status constants
const (
	NodeStatusOn            NodeStatus = "on"
	NodeStatusOff           NodeStatus = "off"
	NodeStatusStarting      NodeStatus = "starting"
	NodeStatusBootstrapping NodeStatus = "bootstrapping"
	NodeStatusStopping      NodeStatus = "stopping"
	NodeStatusCrashed       NodeStatus = "crashed"
	NodeStatusDesynced      NodeStatus = "desynced"
)
