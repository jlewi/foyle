// Code generated by protoc-gen-connect-go. DO NOT EDIT.
//
// Source: foyle/logs/traces.proto

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
	// LogsServiceName is the fully-qualified name of the LogsService service.
	LogsServiceName = "foyle.logs.LogsService"
)

// These constants are the fully-qualified names of the RPCs defined in this package. They're
// exposed at runtime as Spec.Procedure and as the final two segments of the HTTP route.
//
// Note that these are different from the fully-qualified method names used by
// google.golang.org/protobuf/reflect/protoreflect. To convert from these constants to
// reflection-formatted method names, remove the leading slash and convert the remaining slash to a
// period.
const (
	// LogsServiceGetTraceProcedure is the fully-qualified name of the LogsService's GetTrace RPC.
	LogsServiceGetTraceProcedure = "/foyle.logs.LogsService/GetTrace"
)

// These variables are the protoreflect.Descriptor objects for the RPCs defined in this package.
var (
	logsServiceServiceDescriptor        = logs.File_foyle_logs_traces_proto.Services().ByName("LogsService")
	logsServiceGetTraceMethodDescriptor = logsServiceServiceDescriptor.Methods().ByName("GetTrace")
)

// LogsServiceClient is a client for the foyle.logs.LogsService service.
type LogsServiceClient interface {
	// N.B. This is for testing only. Wanted to add a non streaming response which we can use to verify things are working.
	GetTrace(context.Context, *connect.Request[logs.GetTraceRequest]) (*connect.Response[logs.GetTraceResponse], error)
}

// NewLogsServiceClient constructs a client for the foyle.logs.LogsService service. By default, it
// uses the Connect protocol with the binary Protobuf Codec, asks for gzipped responses, and sends
// uncompressed requests. To use the gRPC or gRPC-Web protocols, supply the connect.WithGRPC() or
// connect.WithGRPCWeb() options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewLogsServiceClient(httpClient connect.HTTPClient, baseURL string, opts ...connect.ClientOption) LogsServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	return &logsServiceClient{
		getTrace: connect.NewClient[logs.GetTraceRequest, logs.GetTraceResponse](
			httpClient,
			baseURL+LogsServiceGetTraceProcedure,
			connect.WithSchema(logsServiceGetTraceMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
	}
}

// logsServiceClient implements LogsServiceClient.
type logsServiceClient struct {
	getTrace *connect.Client[logs.GetTraceRequest, logs.GetTraceResponse]
}

// GetTrace calls foyle.logs.LogsService.GetTrace.
func (c *logsServiceClient) GetTrace(ctx context.Context, req *connect.Request[logs.GetTraceRequest]) (*connect.Response[logs.GetTraceResponse], error) {
	return c.getTrace.CallUnary(ctx, req)
}

// LogsServiceHandler is an implementation of the foyle.logs.LogsService service.
type LogsServiceHandler interface {
	// N.B. This is for testing only. Wanted to add a non streaming response which we can use to verify things are working.
	GetTrace(context.Context, *connect.Request[logs.GetTraceRequest]) (*connect.Response[logs.GetTraceResponse], error)
}

// NewLogsServiceHandler builds an HTTP handler from the service implementation. It returns the path
// on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewLogsServiceHandler(svc LogsServiceHandler, opts ...connect.HandlerOption) (string, http.Handler) {
	logsServiceGetTraceHandler := connect.NewUnaryHandler(
		LogsServiceGetTraceProcedure,
		svc.GetTrace,
		connect.WithSchema(logsServiceGetTraceMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	return "/foyle.logs.LogsService/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case LogsServiceGetTraceProcedure:
			logsServiceGetTraceHandler.ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

// UnimplementedLogsServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedLogsServiceHandler struct{}

func (UnimplementedLogsServiceHandler) GetTrace(context.Context, *connect.Request[logs.GetTraceRequest]) (*connect.Response[logs.GetTraceResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("foyle.logs.LogsService.GetTrace is not implemented"))
}
