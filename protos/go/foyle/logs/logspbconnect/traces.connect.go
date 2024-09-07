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
	// LogsServiceGetBlockLogProcedure is the fully-qualified name of the LogsService's GetBlockLog RPC.
	LogsServiceGetBlockLogProcedure = "/foyle.logs.LogsService/GetBlockLog"
	// LogsServiceGetLLMLogsProcedure is the fully-qualified name of the LogsService's GetLLMLogs RPC.
	LogsServiceGetLLMLogsProcedure = "/foyle.logs.LogsService/GetLLMLogs"
	// LogsServiceStatusProcedure is the fully-qualified name of the LogsService's Status RPC.
	LogsServiceStatusProcedure = "/foyle.logs.LogsService/Status"
)

// These variables are the protoreflect.Descriptor objects for the RPCs defined in this package.
var (
	logsServiceServiceDescriptor           = logs.File_foyle_logs_traces_proto.Services().ByName("LogsService")
	logsServiceGetTraceMethodDescriptor    = logsServiceServiceDescriptor.Methods().ByName("GetTrace")
	logsServiceGetBlockLogMethodDescriptor = logsServiceServiceDescriptor.Methods().ByName("GetBlockLog")
	logsServiceGetLLMLogsMethodDescriptor  = logsServiceServiceDescriptor.Methods().ByName("GetLLMLogs")
	logsServiceStatusMethodDescriptor      = logsServiceServiceDescriptor.Methods().ByName("Status")
)

