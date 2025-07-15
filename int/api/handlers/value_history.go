package handlers

import (
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/massalabs/node-manager-plugin/api/models"
	"github.com/massalabs/node-manager-plugin/api/restapi/operations"
	"github.com/massalabs/node-manager-plugin/int/config"
	"github.com/massalabs/node-manager-plugin/int/db"
	historymanager "github.com/massalabs/node-manager-plugin/int/history-manager"
)

// ValueHistorySample is the response item
// type ValueHistorySample struct {
// 	Timestamp time.Time  `json:"timestamp"`
// 	Value     *float64   `json:"value"`
// }

type valueHistoryRequest struct {
	Since      string `json:"since"`
	SampleNum  int    `json:"sampleNum"`
	IsBuildnet bool   `json:"isBuildnet,omitempty"`
}

func HandleGetValueHistory(database db.DB, historyMgr *historymanager.HistoryManager, config *config.PluginConfig) func(params operations.GetValueHistoryParams) middleware.Responder {
	return func(params operations.GetValueHistoryParams) middleware.Responder {

		if params.Body.SampleNum < 1 {
			return createErrorResponse(400, "SampleNum must be greater than 0")
		}

		since, err := time.Parse(time.RFC3339, params.Body.Since)
		if err != nil {
			return createErrorResponse(400, "Invalid date format. Expected RFC3339 format")
		}

		now := time.Now()
		interval := now.Sub(since) / time.Duration(params.Body.SampleNum)
		if interval < time.Duration(config.TotValueRegisterInterval) {
			return createErrorResponse(400, "SampleNum is too large")
		}

		result, err := historyMgr.SampleValueHistory(since, int64(params.Body.SampleNum), params.Body.IsMainnet, interval)
		if err != nil {
			return createErrorResponse(500, err.Error())
		}

		samples := make(models.ValueHistorySamples, len(result))
		for i, r := range result {
			item := &models.ValueHistorySamplesItems0{
				Timestamp: strfmt.DateTime(r.Timestamp),
			}
			if r.Value != nil {
				item.Value = *r.Value
			}
			samples[i] = item
		}
		return operations.NewGetValueHistoryOK().WithPayload(samples)
	}
}
