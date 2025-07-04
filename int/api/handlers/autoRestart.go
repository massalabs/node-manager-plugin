package handlers

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/massalabs/node-manager-plugin/api/restapi/operations"
	configPkg "github.com/massalabs/node-manager-plugin/int/config"
)

func HandleSetAutoRestart() func(operations.SetAutoRestartParams) middleware.Responder {
	return func(params operations.SetAutoRestartParams) middleware.Responder {
		configPkg.GlobalPluginInfo.SetAutoRestart(params.Body.AutoRestart)
		return operations.NewSetAutoRestartNoContent()
	}
}
