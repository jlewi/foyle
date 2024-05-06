// Code generated by protoc-gen-connect-go. DO NOT EDIT.
//
// Source: foyle/v1alpha1/eval.proto

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
	// EvalServiceName is the fully-qualified name of the EvalService service.
	EvalServiceName = "EvalService"
)

// These constants are the fully-qualified names of the RPCs defined in this package. They're
// exposed at runtime as Spec.Procedure and as the final two segments of the HTTP route.
//
// Note that these are different from the fully-qualified method names used by
// google.golang.org/protobuf/reflect/protoreflect. To convert from these constants to
// reflection-formatted method names, remove the leading slash and convert the remaining slash to a
// period.
const (
	// EvalServiceListProcedure is the fully-qualified name of the EvalService's List RPC.
	EvalServiceListProcedure = "/EvalService/List"
)

// These variables are the protoreflect.Descriptor objects for the RPCs defined in this package.
var (
	evalServiceServiceDescriptor    = v1alpha1.File_foyle_v1alpha1_eval_proto.Services().ByName("EvalService")
	evalServiceListMethodDescriptor = evalServiceServiceDescriptor.Methods().ByName("List")
)

// EvalServiceClient is a client for the EvalService service.
type EvalServiceClient interface {
	List(context.Context, *connect.Request[v1alpha1.EvalResultListRequest]) (*connect.Response[v1alpha1.EvalResultListResponse], error)
}

// NewEvalServiceClient constructs a client for the EvalService service. By default, it uses the
// Connect protocol with the binary Protobuf Codec, asks for gzipped responses, and sends
// uncompressed requests. To use the gRPC or gRPC-Web protocols, supply the connect.WithGRPC() or
// connect.WithGRPCWeb() options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewEvalServiceClient(httpClient connect.HTTPClient, baseURL string, opts ...connect.ClientOption) EvalServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	return &evalServiceClient{
		list: connect.NewClient[v1alpha1.EvalResultListRequest, v1alpha1.EvalResultListResponse](
			httpClient,
			baseURL+EvalServiceListProcedure,
			connect.WithSchema(evalServiceListMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
	}
}

// evalServiceClient implements EvalServiceClient.
type evalServiceClient struct {
	list *connect.Client[v1alpha1.EvalResultListRequest, v1alpha1.EvalResultListResponse]
}

// List calls EvalService.List.
func (c *evalServiceClient) List(ctx context.Context, req *connect.Request[v1alpha1.EvalResultListRequest]) (*connect.Response[v1alpha1.EvalResultListResponse], error) {
	return c.list.CallUnary(ctx, req)
}

// EvalServiceHandler is an implementation of the EvalService service.
type EvalServiceHandler interface {
	List(context.Context, *connect.Request[v1alpha1.EvalResultListRequest]) (*connect.Response[v1alpha1.EvalResultListResponse], error)
}

// NewEvalServiceHandler builds an HTTP handler from the service implementation. It returns the path
// on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewEvalServiceHandler(svc EvalServiceHandler, opts ...connect.HandlerOption) (string, http.Handler) {
	evalServiceListHandler := connect.NewUnaryHandler(
		EvalServiceListProcedure,
		svc.List,
		connect.WithSchema(evalServiceListMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	return "/EvalService/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case EvalServiceListProcedure:
			evalServiceListHandler.ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

// UnimplementedEvalServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedEvalServiceHandler struct{}

func (UnimplementedEvalServiceHandler) List(context.Context, *connect.Request[v1alpha1.EvalResultListRequest]) (*connect.Response[v1alpha1.EvalResultListResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("EvalService.List is not implemented"))
}
