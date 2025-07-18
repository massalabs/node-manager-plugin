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
		mainnetVersion, err := (*nodeDirManager).GetVersion(true)
		if err != nil {
			logger.Errorf("failed to retrieve node version for mainnet: %v", err)
			return operations.NewGetPluginInfosInternalServerError().WithPayload(&models.Error{
				Message: err.Error(),
			})
		}

		buildnetVersion, err := (*nodeDirManager).GetVersion(false)
		if err != nil {
			logger.Errorf("failed to retrieve node version for buildnet: %v", err)
			return operations.NewGetPluginInfosInternalServerError().WithPayload(&models.Error{
				Message: err.Error(),
			})
		}
		return operations.NewGetPluginInfosOK().WithPayload(&models.PluginInfos{
			Networks: []*models.Network{
				{
					Version: mainnetVersion,
					HasPwd:  config.GlobalPluginInfo.GetPwdByNetwork(true) != "",
				},
				{
					Version: buildnetVersion,
					HasPwd:  config.GlobalPluginInfo.GetPwdByNetwork(false) != "",
				},
			},
			AutoRestart:   config.GlobalPluginInfo.GetAutoRestart(),
			PluginVersion: config.Version,
			IsMainnet:     isMainnet,
		})
	}
}
