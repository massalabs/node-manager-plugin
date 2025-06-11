package handlers

import (
	"fmt"
	"regexp"

	"github.com/go-openapi/runtime/middleware"
	"github.com/massalabs/node-manager-plugin/api/models"
	"github.com/massalabs/node-manager-plugin/api/restapi/operations"
	nodeManagerPkg "github.com/massalabs/node-manager-plugin/int/node-manager"
)

func HandleGetNodeLogs(nodeManager *nodeManagerPkg.INodeManager) func(operations.GetNodeLogsParams) middleware.Responder {
	return func(params operations.GetNodeLogsParams) middleware.Responder {
		logs, err := (*nodeManager).Logs()
		if err != nil {
			return operations.NewGetNodeLogsInternalServerError().WithPayload(&models.Error{
				Message: err.Error(),
			})
		}

		// Remove all ANSI escape sequences
		cleanLogs := removeAnsiCodes(logs)
		fmt.Println(cleanLogs)

		return operations.NewGetNodeLogsOK().WithPayload(cleanLogs)
	}
}

// removeAnsiCodes removes all ANSI escape sequences from the string
func removeAnsiCodes(str string) string {
	// Regular expression to match ANSI escape sequences
	// This matches ESC[ followed by any number of parameters and ending with a letter
	re := regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
	return re.ReplaceAllString(str, "")
}
