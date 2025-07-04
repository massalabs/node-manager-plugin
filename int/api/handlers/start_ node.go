package handlers

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/massalabs/node-manager-plugin/api/models"
	"github.com/massalabs/node-manager-plugin/api/restapi/operations"
	nodeStatusPkg "github.com/massalabs/node-manager-plugin/int/NodeStatus"
	"github.com/massalabs/node-manager-plugin/int/config"
	nodeManagerPkg "github.com/massalabs/node-manager-plugin/int/node-manager"
)

func HandleStartNode(nodeManager nodeManagerPkg.INodeManager, statusDispatcher nodeStatusPkg.NodeStatusDispatcher) func(operations.StartNodeParams) middleware.Responder {
	return func(params operations.StartNodeParams) middleware.Responder {
		// Check if the node is already running
		if nodeManagerPkg.IsRunning(nodeManager.GetStatus()) {
			return createErrorResponse(400, "Node is already running")
		}

		pwd := params.Body.Password

		if pwd == "" {
			registeredPwd := config.GlobalPluginInfo.GetPwd()
			if registeredPwd == "" {
				return createErrorResponse(400, "Password is required")
			}
			pwd = registeredPwd
		}

		version, err := nodeManager.StartNode(!params.Body.UseBuildnet, pwd)
		if err != nil {
			return operations.NewStartNodeInternalServerError().WithPayload(&models.Error{
				Message: err.Error(),
			})
		}
		return operations.NewStartNodeOK().WithPayload(&models.StartNodeResponse{Version: version})
	}
}
