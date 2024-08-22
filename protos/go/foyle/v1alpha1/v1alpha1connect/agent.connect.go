// Code generated by protoc-gen-connect-go. DO NOT EDIT.
//
// Source: foyle/v1alpha1/agent.proto

package v1alpha1connect

import (
	connect "connectrpc.com/connect"
	context "context"
	errors "errors"
	v1alpha1 "github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
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
	// GenerateServiceName is the fully-qualified name of the GenerateService service.
	GenerateServiceName = "GenerateService"
	// ExecuteServiceName is the fully-qualified name of the ExecuteService service.
	ExecuteServiceName = "ExecuteService"
	// AIServiceName is the fully-qualified name of the AIService service.
	AIServiceName = "AIService"
)

// These constants are the fully-qualified names of the RPCs defined in this package. They're
// exposed at runtime as Spec.Procedure and as the final two segments of the HTTP route.
//
// Note that these are different from the fully-qualified method names used by
// google.golang.org/protobuf/reflect/protoreflect. To convert from these constants to
// reflection-formatted method names, remove the leading slash and convert the remaining slash to a
// period.
const (
	// GenerateServiceGenerateProcedure is the fully-qualified name of the GenerateService's Generate
	// RPC.
	GenerateServiceGenerateProcedure = "/GenerateService/Generate"
	// ExecuteServiceExecuteProcedure is the fully-qualified name of the ExecuteService's Execute RPC.
	ExecuteServiceExecuteProcedure = "/ExecuteService/Execute"
	// AIServiceStreamGenerateProcedure is the fully-qualified name of the AIService's StreamGenerate
	// RPC.
	AIServiceStreamGenerateProcedure = "/AIService/StreamGenerate"
	// AIServiceGenerateCellsProcedure is the fully-qualified name of the AIService's GenerateCells RPC.
	AIServiceGenerateCellsProcedure = "/AIService/GenerateCells"
	// AIServiceGetExampleProcedure is the fully-qualified name of the AIService's GetExample RPC.
	AIServiceGetExampleProcedure = "/AIService/GetExample"
	// AIServiceStatusProcedure is the fully-qualified name of the AIService's Status RPC.
	AIServiceStatusProcedure = "/AIService/Status"
)

// These variables are the protoreflect.Descriptor objects for the RPCs defined in this package.
var (
	generateServiceServiceDescriptor        = v1alpha1.File_foyle_v1alpha1_agent_proto.Services().ByName("GenerateService")
	generateServiceGenerateMethodDescriptor = generateServiceServiceDescriptor.Methods().ByName("Generate")
	executeServiceServiceDescriptor         = v1alpha1.File_foyle_v1alpha1_agent_proto.Services().ByName("ExecuteService")
	executeServiceExecuteMethodDescriptor   = executeServiceServiceDescriptor.Methods().ByName("Execute")
	aIServiceServiceDescriptor              = v1alpha1.File_foyle_v1alpha1_agent_proto.Services().ByName("AIService")
	aIServiceStreamGenerateMethodDescriptor = aIServiceServiceDescriptor.Methods().ByName("StreamGenerate")
	aIServiceGenerateCellsMethodDescriptor  = aIServiceServiceDescriptor.Methods().ByName("GenerateCells")
	aIServiceGetExampleMethodDescriptor     = aIServiceServiceDescriptor.Methods().ByName("GetExample")
	aIServiceStatusMethodDescriptor         = aIServiceServiceDescriptor.Methods().ByName("Status")
)

// GenerateServiceClient is a client for the GenerateService service.
type GenerateServiceClient interface {
	// Generate generates new cells given an existing document.
	Generate(context.Context, *connect.Request[v1alpha1.GenerateRequest]) (*connect.Response[v1alpha1.GenerateResponse], error)
}

// NewGenerateServiceClient constructs a client for the GenerateService service. By default, it uses
// the Connect protocol with the binary Protobuf Codec, asks for gzipped responses, and sends
// uncompressed requests. To use the gRPC or gRPC-Web protocols, supply the connect.WithGRPC() or
// connect.WithGRPCWeb() options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewGenerateServiceClient(httpClient connect.HTTPClient, baseURL string, opts ...connect.ClientOption) GenerateServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	return &generateServiceClient{
		generate: connect.NewClient[v1alpha1.GenerateRequest, v1alpha1.GenerateResponse](
			httpClient,
			baseURL+GenerateServiceGenerateProcedure,
			connect.WithSchema(generateServiceGenerateMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
	}
}