// LogsServiceClient is a client for the foyle.logs.LogsService service.
type LogsServiceClient interface {
	GetTrace(context.Context, *connect.Request[logs.GetTraceRequest]) (*connect.Response[logs.GetTraceResponse], error)
	GetBlockLog(context.Context, *connect.Request[logs.GetBlockLogRequest]) (*connect.Response[logs.GetBlockLogResponse], error)
	// GetLLMLogs returns the logs associated with an LLM call.
	// These will include the rendered prompt and response. Unlike GetTraceRequest this has the
	// actual prompt and response of the LLM.
	GetLLMLogs(context.Context, *connect.Request[logs.GetLLMLogsRequest]) (*connect.Response[logs.GetLLMLogsResponse], error)
	Status(context.Context, *connect.Request[logs.GetLogsStatusRequest]) (*connect.Response[logs.GetLogsStatusResponse], error)
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
		getBlockLog: connect.NewClient[logs.GetBlockLogRequest, logs.GetBlockLogResponse](
			httpClient,
			baseURL+LogsServiceGetBlockLogProcedure,
			connect.WithSchema(logsServiceGetBlockLogMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
		getLLMLogs: connect.NewClient[logs.GetLLMLogsRequest, logs.GetLLMLogsResponse](
			httpClient,
			baseURL+LogsServiceGetLLMLogsProcedure,
			connect.WithSchema(logsServiceGetLLMLogsMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
		status: connect.NewClient[logs.GetLogsStatusRequest, logs.GetLogsStatusResponse](
			httpClient,
			baseURL+LogsServiceStatusProcedure,
			connect.WithSchema(logsServiceStatusMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
	}
}

// logsServiceClient implements LogsServiceClient.
type logsServiceClient struct {
	getTrace    *connect.Client[logs.GetTraceRequest, logs.GetTraceResponse]
	getBlockLog *connect.Client[logs.GetBlockLogRequest, logs.GetBlockLogResponse]
	getLLMLogs  *connect.Client[logs.GetLLMLogsRequest, logs.GetLLMLogsResponse]
	status      *connect.Client[logs.GetLogsStatusRequest, logs.GetLogsStatusResponse]
}

// GetTrace calls foyle.logs.LogsService.GetTrace.
func (c *logsServiceClient) GetTrace(ctx context.Context, req *connect.Request[logs.GetTraceRequest]) (*connect.Response[logs.GetTraceResponse], error) {
	return c.getTrace.CallUnary(ctx, req)
}

// GetBlockLog calls foyle.logs.LogsService.GetBlockLog.
func (c *logsServiceClient) GetBlockLog(ctx context.Context, req *connect.Request[logs.GetBlockLogRequest]) (*connect.Response[logs.GetBlockLogResponse], error) {
	return c.getBlockLog.CallUnary(ctx, req)
}

// GetLLMLogs calls foyle.logs.LogsService.GetLLMLogs.
func (c *logsServiceClient) GetLLMLogs(ctx context.Context, req *connect.Request[logs.GetLLMLogsRequest]) (*connect.Response[logs.GetLLMLogsResponse], error) {
	return c.getLLMLogs.CallUnary(ctx, req)
}

// Status calls foyle.logs.LogsService.Status.
func (c *logsServiceClient) Status(ctx context.Context, req *connect.Request[logs.GetLogsStatusRequest]) (*connect.Response[logs.GetLogsStatusResponse], error) {
	return c.status.CallUnary(ctx, req)
}

// LogsServiceHandler is an implementation of the foyle.logs.LogsService service.
type LogsServiceHandler interface {
	GetTrace(context.Context, *connect.Request[logs.GetTraceRequest]) (*connect.Response[logs.GetTraceResponse], error)
	GetBlockLog(context.Context, *connect.Request[logs.GetBlockLogRequest]) (*connect.Response[logs.GetBlockLogResponse], error)
	// GetLLMLogs returns the logs associated with an LLM call.
	// These will include the rendered prompt and response. Unlike GetTraceRequest this has the
	// actual prompt and response of the LLM.
	GetLLMLogs(context.Context, *connect.Request[logs.GetLLMLogsRequest]) (*connect.Response[logs.GetLLMLogsResponse], error)
	Status(context.Context, *connect.Request[logs.GetLogsStatusRequest]) (*connect.Response[logs.GetLogsStatusResponse], error)
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
	logsServiceGetBlockLogHandler := connect.NewUnaryHandler(
		LogsServiceGetBlockLogProcedure,
		svc.GetBlockLog,
		connect.WithSchema(logsServiceGetBlockLogMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	logsServiceGetLLMLogsHandler := connect.NewUnaryHandler(
		LogsServiceGetLLMLogsProcedure,
		svc.GetLLMLogs,
		connect.WithSchema(logsServiceGetLLMLogsMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	logsServiceStatusHandler := connect.NewUnaryHandler(
		LogsServiceStatusProcedure,
		svc.Status,
		connect.WithSchema(logsServiceStatusMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	return "/foyle.logs.LogsService/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case LogsServiceGetTraceProcedure:
			logsServiceGetTraceHandler.ServeHTTP(w, r)
		case LogsServiceGetBlockLogProcedure:
			logsServiceGetBlockLogHandler.ServeHTTP(w, r)
		case LogsServiceGetLLMLogsProcedure:
			logsServiceGetLLMLogsHandler.ServeHTTP(w, r)
		case LogsServiceStatusProcedure:
			logsServiceStatusHandler.ServeHTTP(w, r)
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

func (UnimplementedLogsServiceHandler) GetBlockLog(context.Context, *connect.Request[logs.GetBlockLogRequest]) (*connect.Response[logs.GetBlockLogResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("foyle.logs.LogsService.GetBlockLog is not implemented"))
}

func (UnimplementedLogsServiceHandler) GetLLMLogs(context.Context, *connect.Request[logs.GetLLMLogsRequest]) (*connect.Response[logs.GetLLMLogsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("foyle.logs.LogsService.GetLLMLogs is not implemented"))
}

func (UnimplementedLogsServiceHandler) Status(context.Context, *connect.Request[logs.GetLogsStatusRequest]) (*connect.Response[logs.GetLogsStatusResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("foyle.logs.LogsService.Status is not implemented"))
}
