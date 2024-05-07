package logsviewer

import (
	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type view string

const (
	getAction = "/et"

	errorView          view = "error"
	generatedBlockView view = "generatedBlock"
	executedBlockView  view = "executedBlock"
	rawView            view = "raw"

	getErrorState = "/getError"
	blockLogState = "/blocklog"
)

// How it works:
// Clicking load fetches the blocklog from the server.
// The log is then stored in the application context (https://go-app.dev/states)
// this allows other components to use it. Load then fires off an UpdateView event to trigger
// the mainWindow to update its content.
// The UpdateView event takes a string argument which is what view should be rendered.
// There is a left hand navigation bar  with buttons to display different views of the current log.
// Changing the view is achieved by sending UpdateView events to change the view

// MainApp is the main window of the application.
type MainApp struct {
	app.Compo
}

func (c *MainApp) Render() app.UI {
	return app.Div().Class("main-layout").Body(
		app.Div().Class("content").Body(
			app.Div().Class("sidebar").Body(
				&navigationBar{},
			),
			app.Div().Class("page-window").Body(
				// TODO(jeremy): How do we change this when the user clicks the left hand navigation bar?
				// Do we need to find and update the div?
				&BlockViewer{},
			),
		),
	)
}
