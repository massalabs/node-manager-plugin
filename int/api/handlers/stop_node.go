package handlers

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/massalabs/node-manager-plugin/api/models"
	"github.com/massalabs/node-manager-plugin/api/restapi/operations"
	nodeStatusPkg "github.com/massalabs/node-manager-plugin/int/NodeStatus"
	NodeManagerPkg "github.com/massalabs/node-manager-plugin/int/node-manager"
	"github.com/massalabs/station/pkg/logger"
)

func HandleStopNode(nodeManager NodeManagerPkg.INodeManager, statusDispatcher nodeStatusPkg.NodeStatusDispatcher) func(operations.StopNodeParams) middleware.Responder {
	return func(params operations.StopNodeParams) middleware.Responder {
		// Check if the node is already stopped
		if !NodeManagerPkg.IsRunning(nodeManager.GetStatus()) {
			return createErrorResponse(400, "Node is already stopped")
		}

		logger.Infof("Current node status is %s", statusDispatcher.GetCurrentStatus())

		if err := nodeManager.StopNode(); err != nil {
			return operations.NewStopNodeInternalServerError().WithPayload(&models.Error{
				Message: err.Error(),
			})
		}
		return operations.NewStopNodeNoContent()
	}
}
