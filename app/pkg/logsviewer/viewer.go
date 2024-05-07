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
	main *mainWindow
}

func (c *MainApp) Render() app.UI {
	if c.main == nil {
		c.main = &mainWindow{}
	}
	return app.Div().Class("main-layout").Body(
		app.Div().Class("header").Body(
			&blockSelector{},
		),
		app.Div().Class("content").Body(
			app.Div().Class("sidebar").Body(
				&navigationBar{},
			),
			app.Div().Class("sidebar").Body(
				&sideBar{},
			),
			app.Div().Class("main-window").Body(
				c.main,
			),
		),
	)
}

// sideBar adds a navigation bar between the views to the left side.
type sideBar struct {
	app.Compo
}

func (s *sideBar) Render() app.UI {
	return app.Div().Body(
		// Each button needs to be enclosed in a div. Otherwise events get triggered for all the buttons.
		app.Div().Body(
			app.Button().Text("Generated Block").OnClick(func(ctx app.Context, e app.Event) {
				ctx.NewActionWithValue(getAction, generatedBlockView)
			}),
		),
		app.Div().Body(
			app.Button().Text("Executed Block")).OnClick(func(ctx app.Context, e app.Event) {
			ctx.NewActionWithValue(getAction, executedBlockView)
		}),
		app.Div().Body(
			app.Button().Text("Raw")).OnClick(func(ctx app.Context, e app.Event) {
			ctx.NewActionWithValue(getAction, rawView)
		}))
}
