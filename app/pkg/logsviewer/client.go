package logsviewer

import (
	"github.com/go-logr/zapr"
	"github.com/jlewi/foyle/protos/go/foyle/logs/logspbconnect"
	"github.com/maxence-charriere/go-app/v9/pkg/app"
	"go.uber.org/zap"
	"net/http"
	"strings"
)

var (
	defaultClient logspbconnect.LogsServiceClient
)

func GetClient() logspbconnect.LogsServiceClient {
	if defaultClient == nil {
		log := zapr.NewLogger(zap.L())
		// Because of CORS we need the baseHref to match whatever origin we are accessing the server on
		// e.g. 127.0.0.1 vs. localhost
		// So we get the URL of the window.
		// We then strip the AppPath from it. This is the path where the app is served from.
		// Then we add whatever prefix is used for serving the API which is passed in from the server.
		baseHref := app.Window().URL().String()
		baseURL := baseHref

		baseURL = strings.TrimSuffix(baseURL, AppPath+"/")
		apiPrefix := app.Getenv(APIPrefixEnvVar)
		baseURL = strings.TrimSuffix(baseURL, "/")
		apiPrefix = strings.TrimPrefix(apiPrefix, "/")
		baseURL += "/" + apiPrefix

		log.Info("Creating logs client", "baseURL", baseURL, "baseHREF", baseHref, "APIPrefix", apiPrefix)

		defaultClient = logspbconnect.NewLogsServiceClient(http.DefaultClient, baseURL)
		// TODO(jeremy): Should we try sending a status request to see if its working?
	}
	return defaultClient
}
