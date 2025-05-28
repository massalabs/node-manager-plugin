package handlers

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/massalabs/node-manager-plugin/api/models"
	"github.com/massalabs/node-manager-plugin/api/restapi/operations"
	nodeManagerPkg "github.com/massalabs/node-manager-plugin/int/node-manager"
)

func HandleGetNodeLogs(nodeManager *nodeManagerPkg.INodeManager) func(operations.GetNodeLogsParams) middleware.Responder {
	return func(params operations.GetNodeLogsParams) middleware.Responder {
		logs, err := (*nodeManager).Logs()
		if err != nil {
			return operations.NewGetNodeLogsInternalServerError().WithPayload(&models.Error{
				Message: err.Error(),
			})
		}
		return operations.NewGetNodeLogsOK().WithPayload(logs)
	}
}
