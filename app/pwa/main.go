package main

import (
	"github.com/jlewi/foyle/app/pkg/logsviewer"
	"log"
	"net/http"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

// The main function is the entry point where the app is configured and started.
// It is executed in 2 different environments: A client (the web browser) and a
// server.
func main() {
	// The first thing to do is to associate the hello component with a path.
	//
	// This is done by calling the Route() function,  which tells go-app what
	// component to display for a given path, on both client and server-side.
	app.Route("/", &logsviewer.MainApp{})

	// Once the routes set up, the next thing to do is to either launch the app
	// or the server that serves the app.
	//
	// When executed on the client-side, the RunWhenOnBrowser() function
	// launches the app,  starting a loop that listens for app events and
	// executes client instructions. Since it is a blocking call, the code below
	// it will never be executed.
	//
	// When executed on the server-side, RunWhenOnBrowser() does nothing, which
	// lets room for server implementation without the need for precompiling
	// instructions.
	app.RunWhenOnBrowser()

	// N.B. This code isn't actually used to serve because we serve it off our existing gin server defined in the
	// server pkg. But its useful for debugging/
	http.Handle("/", http.StripPrefix("/viewer", &app.Handler{
		Name:        "Hello",
		Description: "An Hello World! example",
		Resources:   app.CustomProvider("", "/viewer"),
		Styles: []string{
			"/web/viewer.css", // Loads traceSelector.css file.
		},
		Env: map[string]string{
			logsviewer.EndpointEnvVar: "http://localhost:8000",
		},
	}))

	if err := http.ListenAndServe(":8000", nil); err != nil {
		log.Fatal(err)
	}
}
