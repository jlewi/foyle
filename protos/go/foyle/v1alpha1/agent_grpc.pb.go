// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             (unknown)
// source: foyle/v1alpha1/agent.proto

package v1alpha1

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// GenerateServiceClient is the client API for GenerateService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type GenerateServiceClient interface {
	// Generate generates new cells given an existing document.
	Generate(ctx context.Context, in *GenerateRequest, opts ...grpc.CallOption) (*GenerateResponse, error)
}

type generateServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewGenerateServiceClient(cc grpc.ClientConnInterface) GenerateServiceClient {
	return &generateServiceClient{cc}
}

func (c *generateServiceClient) Generate(ctx context.Context, in *GenerateRequest, opts ...grpc.CallOption) (*GenerateResponse, error) {
	out := new(GenerateResponse)
	err := c.cc.Invoke(ctx, "/GenerateService/Generate", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GenerateServiceServer is the server API for GenerateService service.
// All implementations must embed UnimplementedGenerateServiceServer
// for forward compatibility
type GenerateServiceServer interface {
	// Generate generates new cells given an existing document.
	Generate(context.Context, *GenerateRequest) (*GenerateResponse, error)
	mustEmbedUnimplementedGenerateServiceServer()
}

// UnimplementedGenerateServiceServer must be embedded to have forward compatible implementations.
type UnimplementedGenerateServiceServer struct {
}

func (UnimplementedGenerateServiceServer) Generate(context.Context, *GenerateRequest) (*GenerateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Generate not implemented")
}
func (UnimplementedGenerateServiceServer) mustEmbedUnimplementedGenerateServiceServer() {}

// UnsafeGenerateServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to GenerateServiceServer will
// result in compilation errors.
type UnsafeGenerateServiceServer interface {
	mustEmbedUnimplementedGenerateServiceServer()
}

func RegisterGenerateServiceServer(s grpc.ServiceRegistrar, srv GenerateServiceServer) {
	s.RegisterService(&GenerateService_ServiceDesc, srv)
}

func _GenerateService_Generate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GenerateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GenerateServiceServer).Generate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/GenerateService/Generate",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GenerateServiceServer).Generate(ctx, req.(*GenerateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// GenerateService_ServiceDesc is the grpc.ServiceDesc for GenerateService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var GenerateService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "GenerateService",
	HandlerType: (*GenerateServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Generate",
			Handler:    _GenerateService_Generate_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "foyle/v1alpha1/agent.proto",
}

// ExecuteServiceClient is the client API for ExecuteService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ExecuteServiceClient interface {
	// Execute executes a cell in an existing document.
	Execute(ctx context.Context, in *ExecuteRequest, opts ...grpc.CallOption) (*ExecuteResponse, error)
}

type executeServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewExecuteServiceClient(cc grpc.ClientConnInterface) ExecuteServiceClient {
	return &executeServiceClient{cc}
}

func (c *executeServiceClient) Execute(ctx context.Context, in *ExecuteRequest, opts ...grpc.CallOption) (*ExecuteResponse, error) {
	out := new(ExecuteResponse)
	err := c.cc.Invoke(ctx, "/ExecuteService/Execute", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ExecuteServiceServer is the server API for ExecuteService service.
// All implementations must embed UnimplementedExecuteServiceServer
// for forward compatibility
type ExecuteServiceServer interface {
	// Execute executes a cell in an existing document.
	Execute(context.Context, *ExecuteRequest) (*ExecuteResponse, error)
	mustEmbedUnimplementedExecuteServiceServer()
}

// UnimplementedExecuteServiceServer must be embedded to have forward compatible implementations.
type UnimplementedExecuteServiceServer struct {
}

func (UnimplementedExecuteServiceServer) Execute(context.Context, *ExecuteRequest) (*ExecuteResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Execute not implemented")
}
func (UnimplementedExecuteServiceServer) mustEmbedUnimplementedExecuteServiceServer() {}

// UnsafeExecuteServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ExecuteServiceServer will
// result in compilation errors.
type UnsafeExecuteServiceServer interface {
	mustEmbedUnimplementedExecuteServiceServer()
}

func RegisterExecuteServiceServer(s grpc.ServiceRegistrar, srv ExecuteServiceServer) {
	s.RegisterService(&ExecuteService_ServiceDesc, srv)
}

func _ExecuteService_Execute_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ExecuteRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ExecuteServiceServer).Execute(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/ExecuteService/Execute",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ExecuteServiceServer).Execute(ctx, req.(*ExecuteRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// ExecuteService_ServiceDesc is the grpc.ServiceDesc for ExecuteService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ExecuteService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "ExecuteService",
	HandlerType: (*ExecuteServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Execute",
			Handler:    _ExecuteService_Execute_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "foyle/v1alpha1/agent.proto",
}
