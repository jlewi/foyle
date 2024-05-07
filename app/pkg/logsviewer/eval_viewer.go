package logsviewer

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"connectrpc.com/connect"
	"github.com/go-logr/zapr"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1/v1alpha1connect"
	"github.com/maxence-charriere/go-app/v9/pkg/app"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
)

type evalViews string

const (
	loadEvalResults = "/loadEvalResults"
	databaseInputID = "databaseInput"

	setEvalView = "/setEvalView"

	evalQueryView          evalViews = "evalQueryView"
	evalActualAnswerView   evalViews = "evalActualAnswerView"
	evalExpectedAnswerView evalViews = "evalExpectedAnswerView"
	evalRawView            evalViews = "evalRawView"
)

var (
	// resultSet keeps track of the current loaded result set. This allows us to easily access it from multiple
	// elements. I guess we could also pass it around using go-app context but this seems easier.
	resultSet *ResultSet
)

// EvalViewer is the page that displays an eval result.
type EvalViewer struct {
	app.Compo
	main         *evalView
	resultsTable *EvalResultsTable
}

func (c *EvalViewer) Render() app.UI {
	if c.main == nil {
		c.main = &evalView{}
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
				ctx.NewActionWithValue(setEvalView, evalQueryView)
			}),
		),
		app.Div().Body(
			app.Button().Text("Actual Answer").OnClick(func(ctx app.Context, e app.Event) {
				ctx.NewActionWithValue(setEvalView, evalActualAnswerView)
			}),
		),
		app.Div().Body(
			app.Button().Text("Expected Answer")).OnClick(func(ctx app.Context, e app.Event) {
			ctx.NewActionWithValue(setEvalView, evalExpectedAnswerView)
		}),
		app.Div().Body(
			app.Button().Text("Raw")).OnClick(func(ctx app.Context, e app.Event) {
			ctx.NewActionWithValue(setEvalView, evalRawView)
		}))
}

type EvalResultsTable struct {
	app.Compo
	SelectedRow int
}

func (c *EvalResultsTable) OnMount(ctx app.Context) {
	ctx.Handle(loadEvalResults, c.handleLoadEvalResults)
}

func (c *EvalResultsTable) handleLoadEvalResults(ctx app.Context, action app.Action) {
	log := zapr.NewLogger(zap.L())
	log.Info("Handling loadEvalResults")

	if resultSet == nil {
		resultSet = &ResultSet{}
	}

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

	resultSet.data = res.Msg.Items
	resultSet.selected = 0
	log.Info("Loaded eval results", "numResults", len(resultSet.data), "instance", c)
	c.SelectedRow = 1
	c.Update()
}

func (c *EvalResultsTable) Render() app.UI {
	log := zapr.NewLogger(zap.L())
	log.Info("Rendering EvalResultsTable", "instance", c)
	if resultSet == nil {
		log.Info("Data is nil", "instance", c)
		resultSet = &ResultSet{}
	}

	if resultSet.data == nil {
		resultSet.data = make([]*v1alpha1.EvalResult, 0)
	}

	table := app.Table().Body(
		app.Tr().Body(
			app.Th().Text("ID"),
			app.Th().Text("File"),
			app.Th().Text("Distance"),
			app.Th().Text("Normalized Distance"),
		),
		app.Range(resultSet.data).Slice(func(i int) app.UI {
			rowStyle := ""
			if i == c.SelectedRow {
				rowStyle = "selected-row" // This is a CSS class that you need to define
			}
			row := app.Tr().Class(rowStyle).Body(
				app.Td().Text(resultSet.data[i].GetExample().GetId()),
				app.Td().Text(resultSet.data[i].GetExampleFile()),
				app.Td().Text(resultSet.data[i].GetDistance()),
				app.Td().Text(resultSet.data[i].GetNormalizedDistance()),
			)

			// For each row we add a click handler to display the corresponding example.
			row.OnClick(func(ctx app.Context, e app.Event) {
				log := zapr.NewLogger(zap.L())
				log.Info("Selected row", "row", i)
				// Mark the selected row and trigger the update.
				// This will redraw the table and change the style on the selected row.
				c.SelectedRow = i
				resultSet.selected = i
				c.Update()

				// TODO(jeremy): We should fire an event and change the context to display the evaluation result.
			})
			return row
		}),
	)
	div := app.Div().Class("scrollable-table").Body(table)
	return div
}

