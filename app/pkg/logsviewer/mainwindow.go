package logsviewer

import (
	"encoding/json"
	"fmt"

	"github.com/go-logr/zapr"
	"github.com/jlewi/foyle/app/api"
	"github.com/maxence-charriere/go-app/v9/pkg/app"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// blockLogView is the main window of the application.
// What it displays will change depending on the view selected.
// The content of the main window is HTML which gets set by the action handler for different events.
//
// The main window registers a handler for the getAction event. The getAction event is triggered when ever
// a blockLog is loaded. The handler for the getAction event will set the HTML content of the main windowß
type blockLogView struct {
	app.Compo
	HTMLContent string
}

func (m *blockLogView) Render() app.UI {
	// Raw requires the value to have a single root element. So we enclose the HTML content in a div to ensure
	// that is all ways true.
	return app.Raw("<div>" + m.HTMLContent + "</div>")
}

func (m *blockLogView) OnMount(ctx app.Context) {
	ctx.Handle(getAction, m.handleGetAction)
}

func (m *blockLogView) handleGetAction(ctx app.Context, action app.Action) {
	log := zapr.NewLogger(zap.L())
	viewValue, ok := action.Value.(view) // Checks if a name was given.
	if !ok {
		log.Error(errors.New("No view provided"), "Invalid action")
		return
	}
	log.Info("Handling get action", "view", viewValue)
	switch viewValue {
	case errorView:
		errState := ""
		ctx.GetState(getErrorState, &errState)

		m.HTMLContent = "<p>Error getting blocklog:</p><br> " + errState
	case generatedBlockView:
		block := &api.BlockLog{}
		ctx.GetState(blockLogState, block)
		value, err := renderGeneratedBlock(block)
		if err == nil {
			m.HTMLContent = value
		} else {
			log.Error(err, "Failed to convert generated block to html")
			m.HTMLContent = fmt.Sprintf("Failed to convert generated block to html : error %+v", err)
		}
	case executedBlockView:
		block := &api.BlockLog{}
		ctx.GetState(blockLogState, block)
		value, err := renderExecutedBlock(block)
		if err == nil {
			m.HTMLContent = value
		} else {
			log.Error(err, "Failed to convert executed block to html")
			m.HTMLContent = fmt.Sprintf("Failed to convert executed block to html: error %+v", err)
		}
	case rawView:
		block := &api.BlockLog{}
		ctx.GetState(blockLogState, block)
		blockJson, err := json.MarshalIndent(block, "", " ")
		if err != nil {
			log.Error(err, "Failed to turn blocklog into json")
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
