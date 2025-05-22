package nodeManager

type INodeManager interface {
	StartNode(isMainnet bool, pwd string) (string, error)
	StopNode() error
	Logs() (string, error)
	GetStatus() (NodeStatus, chan NodeStatus)
}
