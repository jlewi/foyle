package docs

import (
	"bytes"
	"connectrpc.com/connect"
	"context"
	"github.com/jlewi/foyle/app/pkg/logs"
	logspb "github.com/jlewi/foyle/protos/go/foyle/logs"
	"github.com/pkg/errors"
	"github.com/yuin/goldmark"
)

// ConvertersService is a handler for conversion operations of documents.
// It is used as part of operating/observing Foyle. Concretely, we can use to convert requests into markdown/html
type ConvertersService struct {
}

func NewConvertersService() (*ConvertersService, error) {
	return &ConvertersService{}, nil
}

func (s *ConvertersService) ConvertDoc(ctx context.Context, request *connect.Request[logspb.ConvertDocRequest]) (*connect.Response[logspb.ConvertDocResponse], error) {
	log := logs.FromContext(ctx)
	doc := request.Msg.GetDoc()
	if doc == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("No doc provided"))
	}
	format := request.Msg.GetFormat()

	docResponse := &logspb.ConvertDocResponse{}
	md := DocToMarkdown(doc)

	if format == logspb.ConvertDocRequest_MARKDOWN {
		docResponse.Text = md
		return connect.NewResponse(docResponse), nil
	}

	// Convert to HTML
	var buf bytes.Buffer
	if err := goldmark.Convert([]byte(md), &buf); err != nil {
		log.Error(err, "Failed to convert markdown to HTML")
		return nil, connect.NewError(connect.CodeInternal, errors.Wrap(err, "Failed to convert markdown to HTML"))
	}

	if format == logspb.ConvertDocRequest_HTML {
		docResponse.Text = buf.String()
		return connect.NewResponse(docResponse), nil
	}

	return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("Unrecognized format"))
}
