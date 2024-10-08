// Code generated by protoc-gen-connect-go. DO NOT EDIT.
//
// Source: foyle/logs/conversion.proto

package logspbconnect

import (
	connect "connectrpc.com/connect"
	context "context"
	errors "errors"
	logs "github.com/jlewi/foyle/protos/go/foyle/logs"
	http "net/http"
	strings "strings"
)

// This is a compile-time assertion to ensure that this generated file and the connect package are
// compatible. If you get a compiler error that this constant is not defined, this code was
// generated with a version of connect newer than the one compiled into your binary. You can fix the
// problem by either regenerating this code with an older version of connect or updating the connect
// version compiled into your binary.
const _ = connect.IsAtLeastVersion1_13_0

const (
	// ConversionServiceName is the fully-qualified name of the ConversionService service.
	ConversionServiceName = "foyle.logs.ConversionService"
)

// These constants are the fully-qualified names of the RPCs defined in this package. They're
// exposed at runtime as Spec.Procedure and as the final two segments of the HTTP route.
//
// Note that these are different from the fully-qualified method names used by
// google.golang.org/protobuf/reflect/protoreflect. To convert from these constants to
// reflection-formatted method names, remove the leading slash and convert the remaining slash to a
// period.
const (
	// ConversionServiceConvertDocProcedure is the fully-qualified name of the ConversionService's
	// ConvertDoc RPC.
	ConversionServiceConvertDocProcedure = "/foyle.logs.ConversionService/ConvertDoc"
)

// These variables are the protoreflect.Descriptor objects for the RPCs defined in this package.
var (
	conversionServiceServiceDescriptor          = logs.File_foyle_logs_conversion_proto.Services().ByName("ConversionService")
	conversionServiceConvertDocMethodDescriptor = conversionServiceServiceDescriptor.Methods().ByName("ConvertDoc")
)

// ConversionServiceClient is a client for the foyle.logs.ConversionService service.
type ConversionServiceClient interface {
	// ConvertDoc converts a doc representation of a notebook into markdown or HTML
	ConvertDoc(context.Context, *connect.Request[logs.ConvertDocRequest]) (*connect.Response[logs.ConvertDocResponse], error)
}

// NewConversionServiceClient constructs a client for the foyle.logs.ConversionService service. By
// default, it uses the Connect protocol with the binary Protobuf Codec, asks for gzipped responses,
// and sends uncompressed requests. To use the gRPC or gRPC-Web protocols, supply the
// connect.WithGRPC() or connect.WithGRPCWeb() options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewConversionServiceClient(httpClient connect.HTTPClient, baseURL string, opts ...connect.ClientOption) ConversionServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	return &conversionServiceClient{
		convertDoc: connect.NewClient[logs.ConvertDocRequest, logs.ConvertDocResponse](
			httpClient,
			baseURL+ConversionServiceConvertDocProcedure,
			connect.WithSchema(conversionServiceConvertDocMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
	}
}

// conversionServiceClient implements ConversionServiceClient.
type conversionServiceClient struct {
	convertDoc *connect.Client[logs.ConvertDocRequest, logs.ConvertDocResponse]
}

// ConvertDoc calls foyle.logs.ConversionService.ConvertDoc.
func (c *conversionServiceClient) ConvertDoc(ctx context.Context, req *connect.Request[logs.ConvertDocRequest]) (*connect.Response[logs.ConvertDocResponse], error) {
	return c.convertDoc.CallUnary(ctx, req)
}

// ConversionServiceHandler is an implementation of the foyle.logs.ConversionService service.
type ConversionServiceHandler interface {
	// ConvertDoc converts a doc representation of a notebook into markdown or HTML
	ConvertDoc(context.Context, *connect.Request[logs.ConvertDocRequest]) (*connect.Response[logs.ConvertDocResponse], error)
}

// NewConversionServiceHandler builds an HTTP handler from the service implementation. It returns
// the path on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewConversionServiceHandler(svc ConversionServiceHandler, opts ...connect.HandlerOption) (string, http.Handler) {
	conversionServiceConvertDocHandler := connect.NewUnaryHandler(
		ConversionServiceConvertDocProcedure,
		svc.ConvertDoc,
		connect.WithSchema(conversionServiceConvertDocMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	return "/foyle.logs.ConversionService/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case ConversionServiceConvertDocProcedure:
			conversionServiceConvertDocHandler.ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

// UnimplementedConversionServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedConversionServiceHandler struct{}

func (UnimplementedConversionServiceHandler) ConvertDoc(context.Context, *connect.Request[logs.ConvertDocRequest]) (*connect.Response[logs.ConvertDocResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("foyle.logs.ConversionService.ConvertDoc is not implemented"))
}
