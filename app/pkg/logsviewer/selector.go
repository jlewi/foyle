package logsviewer

import (
	"github.com/go-logr/zapr"
	"github.com/maxence-charriere/go-app/v9/pkg/app"
	"go.uber.org/zap"
	"strings"
)

const (
	blockInputID = "blockId"
)

// blockSelector is a component that lets the user select which block log to view
type blockSelector struct {
	app.Compo
	traceValue string
}

// The Render method is where the component appearance is defined. Here, a
// "Hello World!" is displayed as a heading.
func (h *blockSelector) Render() app.UI {
	return app.Div().Body(
		// TODO(jeremy): Should we use an environment variable to set the default value?
		// So that we can have the backend set the default block to one that exists?
		app.Input().
			Type("text").
			ID(blockInputID).
			Value("8761dedd-f7f5-476e-ab20-2204f9c91afb"),
		app.Button().
			Text("Display").
			OnClick(func(ctx app.Context, e app.Event) {
				// Handle button click event here
				// TODO(jeremy): Using os.Getenv is not working as a way of passing values from the server to the
				// client.
				// endpoint := os.Getenv(EndpointEnvVar)
				endpoint := "http://localhost:8080/"
				client := LogsClient{
					Endpoint: endpoint,
				}
				blockID := app.Window().GetElementByID(blockInputID).Get("value").String()
				blockID = strings.TrimSpace(blockID)
				if blockID == "" {
					h.traceValue = "No Block ID provided"
					h.Update()
					return
				}
				log := zapr.NewLogger(zap.L())
				log.Info("Fetching block log", "blockID", blockID, "endpoint", endpoint)
				blockLog, err := client.GetBlockLog(ctx, blockID)
				if err != nil {
					ctx.SetState(getErrorState, err.Error())
					ctx.NewActionWithValue(getAction, errorView)
				} else {
					ctx.SetState(blockLogState, blockLog)
					ctx.NewActionWithValue(getAction, rawView)
				}
			}),
	)
}
