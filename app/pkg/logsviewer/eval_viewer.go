package logsviewer

import (
	"fmt"
	"github.com/go-logr/zapr"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	"github.com/maxence-charriere/go-app/v9/pkg/app"
	"go.uber.org/zap"
)

// EvalViewer is the page that displays an eval result.
type EvalViewer struct {
	app.Compo
	main *mainWindow
}

func (c *EvalViewer) Render() app.UI {
	if c.main == nil {
		c.main = &mainWindow{}
	}
	return app.Div().Class("main-layout").Body(
		app.Div().Class("header").Body(
			&EvalResultsTable{
				SelectedRow: 1,
			},
		),
		app.Div().Class("content").Body(
			app.Div().Class("sidebar").Body(
				&evalSideBar{},
			),
			app.Div().Class("main-window").Body(
				c.main,
			),
		),
	)
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

func (c *EvalResultsTable) Render() app.UI {
	c.Data = make([]*v1alpha1.EvalResult, 0, 20)

	for i := 0; i < 20; i++ {
		c.Data = append(c.Data, &v1alpha1.EvalResult{
			Example: &v1alpha1.Example{
				Id: fmt.Sprintf("%d", i),
			},
			ExampleFile:        fmt.Sprintf("file%d", i),
			Distance:           1,
			NormalizedDistance: 0.2,
		})
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
