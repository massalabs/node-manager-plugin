package html

import (
	"embed"
	"mime"
	"net/http"
	"path/filepath"

	"github.com/go-openapi/runtime/middleware"
	"github.com/massalabs/node-manager-plugin/api/restapi/operations"
	"github.com/massalabs/station-massa-wallet/pkg/openapi"
)

const (
	indexHTML      = "index.html"
	basePathWebApp = "dist/"
)

//nolint:typecheck
//go:embed dist
var contentWebApp embed.FS

// Handle a Web request.
func HandleWebApp(params operations.PluginWebAppParams) middleware.Responder {
	resourceName := params.Resource

	resourceContent, err := contentWebApp.ReadFile(basePathWebApp + resourceName)
	if err != nil {
		resourceName = "index.html"

		resourceContent, err = contentWebApp.ReadFile(basePathWebApp + resourceName)
		if err != nil {
			return operations.NewPluginWebAppNotFound()
		}
	}

	fileExtension := filepath.Ext(resourceName)

	mimeType := mime.TypeByExtension(fileExtension)

	header := map[string]string{"Content-Type": mimeType}

	return openapi.NewCustomResponder(resourceContent, header, http.StatusOK)
}

// DefaultRedirectHandler redirects request to "/" URL to "web-app/index"
func DefaultRedirectHandler(_ operations.DefaultPageParams) middleware.Responder {
	return openapi.NewCustomResponder(nil, map[string]string{"Location": "web/index"}, http.StatusPermanentRedirect)
}

// AppendEndpoints appends web endpoints to the API.
func AppendEndpoints(api *operations.NodeManagerPluginAPI) {
	api.DefaultPageHandler = operations.DefaultPageHandlerFunc(DefaultRedirectHandler)
	api.PluginWebAppHandler = operations.PluginWebAppHandlerFunc(HandleWebApp)
}
