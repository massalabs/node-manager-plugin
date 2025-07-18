package handlers

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/massalabs/node-manager-plugin/api/models"
	"github.com/massalabs/node-manager-plugin/api/restapi/operations"
	nodeStatusPkg "github.com/massalabs/node-manager-plugin/int/NodeStatus"
	clientDriverPkg "github.com/massalabs/node-manager-plugin/int/client-driver"
	"github.com/massalabs/node-manager-plugin/int/config"
	nodeDirManagerPkg "github.com/massalabs/node-manager-plugin/int/node-bin-dir-manager"
	nodeManagerPkg "github.com/massalabs/node-manager-plugin/int/node-manager"
)

func HandleStartNode(nodeManager nodeManagerPkg.INodeManager, statusDispatcher nodeStatusPkg.NodeStatusDispatcher, nodeDirManager *nodeDirManagerPkg.NodeDirManager, pluginConfig *config.PluginConfig) func(operations.StartNodeParams) middleware.Responder {
	return func(params operations.StartNodeParams) middleware.Responder {
		// Check if the node is already running
		if nodeManagerPkg.IsRunning(nodeManager.GetStatus()) {
			return createErrorResponse(400, "Node is already running")
		}

		pwd := params.Body.Password

		// Trim any whitespace from the password
		pwd = strings.TrimSpace(pwd)

		if pwd == "" {
			registeredPwd := config.GlobalPluginInfo.GetPwdByNetwork(!params.Body.UseBuildnet)
			if registeredPwd == "" {
				return createErrorResponse(400, "Password is required")
			}
			pwd = registeredPwd
		} else {
			// Validate the password
			if err := checkPwd(pwd, pluginConfig, nodeDirManager, !params.Body.UseBuildnet); err != nil {
				return operations.NewStartNodeInternalServerError().WithPayload(&models.Error{
					Message: err.Error(),
				})
			}
		}

		err := nodeManager.StartNode(!params.Body.UseBuildnet, pwd)
		if err != nil {
			return operations.NewStartNodeInternalServerError().WithPayload(&models.Error{
				Message: err.Error(),
			})
		}
		return operations.NewStartNodeOK()
	}
}

// checkAndUpdatePwd validates the provided password against stored hash or tests it with WalletInfo
func checkPwd(
	pwd string,
	pluginConfig *config.PluginConfig,
	nodeDirManager *nodeDirManagerPkg.NodeDirManager,
	isMainnet bool,
) error {
	clientDriver, err := clientDriverPkg.NewClientDriver(
		isMainnet,
		*nodeDirManager,
		time.Duration(pluginConfig.ClientTimeout)*time.Second, // 30 second timeout
	)
	if err != nil {
		return fmt.Errorf("failed to create client driver: %v", err)
	}

	// Test the password by calling WalletInfo
	_, err = clientDriver.WalletInfoWithoutNode(pwd)
	if err != nil {
		return fmt.Errorf("password validation failed: %w", err)
	}

	return nil
}
