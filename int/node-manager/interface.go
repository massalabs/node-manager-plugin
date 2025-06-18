package nodeManager

import "github.com/massalabs/node-manager-plugin/api/models"

type INodeManager interface {
	StartNode(isMainnet bool, pwd string) (string, error)
	StopNode() error
	Logs(isMainnet bool) (string, error)
	GetStatus() (NodeStatus, <-chan NodeStatus)
	GetNodeManagerConfig() models.Config
	SetAutoRestart(autoRestart bool)
	Close() error
}
