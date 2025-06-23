package handlers

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/massalabs/node-manager-plugin/api/restapi/operations"
	nodeManagerPkg "github.com/massalabs/node-manager-plugin/int/node-manager"
)

func HandleSetAutoRestart(nodeManager *nodeManagerPkg.INodeManager) func(operations.SetAutoRestartParams) middleware.Responder {
	return func(params operations.SetAutoRestartParams) middleware.Responder {
		(*nodeManager).SetAutoRestart(params.Body.AutoRestart)
		return operations.NewSetAutoRestartNoContent()
	}
}
