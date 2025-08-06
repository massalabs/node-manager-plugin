package nodeAPI

import (
	"context"
	"encoding/json"

	"github.com/massalabs/station/pkg/node"
)

const (
	NodeURL = "http://localhost:33035"
)

type NodeAPI interface {
	GetAddresses(addresses []string) ([]byte, error)
	GetOperation(operationID string) (*node.Operation, error)
	GetStatus() (*node.State, error)
}

type nodeAPI struct {
	nodeClient *node.Client
}

func NewNodeAPI() NodeAPI {
	nodeClient := node.NewClient(NodeURL)
	return &nodeAPI{
		nodeClient: nodeClient,
	}
}

func (n *nodeAPI) GetAddresses(addresses []string) ([]byte, error) {
	RPCresponse, err := n.nodeClient.RPCClient.Call(
		context.Background(),
		"get_addresses",
		[1][]string{addresses})
	if err != nil {
		return nil, err
	}

	if RPCresponse.Error != nil {
		return nil, RPCresponse.Error
	}

	js, err := json.Marshal(RPCresponse.Result)
	if err != nil {
		return nil, err
	}

	return js, nil
}

func (n *nodeAPI) GetOperation(operationID string) (*node.Operation, error) {
	operations, err := node.Operations(n.nodeClient, []string{operationID})
	if err != nil {
		return nil, err
	}

	return &operations[0], nil
}

func (n *nodeAPI) GetStatus() (*node.State, error) {
	return node.Status(n.nodeClient)
}
