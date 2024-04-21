package logsviewer

import "github.com/maxence-charriere/go-app/v9/pkg/app"

// Viewer is a web component to display the logs.
type Viewer struct {
	app.Compo
}

// The Render method is where the component appearance is defined. Here, a
// "Hello World!" is displayed as a heading.
func (h *Viewer) Render() app.UI {
	return app.H1().Text("Hello World!")
}
