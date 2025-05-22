package handlers

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/massalabs/node-manager-plugin/api/models"
	"github.com/massalabs/node-manager-plugin/api/restapi/operations"
	NodeManagerPkg "github.com/massalabs/node-manager-plugin/int/node-manager"
	"github.com/massalabs/station/pkg/logger"
)

func HandleStopNode(nodeManager *NodeManagerPkg.INodeManager) func(operations.StopNodeParams) middleware.Responder {
	return func(params operations.StopNodeParams) middleware.Responder {
		// Check if the node is already stopped
		status, _ := (*nodeManager).GetStatus()
		if !NodeManagerPkg.IsRunning(status) {
			return createErrorResponse(400, "Node is already stopped")
		}

		logger.Infof("Current node status is %s", status)

		if err := (*nodeManager).StopNode(); err != nil {
			return operations.NewStopNodeInternalServerError().WithPayload(&models.Error{
				Message: err.Error(),
			})
		}
		return operations.NewStopNodeNoContent()
	}
}
