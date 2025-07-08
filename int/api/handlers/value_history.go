package handlers

import (
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/massalabs/node-manager-plugin/api/models"
	"github.com/massalabs/node-manager-plugin/api/restapi/operations"
	"github.com/massalabs/node-manager-plugin/int/config"
	"github.com/massalabs/node-manager-plugin/int/db"
	"github.com/massalabs/node-manager-plugin/int/utils"
)

func HandleGetValueHistory(database db.DB) func(operations.GetValueHistoryParams) middleware.Responder {
	return func(params operations.GetValueHistoryParams) middleware.Responder {
		// Parse the since parameter
		since, err := time.Parse(time.RFC3339, params.Since)
		if err != nil {
			return operations.NewGetValueHistoryInternalServerError().WithPayload(&models.Error{
				Message: "Invalid date format. Expected RFC3339 format",
			})
		}

		// Determine the current network
		currentNetwork := utils.NetworkMainnet
		if !config.GlobalPluginInfo.GetIsMainnet() {
			currentNetwork = utils.NetworkBuildnet
		}

		// Get history from database
		histories, err := database.GetHistory(since, currentNetwork)
		if err != nil {
			return operations.NewGetValueHistoryInternalServerError().WithPayload(&models.Error{
				Message: err.Error(),
			})
		}

		// Convert to API models
		var apiHistories []*models.ValueHistoryItems0
		for _, history := range histories {
			apiHistories = append(apiHistories, &models.ValueHistoryItems0{
				Timestamp:  strfmt.DateTime(history.Timestamp),
				TotalValue: history.TotalValue,
			})
		}

		return operations.NewGetValueHistoryOK().WithPayload(apiHistories)
	}
}
