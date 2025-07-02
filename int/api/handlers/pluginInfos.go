package handlers

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/massalabs/node-manager-plugin/api/models"
	"github.com/massalabs/node-manager-plugin/api/restapi/operations"
	nodeDirManagerPkg "github.com/massalabs/node-manager-plugin/int/node-bin-dir-manager"
	nodeManagerPkg "github.com/massalabs/node-manager-plugin/int/node-manager"
	"github.com/massalabs/station/pkg/logger"
)

func HandleGetPluginInfos(nodeManager *nodeManagerPkg.INodeManager, nodeDirManager *nodeDirManagerPkg.NodeDirManager) func(operations.GetPluginInfosParams) middleware.Responder {
	return func(params operations.GetPluginInfosParams) middleware.Responder {
		nodeInfos := (*nodeManager).GetNodeInfos()
		version, err := (*nodeDirManager).GetVersion(nodeInfos.IsMainnet)
		if err != nil {
			logger.Errorf("failed to get node version: %v", err)
			return operations.NewGetPluginInfosInternalServerError().WithPayload(&models.Error{
				Message: err.Error(),
			})
		}
		return operations.NewGetPluginInfosOK().WithPayload(&models.PluginInfos{
			Version:     version,
			AutoRestart: nodeInfos.AutoRestart,
		})
	}
}
