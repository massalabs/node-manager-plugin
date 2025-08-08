package handlers

import (
	"fmt"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/massalabs/node-manager-plugin/api/models"
	"github.com/massalabs/node-manager-plugin/api/restapi/operations"
	"github.com/massalabs/node-manager-plugin/int/config"
	historymanager "github.com/massalabs/node-manager-plugin/int/core/history-manager"
	dbPkg "github.com/massalabs/node-manager-plugin/int/db"
	"github.com/massalabs/node-manager-plugin/int/utils"
)

// convertUTCToLocal converts a UTC timestamp to local timezone
// SQLite stores timestamps in UTC, so we need to convert them to local timezone for display
func convertUTCToLocal(utcTime time.Time) time.Time {
	// Get the local timezone
	loc, _ := time.LoadLocation("Local")
	// Convert UTC time to local timezone
	return utcTime.In(loc)
}

func HandleGetValueHistory(db dbPkg.DB, historyMgr *historymanager.HistoryManager, config *config.PluginConfig) func(params operations.GetValueHistoryParams) middleware.Responder {
	return func(params operations.GetValueHistoryParams) middleware.Responder {
		if params.SampleNum < 1 {
			return createErrorResponse(400, "SampleNum must be greater than 0")
		}

		if params.Since == "" {
			return createErrorResponse(400, "Missing required parameter: since")
		}

		since, err := time.Parse(time.RFC3339, params.Since)
		if err != nil {
			return createErrorResponse(400, "Invalid date format. Expected RFC3339 format")
		}

		if since.After(time.Now()) {
			return createErrorResponse(400, "Since cannot be in the future")
		}

		if since.After(time.Now().Add(-time.Second * time.Duration(config.TotValueRegisterInterval))) {
			return createErrorResponse(400, fmt.Sprintf("Since param is too short, must be at least %d seconds ago", config.TotValueRegisterInterval))
		}

		now := time.Now()
		interval := now.Sub(since) / time.Duration(params.SampleNum)
		if interval < time.Duration(config.TotValueRegisterInterval) {
			return createErrorResponse(400, "SampleNum is too large")
		}

		result, err := historyMgr.SampleValueHistory(since, int64(params.SampleNum), params.IsMainnet, interval)
		if err != nil {
			return createErrorResponse(500, err.Error())
		}

		samples := make([]*models.ValueHistorySamplesResponseSamplesItems0, len(result))
		emptyDataPointNum := int64(0)
		for i, r := range result {
			// Convert UTC timestamp to local timezone for frontend display
			localTimestamp := convertUTCToLocal(r.Timestamp)
			samples[i] = &models.ValueHistorySamplesResponseSamplesItems0{
				Timestamp: strfmt.DateTime(localTimestamp),
			}
			if r.Value != nil {
				samples[i].Value = *r.Value
			} else {
				emptyDataPointNum++
			}
		}
		response := &models.ValueHistorySamplesResponse{
			Samples:           samples,
			EmptyDataPointNum: &emptyDataPointNum,
		}
		return operations.NewGetValueHistoryOK().WithPayload(response)
	}
}

func HandleGetRollOpHistory(db dbPkg.DB) func(operations.GetRollOpHistoryParams) middleware.Responder {
	return func(params operations.GetRollOpHistoryParams) middleware.Responder {
		network := utils.NetworkBuildnet
		if params.IsMainnet {
			network = utils.NetworkMainnet
		}

		histories, err := db.GetRollOpHistory(params.Address, network)
		if err != nil {
			return operations.NewGetRollOpHistoryInternalServerError().WithPayload(&models.Error{
				Message: err.Error(),
			})
		}

		rollOpHistory := make([]*models.RollOpHistory, len(histories))
		for i, history := range histories {
			// Convert UTC timestamp to local timezone for frontend display
			localTimestamp := convertUTCToLocal(history.Timestamp)
			timestamp := strfmt.DateTime(localTimestamp)
			rollOpHistory[i] = &models.RollOpHistory{
				OpID:      &history.OpId,
				Op:        &history.Op,
				Amount:    &history.Amount,
				Timestamp: &timestamp,
			}
		}

		return operations.NewGetRollOpHistoryOK().WithPayload(&models.RollOpHistoryResponse{
			Operations: rollOpHistory,
		})
	}
}