// generateServiceClient implements GenerateServiceClient.
type generateServiceClient struct {
	generate *connect.Client[v1alpha1.GenerateRequest, v1alpha1.GenerateResponse]
}

// Generate calls GenerateService.Generate.
func (c *generateServiceClient) Generate(ctx context.Context, req *connect.Request[v1alpha1.GenerateRequest]) (*connect.Response[v1alpha1.GenerateResponse], error) {
	return c.generate.CallUnary(ctx, req)
}

// GenerateServiceHandler is an implementation of the GenerateService service.
type GenerateServiceHandler interface {
	// Generate generates new cells given an existing document.
	Generate(context.Context, *connect.Request[v1alpha1.GenerateRequest]) (*connect.Response[v1alpha1.GenerateResponse], error)
}

// NewGenerateServiceHandler builds an HTTP handler from the service implementation. It returns the
// path on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewGenerateServiceHandler(svc GenerateServiceHandler, opts ...connect.HandlerOption) (string, http.Handler) {
	generateServiceGenerateHandler := connect.NewUnaryHandler(
		GenerateServiceGenerateProcedure,
		svc.Generate,
		connect.WithSchema(generateServiceGenerateMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	return "/GenerateService/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case GenerateServiceGenerateProcedure:
			generateServiceGenerateHandler.ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

// UnimplementedGenerateServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedGenerateServiceHandler struct{}

func (UnimplementedGenerateServiceHandler) Generate(context.Context, *connect.Request[v1alpha1.GenerateRequest]) (*connect.Response[v1alpha1.GenerateResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("GenerateService.Generate is not implemented"))
}

// ExecuteServiceClient is a client for the ExecuteService service.
type ExecuteServiceClient interface {
	// Execute executes a cell in an existing document.
	Execute(context.Context, *connect.Request[v1alpha1.ExecuteRequest]) (*connect.Response[v1alpha1.ExecuteResponse], error)
}

// NewExecuteServiceClient constructs a client for the ExecuteService service. By default, it uses
// the Connect protocol with the binary Protobuf Codec, asks for gzipped responses, and sends
// uncompressed requests. To use the gRPC or gRPC-Web protocols, supply the connect.WithGRPC() or
// connect.WithGRPCWeb() options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewExecuteServiceClient(httpClient connect.HTTPClient, baseURL string, opts ...connect.ClientOption) ExecuteServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	return &executeServiceClient{
		execute: connect.NewClient[v1alpha1.ExecuteRequest, v1alpha1.ExecuteResponse](
			httpClient,
			baseURL+ExecuteServiceExecuteProcedure,
			connect.WithSchema(executeServiceExecuteMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
	}
}

// executeServiceClient implements ExecuteServiceClient.
type executeServiceClient struct {
	execute *connect.Client[v1alpha1.ExecuteRequest, v1alpha1.ExecuteResponse]
}

// Execute calls ExecuteService.Execute.
func (c *executeServiceClient) Execute(ctx context.Context, req *connect.Request[v1alpha1.ExecuteRequest]) (*connect.Response[v1alpha1.ExecuteResponse], error) {
	return c.execute.CallUnary(ctx, req)
}

// ExecuteServiceHandler is an implementation of the ExecuteService service.
type ExecuteServiceHandler interface {
	// Execute executes a cell in an existing document.
	Execute(context.Context, *connect.Request[v1alpha1.ExecuteRequest]) (*connect.Response[v1alpha1.ExecuteResponse], error)
}

// NewExecuteServiceHandler builds an HTTP handler from the service implementation. It returns the
// path on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewExecuteServiceHandler(svc ExecuteServiceHandler, opts ...connect.HandlerOption) (string, http.Handler) {
	executeServiceExecuteHandler := connect.NewUnaryHandler(
		ExecuteServiceExecuteProcedure,
		svc.Execute,
		connect.WithSchema(executeServiceExecuteMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	return "/ExecuteService/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case ExecuteServiceExecuteProcedure:
			executeServiceExecuteHandler.ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

// UnimplementedExecuteServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedExecuteServiceHandler struct{}

func (UnimplementedExecuteServiceHandler) Execute(context.Context, *connect.Request[v1alpha1.ExecuteRequest]) (*connect.Response[v1alpha1.ExecuteResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("ExecuteService.Execute is not implemented"))
}

// AIServiceClient is a client for the AIService service.
type AIServiceClient interface {
	// StreamGenerate is a bidirectional streaming RPC for generating completions
	StreamGenerate(context.Context) *connect.BidiStreamForClient[v1alpha1.StreamGenerateRequest, v1alpha1.StreamGenerateResponse]
	// GenerateCells uses the AI to generate cells to insert into the notebook.
	GenerateCells(context.Context, *connect.Request[v1alpha1.GenerateCellsRequest]) (*connect.Response[v1alpha1.GenerateCellsResponse], error)
	// GetExample returns a learned example.
	// This is mostly for observability.
	GetExample(context.Context, *connect.Request[v1alpha1.GetExampleRequest]) (*connect.Response[v1alpha1.GetExampleResponse], error)
	// N.B. This is for testing only. Wanted to add a non streaming response which we can use to verify things are working.
	Status(context.Context, *connect.Request[v1alpha1.StatusRequest]) (*connect.Response[v1alpha1.StatusResponse], error)
}

// NewAIServiceClient constructs a client for the AIService service. By default, it uses the Connect
// protocol with the binary Protobuf Codec, asks for gzipped responses, and sends uncompressed
// requests. To use the gRPC or gRPC-Web protocols, supply the connect.WithGRPC() or
// connect.WithGRPCWeb() options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewAIServiceClient(httpClient connect.HTTPClient, baseURL string, opts ...connect.ClientOption) AIServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	return &aIServiceClient{
		streamGenerate: connect.NewClient[v1alpha1.StreamGenerateRequest, v1alpha1.StreamGenerateResponse](
			httpClient,
			baseURL+AIServiceStreamGenerateProcedure,
			connect.WithSchema(aIServiceStreamGenerateMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
		generateCells: connect.NewClient[v1alpha1.GenerateCellsRequest, v1alpha1.GenerateCellsResponse](
			httpClient,
			baseURL+AIServiceGenerateCellsProcedure,
			connect.WithSchema(aIServiceGenerateCellsMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
		getExample: connect.NewClient[v1alpha1.GetExampleRequest, v1alpha1.GetExampleResponse](
			httpClient,
			baseURL+AIServiceGetExampleProcedure,
			connect.WithSchema(aIServiceGetExampleMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
		status: connect.NewClient[v1alpha1.StatusRequest, v1alpha1.StatusResponse](
			httpClient,
			baseURL+AIServiceStatusProcedure,
			connect.WithSchema(aIServiceStatusMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
	}
}

// aIServiceClient implements AIServiceClient.
type aIServiceClient struct {
	streamGenerate *connect.Client[v1alpha1.StreamGenerateRequest, v1alpha1.StreamGenerateResponse]
	generateCells  *connect.Client[v1alpha1.GenerateCellsRequest, v1alpha1.GenerateCellsResponse]
	getExample     *connect.Client[v1alpha1.GetExampleRequest, v1alpha1.GetExampleResponse]
	status         *connect.Client[v1alpha1.StatusRequest, v1alpha1.StatusResponse]
}

// StreamGenerate calls AIService.StreamGenerate.
func (c *aIServiceClient) StreamGenerate(ctx context.Context) *connect.BidiStreamForClient[v1alpha1.StreamGenerateRequest, v1alpha1.StreamGenerateResponse] {
	return c.streamGenerate.CallBidiStream(ctx)
}

// GenerateCells calls AIService.GenerateCells.
func (c *aIServiceClient) GenerateCells(ctx context.Context, req *connect.Request[v1alpha1.GenerateCellsRequest]) (*connect.Response[v1alpha1.GenerateCellsResponse], error) {
	return c.generateCells.CallUnary(ctx, req)
}

// GetExample calls AIService.GetExample.
func (c *aIServiceClient) GetExample(ctx context.Context, req *connect.Request[v1alpha1.GetExampleRequest]) (*connect.Response[v1alpha1.GetExampleResponse], error) {
	return c.getExample.CallUnary(ctx, req)
}

// Status calls AIService.Status.
func (c *aIServiceClient) Status(ctx context.Context, req *connect.Request[v1alpha1.StatusRequest]) (*connect.Response[v1alpha1.StatusResponse], error) {
	return c.status.CallUnary(ctx, req)
}

// AIServiceHandler is an implementation of the AIService service.
type AIServiceHandler interface {
	// StreamGenerate is a bidirectional streaming RPC for generating completions
	StreamGenerate(context.Context, *connect.BidiStream[v1alpha1.StreamGenerateRequest, v1alpha1.StreamGenerateResponse]) error
	// GenerateCells uses the AI to generate cells to insert into the notebook.
	GenerateCells(context.Context, *connect.Request[v1alpha1.GenerateCellsRequest]) (*connect.Response[v1alpha1.GenerateCellsResponse], error)
	// GetExample returns a learned example.
	// This is mostly for observability.
	GetExample(context.Context, *connect.Request[v1alpha1.GetExampleRequest]) (*connect.Response[v1alpha1.GetExampleResponse], error)
	// N.B. This is for testing only. Wanted to add a non streaming response which we can use to verify things are working.
	Status(context.Context, *connect.Request[v1alpha1.StatusRequest]) (*connect.Response[v1alpha1.StatusResponse], error)
}

// NewAIServiceHandler builds an HTTP handler from the service implementation. It returns the path
// on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewAIServiceHandler(svc AIServiceHandler, opts ...connect.HandlerOption) (string, http.Handler) {
	aIServiceStreamGenerateHandler := connect.NewBidiStreamHandler(
		AIServiceStreamGenerateProcedure,
		svc.StreamGenerate,
		connect.WithSchema(aIServiceStreamGenerateMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	aIServiceGenerateCellsHandler := connect.NewUnaryHandler(
		AIServiceGenerateCellsProcedure,
		svc.GenerateCells,
		connect.WithSchema(aIServiceGenerateCellsMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	aIServiceGetExampleHandler := connect.NewUnaryHandler(
		AIServiceGetExampleProcedure,
		svc.GetExample,
		connect.WithSchema(aIServiceGetExampleMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	aIServiceStatusHandler := connect.NewUnaryHandler(
		AIServiceStatusProcedure,
		svc.Status,
		connect.WithSchema(aIServiceStatusMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	return "/AIService/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case AIServiceStreamGenerateProcedure:
			aIServiceStreamGenerateHandler.ServeHTTP(w, r)
		case AIServiceGenerateCellsProcedure:
			aIServiceGenerateCellsHandler.ServeHTTP(w, r)
		case AIServiceGetExampleProcedure:
			aIServiceGetExampleHandler.ServeHTTP(w, r)
		case AIServiceStatusProcedure:
			aIServiceStatusHandler.ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

// UnimplementedAIServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedAIServiceHandler struct{}

func (UnimplementedAIServiceHandler) StreamGenerate(context.Context, *connect.BidiStream[v1alpha1.StreamGenerateRequest, v1alpha1.StreamGenerateResponse]) error {
	return connect.NewError(connect.CodeUnimplemented, errors.New("AIService.StreamGenerate is not implemented"))
}

func (UnimplementedAIServiceHandler) GenerateCells(context.Context, *connect.Request[v1alpha1.GenerateCellsRequest]) (*connect.Response[v1alpha1.GenerateCellsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("AIService.GenerateCells is not implemented"))
}

func (UnimplementedAIServiceHandler) GetExample(context.Context, *connect.Request[v1alpha1.GetExampleRequest]) (*connect.Response[v1alpha1.GetExampleResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("AIService.GetExample is not implemented"))
}

func (UnimplementedAIServiceHandler) Status(context.Context, *connect.Request[v1alpha1.StatusRequest]) (*connect.Response[v1alpha1.StatusResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("AIService.Status is not implemented"))
}
