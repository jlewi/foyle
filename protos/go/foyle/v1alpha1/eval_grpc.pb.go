// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             (unknown)
// source: foyle/v1alpha1/eval.proto

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

// EvalServiceClient is the client API for EvalService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type EvalServiceClient interface {
	List(ctx context.Context, in *EvalResultListRequest, opts ...grpc.CallOption) (*EvalResultListResponse, error)
	AssertionTable(ctx context.Context, in *AssertionTableRequest, opts ...grpc.CallOption) (*AssertionTableResponse, error)
}

type evalServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewEvalServiceClient(cc grpc.ClientConnInterface) EvalServiceClient {
	return &evalServiceClient{cc}
}

func (c *evalServiceClient) List(ctx context.Context, in *EvalResultListRequest, opts ...grpc.CallOption) (*EvalResultListResponse, error) {
	out := new(EvalResultListResponse)
	err := c.cc.Invoke(ctx, "/EvalService/List", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *evalServiceClient) AssertionTable(ctx context.Context, in *AssertionTableRequest, opts ...grpc.CallOption) (*AssertionTableResponse, error) {
	out := new(AssertionTableResponse)
	err := c.cc.Invoke(ctx, "/EvalService/AssertionTable", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// EvalServiceServer is the server API for EvalService service.
// All implementations must embed UnimplementedEvalServiceServer
// for forward compatibility
type EvalServiceServer interface {
	List(context.Context, *EvalResultListRequest) (*EvalResultListResponse, error)
	AssertionTable(context.Context, *AssertionTableRequest) (*AssertionTableResponse, error)
	mustEmbedUnimplementedEvalServiceServer()
}

// UnimplementedEvalServiceServer must be embedded to have forward compatible implementations.
type UnimplementedEvalServiceServer struct {
}

func (UnimplementedEvalServiceServer) List(context.Context, *EvalResultListRequest) (*EvalResultListResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method List not implemented")
}
func (UnimplementedEvalServiceServer) AssertionTable(context.Context, *AssertionTableRequest) (*AssertionTableResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AssertionTable not implemented")
}
func (UnimplementedEvalServiceServer) mustEmbedUnimplementedEvalServiceServer() {}

// UnsafeEvalServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to EvalServiceServer will
// result in compilation errors.
type UnsafeEvalServiceServer interface {
	mustEmbedUnimplementedEvalServiceServer()
}

func RegisterEvalServiceServer(s grpc.ServiceRegistrar, srv EvalServiceServer) {
	s.RegisterService(&EvalService_ServiceDesc, srv)
}

func _EvalService_List_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EvalResultListRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EvalServiceServer).List(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/EvalService/List",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(EvalServiceServer).List(ctx, req.(*EvalResultListRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _EvalService_AssertionTable_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AssertionTableRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EvalServiceServer).AssertionTable(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/EvalService/AssertionTable",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(EvalServiceServer).AssertionTable(ctx, req.(*AssertionTableRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// EvalService_ServiceDesc is the grpc.ServiceDesc for EvalService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var EvalService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "EvalService",
	HandlerType: (*EvalServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "List",
			Handler:    _EvalService_List_Handler,
		},
		{
			MethodName: "AssertionTable",
			Handler:    _EvalService_AssertionTable_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "foyle/v1alpha1/eval.proto",
}
