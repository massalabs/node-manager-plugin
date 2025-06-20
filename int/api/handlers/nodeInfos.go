package handlers

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/massalabs/node-manager-plugin/api/models"
	"github.com/massalabs/node-manager-plugin/api/restapi/operations"
	nodeManagerPkg "github.com/massalabs/node-manager-plugin/int/node-manager"
)

func HandleSetAutoRestart(nodeManager *nodeManagerPkg.INodeManager) func(operations.SetAutoRestartParams) middleware.Responder {
	return func(params operations.SetAutoRestartParams) middleware.Responder {
		(*nodeManager).SetAutoRestart(params.Body.AutoRestart)
		return operations.NewSetAutoRestartNoContent()
	}
}

func HandleGetNodeInfos(nodeManager *nodeManagerPkg.INodeManager) func(operations.GetNodeInfosParams) middleware.Responder {
	return func(params operations.GetNodeInfosParams) middleware.Responder {
		nodeInfos := (*nodeManager).GetNodeInfos()
		return operations.NewGetNodeInfosOK().WithPayload(&models.NodeInfos{
			AutoRestart: nodeInfos.AutoRestart,
			Version:     nodeInfos.Version,
		})
	}
}
