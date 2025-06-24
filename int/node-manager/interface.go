package nodeManager

type INodeManager interface {
	StartNode(isMainnet bool, pwd string) (string, error)
	StopNode() error
	Logs(isMainnet bool) (string, error)
	GetStatus() (NodeStatus, <-chan NodeStatus)
	GetNodeInfos() NodeInfos
	SetAutoRestart(autoRestart bool)
	Close() error
}
