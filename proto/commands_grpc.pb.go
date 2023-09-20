// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.6.1
// source: commands.proto

package proto

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

// CommandsClient is the client API for Commands service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type CommandsClient interface {
	Command(ctx context.Context, in *CommandRequest, opts ...grpc.CallOption) (*CommandResponse, error)
}

type commandsClient struct {
	cc grpc.ClientConnInterface
}

func NewCommandsClient(cc grpc.ClientConnInterface) CommandsClient {
	return &commandsClient{cc}
}

func (c *commandsClient) Command(ctx context.Context, in *CommandRequest, opts ...grpc.CallOption) (*CommandResponse, error) {
	out := new(CommandResponse)
	err := c.cc.Invoke(ctx, "/borealis.v1beta1.Commands/Command", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// CommandsServer is the server API for Commands service.
// All implementations must embed UnimplementedCommandsServer
// for forward compatibility
type CommandsServer interface {
	Command(context.Context, *CommandRequest) (*CommandResponse, error)
	mustEmbedUnimplementedCommandsServer()
}

// UnimplementedCommandsServer must be embedded to have forward compatible implementations.
type UnimplementedCommandsServer struct {
}

func (UnimplementedCommandsServer) Command(context.Context, *CommandRequest) (*CommandResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Command not implemented")
}
func (UnimplementedCommandsServer) mustEmbedUnimplementedCommandsServer() {}

// UnsafeCommandsServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to CommandsServer will
// result in compilation errors.
type UnsafeCommandsServer interface {
	mustEmbedUnimplementedCommandsServer()
}

func RegisterCommandsServer(s grpc.ServiceRegistrar, srv CommandsServer) {
	s.RegisterService(&Commands_ServiceDesc, srv)
}

func _Commands_Command_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CommandRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CommandsServer).Command(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/borealis.v1beta1.Commands/Command",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CommandsServer).Command(ctx, req.(*CommandRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Commands_ServiceDesc is the grpc.ServiceDesc for Commands service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Commands_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "borealis.v1beta1.Commands",
	HandlerType: (*CommandsServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Command",
			Handler:    _Commands_Command_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "commands.proto",
}