// evalView is the main viewer of the evaluation viewer.
// What it displays will change depending on the view selected.
// The content of the window is HTML which gets set by the action handler for different events.
//
// The view registers a handler for the setEvalViewAction event. The setEvalViewAction event is triggered when ever
// the view needs to be changed; e.g. because the view has changed or the selected data has changed
type evalView struct {
	app.Compo
	HTMLContent string
}

func (m *evalView) Render() app.UI {
	// Raw requires the value to have a single root element. So we enclose the HTML content in a div to ensure
	// that is all ways true.
	return app.Raw("<div>" + m.HTMLContent + "</div>")
}

func (m *evalView) OnMount(ctx app.Context) {
	ctx.Handle(setEvalView, m.handleSetEvalView)
}

func (m *evalView) handleSetEvalView(ctx app.Context, action app.Action) {
	log := zapr.NewLogger(zap.L())
	viewValue, ok := action.Value.(evalViews) // Checks if a name was given.
	if !ok {
		log.Error(errors.New("No view provided"), "Invalid action")
		return
	}
	log.Info("Handling get action", "view", viewValue)
	switch viewValue {
	case evalQueryView:
		current := resultSet.GetSelected()
		if current == nil {
			m.HTMLContent = "No evaluation result is currently selected"
			break
		}
		value, err := docToHTML(current.Example.Query)
		if err == nil {
			m.HTMLContent = value
		} else {
			log.Error(err, "Failed to convert generated block to html")
			m.HTMLContent = fmt.Sprintf("Failed to convert generated block to html : error %+v", err)
		}
	case evalActualAnswerView:
		current := resultSet.GetSelected()
		if current == nil {
			m.HTMLContent = "No evaluation result is currently selected"
			break
		}
		doc := &v1alpha1.Doc{
			Blocks: current.Actual,
		}
		value, err := docToHTML(doc)
		if err == nil {
			m.HTMLContent = value
		} else {
			log.Error(err, "Failed to convert actual answer to html")
			m.HTMLContent = fmt.Sprintf("Failed to convert actual answer to html : error %+v", err)
		}
	case evalExpectedAnswerView:
		current := resultSet.GetSelected()
		if current == nil {
			m.HTMLContent = "No evaluation result is currently selected"
			break
		}
		doc := &v1alpha1.Doc{
			Blocks: current.Example.Answer,
		}
		value, err := docToHTML(doc)
		if err == nil {
			m.HTMLContent = value
		} else {
			log.Error(err, "Failed to convert expected blocks to html")
			m.HTMLContent = fmt.Sprintf("Failed to convert expected blocks to html : error %+v", err)
		}
	case evalRawView:
		current := resultSet.GetSelected()
		if current == nil {
			m.HTMLContent = "No evaluation result is currently selected"
			break
		}
		marshaler := protojson.MarshalOptions{
			Indent: "  ",
		}
		blockJson, err := marshaler.Marshal(current)
		if err != nil {
			log.Error(err, "Failed to turn result into json")
			m.HTMLContent = fmt.Sprintf("Failed to turn blocklog into json: error %+v", err)
		} else {
			raw := "<pre>" + string(blockJson) + "</pre>"
			m.HTMLContent = raw
		}
	default:
		m.HTMLContent = "Unknown view: " + string(viewValue)
	}
	// We need to call update to trigger a re-render of the component.
	m.Update()
}

// ResultSet keeps track of the current loaded result set. This allows us to easily access it from multiple components.
// N.B. we also might want to wrap the data with accessors so we can access data in a thread safe way
type ResultSet struct {
	data     []*v1alpha1.EvalResult
	selected int
}

func (c *ResultSet) GetSelected() *v1alpha1.EvalResult {
	if c.data == nil || len(c.data) == 0 {
		return nil
	}
	if c.selected < 0 || c.selected >= len(c.data) {
		return nil
	}
	return c.data[c.selected]
}
