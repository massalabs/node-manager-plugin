package handlers

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/massalabs/node-manager-plugin/api/models"
	"github.com/massalabs/node-manager-plugin/api/restapi/operations"
	nodeStatusPkg "github.com/massalabs/node-manager-plugin/int/NodeStatus"
	nodeManagerPkg "github.com/massalabs/node-manager-plugin/int/node-manager"
)

func HandleStartNode(nodeManager nodeManagerPkg.INodeManager, statusDispatcher nodeStatusPkg.NodeStatusDispatcher) func(operations.StartNodeParams) middleware.Responder {
	return func(params operations.StartNodeParams) middleware.Responder {
		// Check if the node is already running
		if nodeManagerPkg.IsRunning(nodeManager.GetStatus()) {
			return createErrorResponse(400, "Node is already running")
		}

		version, err := nodeManager.StartNode(!params.Body.UseBuildnet, params.Body.Password)
		if err != nil {
			return operations.NewStartNodeInternalServerError().WithPayload(&models.Error{
				Message: err.Error(),
			})
		}
		return operations.NewStartNodeOK().WithPayload(&models.StartNodeResponse{Version: version})
	}
}
