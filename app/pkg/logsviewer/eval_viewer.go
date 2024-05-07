package logsviewer

import (
	"connectrpc.com/connect"
	"context"
	"github.com/go-logr/zapr"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1/v1alpha1connect"
	"github.com/maxence-charriere/go-app/v9/pkg/app"
	"go.uber.org/zap"
	"net/http"
	"strings"
)

const (
	loadEvalResults = "/loadEvalResults"
	databaseInputID = "databaseInput"
)

// EvalViewer is the page that displays an eval result.
type EvalViewer struct {
	app.Compo
	main         *mainWindow
	resultsTable *EvalResultsTable
}

func (c *EvalViewer) Render() app.UI {
	if c.main == nil {
		c.main = &mainWindow{}
	}
	if c.resultsTable == nil {
		c.resultsTable = &EvalResultsTable{}
	}
	loadButton := app.Button().
		Text("Load").
		OnClick(func(ctx app.Context, e app.Event) {
			ctx.NewAction(loadEvalResults)
		})

	// The top row will contain the input box to select the database
	// and the results table to scroll though them.
	// These will be arranged vertically in the row
	topRow := app.Div().Class("row").Body(
		app.Div().Body(
			app.Input().
				Type("text").
				ID(databaseInputID).
				Value("/Users/jlewi/foyle_experiments/learning"),
			loadButton,
		),
		app.Div().Body(
			c.resultsTable,
		))

	// The bottom row will contain the main window.
	bottomRow := app.Div().Class("row").Body(
		app.Div().Class("sidebar").Body(
			&evalSideBar{},
		),
		app.Div().Class("main-window").Body(
			c.main,
		),
	)
	return app.Div().Body(topRow, bottomRow)
}

// evalSideBar adds a navigation bar between the views to the left side.
type evalSideBar struct {
	app.Compo
}

func (s *evalSideBar) Render() app.UI {
	return app.Div().Body(
		// Each button needs to be enclosed in a div. Otherwise events get triggered for all the buttons.
		app.Div().Body(
			app.Button().Text("Query").OnClick(func(ctx app.Context, e app.Event) {
				//ctx.NewActionWithValue(getAction, generatedBlockView)
			}),
		),
		app.Div().Body(
			app.Button().Text("Actual Answer").OnClick(func(ctx app.Context, e app.Event) {
				//ctx.NewActionWithValue(getAction, generatedBlockView)
			}),
		),
		app.Div().Body(
			app.Button().Text("Expected Answer")).OnClick(func(ctx app.Context, e app.Event) {
			//ctx.NewActionWithValue(getAction, executedBlockView)
		}),
		app.Div().Body(
			app.Button().Text("Raw")).OnClick(func(ctx app.Context, e app.Event) {
			//ctx.NewActionWithValue(getAction, rawView)
		}))
}

type EvalResultsTable struct {
	app.Compo
	Data        []*v1alpha1.EvalResult
	SelectedRow int
}

func (c *EvalResultsTable) OnMount(ctx app.Context) {
	ctx.Handle(loadEvalResults, c.handleLoadEvalResults)
}

func (c *EvalResultsTable) handleLoadEvalResults(ctx app.Context, action app.Action) {
	log := zapr.NewLogger(zap.L())
	log.Info("Handling loadEvalResults")

	database := app.Window().GetElementByID(databaseInputID).Get("value").String()
	database = strings.TrimSpace(database)
	if database == "" {
		// TODO(jeremy): We should surface an error message. We could use the status bar to show the error message
		log.Info("No database provided; can't load eval results")
		return
	}

	// TODO(jeremy): We should cache the client; see GetClient in block_viewer.go for an example.
	client := v1alpha1connect.NewEvalServiceClient(
		http.DefaultClient,
		// TODO(jeremy): How should we make this configurable?
		"http://localhost:8080/api",
	)

	listRequest := &v1alpha1.EvalResultListRequest{
		// TODO(jeremy): We need a UI element to let you enter this
		Database: database,
	}

	res, err := client.List(
		context.Background(),
		connect.NewRequest(listRequest),
	)

	if err != nil {
		log.Error(err, "Error listing eval results")
		return
	}

	c.Data = res.Msg.Items
	log.Info("Loaded eval results", "numResults", len(c.Data), "instance", c)
	c.SelectedRow = 1
	c.Update()
}

func (c *EvalResultsTable) Render() app.UI {
	log := zapr.NewLogger(zap.L())
	log.Info("Rendering EvalResultsTable", "instance", c)
	if c.Data == nil {
		log.Info("Data is nil", "instance", c)
		c.Data = make([]*v1alpha1.EvalResult, 0)
	}

	table := app.Table().Body(
		app.Tr().Body(
			app.Th().Text("ID"),
			app.Th().Text("File"),
			app.Th().Text("Distance"),
			app.Th().Text("Normalized Distance"),
		),
		app.Range(c.Data).Slice(func(i int) app.UI {
			rowStyle := ""
			if i == c.SelectedRow {
				rowStyle = "selected-row" // This is a CSS class that you need to define
			}
			row := app.Tr().Class(rowStyle).Body(
				app.Td().Text(c.Data[i].GetExample().GetId()),
				app.Td().Text(c.Data[i].GetExampleFile()),
				app.Td().Text(c.Data[i].GetDistance()),
				app.Td().Text(c.Data[i].GetNormalizedDistance()),
			)

			// For each row we add a click handler to display the corresponding example.
			row.OnClick(func(ctx app.Context, e app.Event) {
				log := zapr.NewLogger(zap.L())
				log.Info("Selected row", "row", i)
				// Mark the selected row and trigger the update.
				// This will redraw the table and change the style on the selected row.
				c.SelectedRow = i
				c.Update()

				// TODO(jeremy): We should fire an event and change the context to display the evaluation result.
			})
			return row
		}),
	)
	div := app.Div().Class("scrollable-table").Body(table)
	return div
}
