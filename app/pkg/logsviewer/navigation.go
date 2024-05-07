package logsviewer

import "github.com/maxence-charriere/go-app/v9/pkg/app"

// navigationBar is the left hand side navigation bar.
// It is used to select the different pages in the application
type navigationBar struct {
	app.Compo
}

func (s *navigationBar) Render() app.UI {
	return app.Div().Body(
		// Each button needs to be enclosed in a div. Otherwise events get triggered for all the buttons.
		app.Div().Body(
			app.Button().Text("Block Logs").OnClick(func(ctx app.Context, e app.Event) {
				//ctx.NewActionWithValue(getAction, generatedBlockView)
			}),
		),
		app.Div().Body(
			app.Button().Text("Eval Results")).OnClick(func(ctx app.Context, e app.Event) {
			//ctx.NewActionWithValue(getAction, executedBlockView)
		}),
		//app.Div().Body(
		//	app.Button().Text("Raw")).OnClick(func(ctx app.Context, e app.Event) {
		//	ctx.NewActionWithValue(getAction, rawView)
		//})
	)
}
