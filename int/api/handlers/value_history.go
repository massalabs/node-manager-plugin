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
	"github.com/massalabs/node-manager-plugin/int/db"
)

func HandleGetValueHistory(database db.DB, historyMgr *historymanager.HistoryManager, config *config.PluginConfig) func(params operations.GetValueHistoryParams) middleware.Responder {
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
			samples[i] = &models.ValueHistorySamplesResponseSamplesItems0{
				Timestamp: strfmt.DateTime(r.Timestamp),
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
