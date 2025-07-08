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
			if err := checkPwd(pwd, pluginConfig, nodeDirManager); err != nil {
				return operations.NewStartNodeInternalServerError().WithPayload(&models.Error{
					Message: err.Error(),
				})
			}
		}

		// // If the password is valid and the password hash is not set, update the password hash in the config
		// if pluginConfig.PwdHash == "" {
		// 	if err := updatePwdHash(pwd, pluginConfig); err != nil {
		// 		return operations.NewStartNodeInternalServerError().WithPayload(&models.Error{
		// 			Message: fmt.Errorf("Password is valid but failed to update password hash in the config: %v", err).Error(),
		// 		})
		// 	}
		// }

		version, err := nodeManager.StartNode(!params.Body.UseBuildnet, pwd)
		if err != nil {
			return operations.NewStartNodeInternalServerError().WithPayload(&models.Error{
				Message: err.Error(),
			})
		}
		return operations.NewStartNodeOK().WithPayload(&models.StartNodeResponse{Version: version})
	}
}

// checkAndUpdatePwd validates the provided password against stored hash or tests it with WalletInfo
func checkPwd(pwd string, pluginConfig *config.PluginConfig, nodeDirManager *nodeDirManagerPkg.NodeDirManager) error {
	// // If pwdHash exists, check that the provided password matches the hash
	// if pluginConfig.PwdHash != "" {
	// 	hash := sha256.Sum256([]byte(pwd))
	// 	providedHash := hex.EncodeToString(hash[:])

	// 	if providedHash != pluginConfig.PwdHash {
	// 		return fmt.Errorf("password does not match stored hash")
	// 	}

	// 	logger.Infof("Password validated against stored hash")
	// 	return nil
	// }

	// // If no pwdHash exists, test the password with WalletInfo
	// logger.Infof("No stored password hash found, testing password with WalletInfo")

	// Create client driver to test the password
	clientDriver, err := clientDriverPkg.NewClientDriver(
		config.GlobalPluginInfo.GetIsMainnet(),
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

// func updatePwdHash(pwd string, pluginConfig *config.PluginConfig) error {
// 	hash := sha256.Sum256([]byte(pwd))
// 	pwdHash := hex.EncodeToString(hash[:])
// 	err := config.UpdateConfigField("PwdHash", pwdHash)
// 	if err != nil {
// 		return fmt.Errorf("Failed to save password hash to config: %v", err)
// 	}
// 	logger.Infof("Password hash saved to config")

// 	// Update the config pointer of the plugin with the new password hash
// 	pluginConfig.PwdHash = pwdHash
// 	return nil
// }
