package handlers

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/massalabs/node-manager-plugin/api/restapi/operations"
	nodeManagerPkg "github.com/massalabs/node-manager-plugin/int/node-manager"
)

func HandleSetConfig(nodeManager *nodeManagerPkg.INodeManager) func(operations.SetConfigParams) middleware.Responder {
	return func(params operations.SetConfigParams) middleware.Responder {
		(*nodeManager).SetAutoRestart(params.Body.AutoRestart)
		return operations.NewSetConfigNoContent()
	}
}

func HandleGetConfig(nodeManager *nodeManagerPkg.INodeManager) func(operations.GetConfigParams) middleware.Responder {
	return func(params operations.GetConfigParams) middleware.Responder {
		conf := (*nodeManager).GetNodeManagerConfig()
		return operations.NewGetConfigOK().WithPayload(&conf)
	}
}
