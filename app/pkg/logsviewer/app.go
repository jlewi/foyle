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

// MainApp is the main window of the application.
//
// The main application consists of a left hand navigation bar and a right hand side component that is the page
// to display. When you click on one of the left hand navigation buttons it fires of an action setPage to change the
// view. The handler for this action loads the appropriate page and sets MainApp.page to the component for that
// page.
type MainApp struct {
	app.Compo
	// Page keeps track of the page to display in the right hand side.
	page app.UI
}

func (m *MainApp) Render() app.UI {
	if m.page == nil {
		// TODO(jeremy): Could we keep track of the last view so if we refresh we show the same data?
		// One way to do that is to update the URL with query arguments containing the relevant state information.
		// Then when we click refresh we could get the information directly from the URL
		m.page = &BlockViewer{}
	}
	return app.Div().Class("main-layout").Body(
		app.Div().Class("content").Body(
			app.Div().Class("sidebar").Body(
				&navigationBar{},
			),
			app.Div().Class("page-window").Body(
				m.page,
			),
		), &StatusBar{},
	)
}

func (m *MainApp) OnMount(ctx app.Context) {
	// register to handle the setPage action
	ctx.Handle(setPage, m.handleSetPage)
}

// handleSetPage handles the setPage action. The event will tell us which view to display.
func (m *MainApp) handleSetPage(ctx app.Context, action app.Action) {
	log := zapr.NewLogger(zap.L())
	pageValue, ok := action.Value.(page)
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

// StatusBar at the bottom of the page. Inspired by the vscode/intellij status bar.
// We use this to show useful information like the version number.
type StatusBar struct {
	app.Compo
}

func (s *StatusBar) Render() app.UI {
	version := app.Getenv("GOAPP_VERSION")
	return app.Div().Class("status-bar").Text("goapp version: " + version)
}
