package handlers

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/massalabs/node-manager-plugin/api/models"
	"github.com/massalabs/node-manager-plugin/api/restapi/operations"
	"github.com/massalabs/node-manager-plugin/int/config"
)

func HandleGetPluginInfos() func(operations.GetPluginInfosParams) middleware.Responder {
	return func(params operations.GetPluginInfosParams) middleware.Responder {
		isMainnet := config.GlobalPluginInfo.GetIsMainnet()
		return operations.NewGetPluginInfosOK().WithPayload(&models.PluginInfos{
			Networks: []*models.Network{
				{
					Version: config.GlobalPluginInfo.MainnetVersion,
					HasPwd:  config.GlobalPluginInfo.GetPwdByNetwork(true) != "",
				},
				{
					Version: config.GlobalPluginInfo.BuildnetVersion,
					HasPwd:  config.GlobalPluginInfo.GetPwdByNetwork(false) != "",
				},
			},
			AutoRestart:   config.GlobalPluginInfo.GetAutoRestart(),
			PluginVersion: config.Version,
			IsMainnet:     isMainnet,
		})
	}
}
