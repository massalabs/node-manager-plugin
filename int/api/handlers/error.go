package handlers

import (
	"net/http"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/massalabs/node-manager-plugin/api/models"
	"github.com/massalabs/station/pkg/logger"
)

// Create an error response with the given status code and message
func createErrorResponse(statusCode int, message string) middleware.Responder {
	return &customErrorResponder{
		statusCode: statusCode,
		payload: &models.Error{
			Message: message,
		},
	}
}

// customErrorResponder implements the Responder interface for error responses
type customErrorResponder struct {
	statusCode int
	payload    *models.Error
}

// WriteResponse writes the error response
func (r *customErrorResponder) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {
	rw.WriteHeader(r.statusCode)
	if err := producer.Produce(rw, r.payload); err != nil {
		logger.Errorf("Failed to produce error response: %v", err)
	}
}
