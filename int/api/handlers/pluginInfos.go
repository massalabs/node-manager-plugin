package handlers

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/massalabs/node-manager-plugin/api/models"
	"github.com/massalabs/node-manager-plugin/api/restapi/operations"
	"github.com/massalabs/node-manager-plugin/int/config"
	nodeDirManagerPkg "github.com/massalabs/node-manager-plugin/int/node-bin-dir-manager"
	"github.com/massalabs/station/pkg/logger"
)

func HandleGetPluginInfos(nodeDirManager *nodeDirManagerPkg.NodeDirManager) func(operations.GetPluginInfosParams) middleware.Responder {
	return func(params operations.GetPluginInfosParams) middleware.Responder {
		isMainnet := config.GlobalPluginInfo.GetIsMainnet()
		version, err := (*nodeDirManager).GetVersion(isMainnet)
		if err != nil {
			logger.Errorf("failed to get node version: %v", err)
			return operations.NewGetPluginInfosInternalServerError().WithPayload(&models.Error{
				Message: err.Error(),
			})
		}
		return operations.NewGetPluginInfosOK().WithPayload(&models.PluginInfos{
			Version:        version,
			AutoRestart:    config.GlobalPluginInfo.GetAutoRestart(),
			PluginVersion:  config.Version,
			HasPwdMainnet:  config.GlobalPluginInfo.GetPwdByNetwork(true) != "",
			HasPwdBuildnet: config.GlobalPluginInfo.GetPwdByNetwork(false) != "",
		})
	}
}
