package handlers

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/massalabs/node-manager-plugin/api/models"
	"github.com/massalabs/node-manager-plugin/api/restapi/operations"
	nodeManagerPkg "github.com/massalabs/node-manager-plugin/int/node-manager"
)

func HandleGetPluginInfos(nodeManager *nodeManagerPkg.INodeManager) func(operations.GetPluginInfosParams) middleware.Responder {
	return func(params operations.GetPluginInfosParams) middleware.Responder {
		nodeInfos := (*nodeManager).GetNodeInfos()
		return operations.NewGetPluginInfosOK().WithPayload(&models.PluginInfos{
			Version:     nodeInfos.Version,
			AutoRestart: nodeInfos.AutoRestart,
		})
	}
}
