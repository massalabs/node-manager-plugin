// This file is safe to edit. Once it exists it will not be overwritten

package restapi

import (
	"crypto/tls"
	"io"
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/massalabs/node-manager-plugin/api/restapi/operations"
)

//go:generate swagger generate server --target ../../api --name NodeManagerPlugin --spec ../pluginAPI-V0.yml --principal interface{} --exclude-main

func configureFlags(api *operations.NodeManagerPluginAPI) {
	// api.CommandLineOptionsGroups = []swag.CommandLineOptionsGroup{ ... }
}

func configureAPI(api *operations.NodeManagerPluginAPI) http.Handler {
	// configure the api here
	api.ServeError = errors.ServeError

	// Set your custom logger if needed. Default one is log.Printf
	// Expected interface func(string, ...interface{})
	//
	// Example:
	// api.Logger = log.Printf

	api.UseSwaggerUI()
	// To continue using redoc as your UI, uncomment the following line
	// api.UseRedoc()

	api.JSONConsumer = runtime.JSONConsumer()

	api.BinProducer = runtime.ByteStreamProducer()
	api.CSSProducer = runtime.ProducerFunc(func(w io.Writer, data interface{}) error {
		return errors.NotImplemented("css producer has not yet been implemented")
	})
	api.HTMLProducer = runtime.ProducerFunc(func(w io.Writer, data interface{}) error {
		return errors.NotImplemented("html producer has not yet been implemented")
	})
	api.JsProducer = runtime.ProducerFunc(func(w io.Writer, data interface{}) error {
		return errors.NotImplemented("js producer has not yet been implemented")
	})
	api.JSONProducer = runtime.JSONProducer()
	api.TextWebpProducer = runtime.ProducerFunc(func(w io.Writer, data interface{}) error {
		return errors.NotImplemented("textWebp producer has not yet been implemented")
	})

	if api.DefaultPageHandler == nil {
		api.DefaultPageHandler = operations.DefaultPageHandlerFunc(func(params operations.DefaultPageParams) middleware.Responder {
			return middleware.NotImplemented("operation operations.DefaultPage has not yet been implemented")
		})
	}
	if api.GetMassaNodeStatusHandler == nil {
		api.GetMassaNodeStatusHandler = operations.GetMassaNodeStatusHandlerFunc(func(params operations.GetMassaNodeStatusParams) middleware.Responder {
			return middleware.NotImplemented("operation operations.GetMassaNodeStatus has not yet been implemented")
		})
	}
	if api.PluginWebAppHandler == nil {
		api.PluginWebAppHandler = operations.PluginWebAppHandlerFunc(func(params operations.PluginWebAppParams) middleware.Responder {
			return middleware.NotImplemented("operation operations.PluginWebApp has not yet been implemented")
		})
	}
	if api.StartNodeHandler == nil {
		api.StartNodeHandler = operations.StartNodeHandlerFunc(func(params operations.StartNodeParams) middleware.Responder {
			return middleware.NotImplemented("operation operations.StartNode has not yet been implemented")
		})
	}
	if api.StopNodeHandler == nil {
		api.StopNodeHandler = operations.StopNodeHandlerFunc(func(params operations.StopNodeParams) middleware.Responder {
			return middleware.NotImplemented("operation operations.StopNode has not yet been implemented")
		})
	}

	api.PreServerShutdown = func() {}

	api.ServerShutdown = func() {}

	return setupGlobalMiddleware(api.Serve(setupMiddlewares))
}

// The TLS configuration before HTTPS server starts.
func configureTLS(tlsConfig *tls.Config) {
	// Make all necessary changes to the TLS configuration here.
}

// As soon as server is initialized but not run yet, this function will be called.
// If you need to modify a config, store server instance to stop it individually later, this is the place.
// This function can be called multiple times, depending on the number of serving schemes.
// scheme value will be set accordingly: "http", "https" or "unix".
func configureServer(s *http.Server, scheme, addr string) {
}

// The middleware configuration is for the handler executors. These do not apply to the swagger.json document.
// The middleware executes after routing but before authentication, binding and validation.
func setupMiddlewares(handler http.Handler) http.Handler {
	return handler
}

// The middleware configuration happens before anything, this middleware also applies to serving the swagger.json document.
// So this is a good place to plug in a panic handling middleware, logging and metrics.
func setupGlobalMiddleware(handler http.Handler) http.Handler {
	return handler
}
