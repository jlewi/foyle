package logsviewer

import "github.com/maxence-charriere/go-app/v9/pkg/app"

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
			&blockSelector{},
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
