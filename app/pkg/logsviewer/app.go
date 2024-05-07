package logsviewer

import (
	"github.com/go-logr/zapr"
	"github.com/maxence-charriere/go-app/v9/pkg/app"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type page string
type view string

const (
	getAction = "/get"
	setPage   = "/setPage"

	errorView          view = "error"
	generatedBlockView view = "generatedBlock"
	executedBlockView  view = "executedBlock"
	rawView            view = "raw"

	blockLogsView page = "blockLogs"
	evalsView     page = "evals"

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
	// Page keeps track of the page to display in the right hand side.
	page app.UI
}

func (m *MainApp) Render() app.UI {
	if m.page == nil {
		// Default to the Blockvier
		m.page = &BlockViewer{}
	}
	return app.Div().Class("main-layout").Body(
		app.Div().Class("content").Body(
			app.Div().Class("sidebar").Body(
				&navigationBar{},
			),
			app.Div().Class("page-window").Body(
				// TODO(jeremy): How do we change this when the user clicks the left hand navigation bar?
				// Do we need to find and update the div?
				m.page,
			),
		),
	)
}

func (m *MainApp) OnMount(ctx app.Context) {
	// register to handle the setPage action
	ctx.Handle(setPage, m.handleSetPage)
}

// handleSetPage handles the setPage action. The event will tell us which view to display.
func (m *MainApp) handleSetPage(ctx app.Context, action app.Action) {
	log := zapr.NewLogger(zap.L())
	pageValue, ok := action.Value.(page) // Checks if a name was given.
	if !ok {
		log.Error(errors.New("No page provided"), "Invalid action")
		return
	}
	log.Info("Handling set page action", "page", pageValue)
	switch pageValue {
	case blockLogsView:
		if _, ok := m.page.(*BlockViewer); !ok {
			log.Info("Setting page to BlockViewer")
			m.page = &BlockViewer{}
		}
	case evalsView:
		if _, ok := m.page.(*EvalViewer); !ok {
			log.Info("Setting page to EvalViewer")
			m.page = &EvalViewer{}
		}
	}
	// We need to call update to trigger a re-render of the component.
	m.Update()
}
