// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package fmtrpc

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

// FmtClient is the client API for Fmt service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type FmtClient interface {
	// fmtcli: `stop`
	//StopDaemon will send a shutdown request to the interrupt handler, triggering
	//a graceful shutdown of the daemon.
	StopDaemon(ctx context.Context, in *StopRequest, opts ...grpc.CallOption) (*StopResponse, error)
	// fmtcli: `admin-test`
	//AdminTest will send a string response if the proper macaroon is provided.
	AdminTest(ctx context.Context, in *AdminTestRequest, opts ...grpc.CallOption) (*AdminTestResponse, error)
	// fmtcli: `test`
	//TestCommand will send a string response regardless if a macaroon is provided or not.
	TestCommand(ctx context.Context, in *TestRequest, opts ...grpc.CallOption) (*TestResponse, error)
	// fmtcli: `bake-macaroon`
	//BakeMacaroon will bake a new macaroon based on input permissions and constraints.
	BakeMacaroon(ctx context.Context, in *BakeMacaroonRequest, opts ...grpc.CallOption) (*BakeMacaroonResponse, error)
}

type fmtClient struct {
	cc grpc.ClientConnInterface
}

func NewFmtClient(cc grpc.ClientConnInterface) FmtClient {
	return &fmtClient{cc}
}

func (c *fmtClient) StopDaemon(ctx context.Context, in *StopRequest, opts ...grpc.CallOption) (*StopResponse, error) {
	out := new(StopResponse)
	err := c.cc.Invoke(ctx, "/fmtrpc.Fmt/StopDaemon", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fmtClient) AdminTest(ctx context.Context, in *AdminTestRequest, opts ...grpc.CallOption) (*AdminTestResponse, error) {
	out := new(AdminTestResponse)
	err := c.cc.Invoke(ctx, "/fmtrpc.Fmt/AdminTest", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fmtClient) TestCommand(ctx context.Context, in *TestRequest, opts ...grpc.CallOption) (*TestResponse, error) {
	out := new(TestResponse)
	err := c.cc.Invoke(ctx, "/fmtrpc.Fmt/TestCommand", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fmtClient) BakeMacaroon(ctx context.Context, in *BakeMacaroonRequest, opts ...grpc.CallOption) (*BakeMacaroonResponse, error) {
	out := new(BakeMacaroonResponse)
	err := c.cc.Invoke(ctx, "/fmtrpc.Fmt/BakeMacaroon", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// FmtServer is the server API for Fmt service.
// All implementations must embed UnimplementedFmtServer
// for forward compatibility
type FmtServer interface {
	// fmtcli: `stop`
	//StopDaemon will send a shutdown request to the interrupt handler, triggering
	//a graceful shutdown of the daemon.
	StopDaemon(context.Context, *StopRequest) (*StopResponse, error)
	// fmtcli: `admin-test`
	//AdminTest will send a string response if the proper macaroon is provided.
	AdminTest(context.Context, *AdminTestRequest) (*AdminTestResponse, error)
	// fmtcli: `test`
	//TestCommand will send a string response regardless if a macaroon is provided or not.
	TestCommand(context.Context, *TestRequest) (*TestResponse, error)
	// fmtcli: `bake-macaroon`
	//BakeMacaroon will bake a new macaroon based on input permissions and constraints.
	BakeMacaroon(context.Context, *BakeMacaroonRequest) (*BakeMacaroonResponse, error)
	mustEmbedUnimplementedFmtServer()
}

// UnimplementedFmtServer must be embedded to have forward compatible implementations.
type UnimplementedFmtServer struct {
}

func (UnimplementedFmtServer) StopDaemon(context.Context, *StopRequest) (*StopResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method StopDaemon not implemented")
}
func (UnimplementedFmtServer) AdminTest(context.Context, *AdminTestRequest) (*AdminTestResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AdminTest not implemented")
}
func (UnimplementedFmtServer) TestCommand(context.Context, *TestRequest) (*TestResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method TestCommand not implemented")
}
func (UnimplementedFmtServer) BakeMacaroon(context.Context, *BakeMacaroonRequest) (*BakeMacaroonResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method BakeMacaroon not implemented")
}
func (UnimplementedFmtServer) mustEmbedUnimplementedFmtServer() {}

// UnsafeFmtServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to FmtServer will
// result in compilation errors.
type UnsafeFmtServer interface {
	mustEmbedUnimplementedFmtServer()
}

func RegisterFmtServer(s grpc.ServiceRegistrar, srv FmtServer) {
	s.RegisterService(&Fmt_ServiceDesc, srv)
}

func _Fmt_StopDaemon_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StopRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FmtServer).StopDaemon(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/fmtrpc.Fmt/StopDaemon",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FmtServer).StopDaemon(ctx, req.(*StopRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Fmt_AdminTest_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AdminTestRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FmtServer).AdminTest(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/fmtrpc.Fmt/AdminTest",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FmtServer).AdminTest(ctx, req.(*AdminTestRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Fmt_TestCommand_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TestRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FmtServer).TestCommand(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/fmtrpc.Fmt/TestCommand",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FmtServer).TestCommand(ctx, req.(*TestRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Fmt_BakeMacaroon_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BakeMacaroonRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FmtServer).BakeMacaroon(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/fmtrpc.Fmt/BakeMacaroon",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FmtServer).BakeMacaroon(ctx, req.(*BakeMacaroonRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Fmt_ServiceDesc is the grpc.ServiceDesc for Fmt service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Fmt_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "fmtrpc.Fmt",
	HandlerType: (*FmtServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "StopDaemon",
			Handler:    _Fmt_StopDaemon_Handler,
		},
		{
			MethodName: "AdminTest",
			Handler:    _Fmt_AdminTest_Handler,
		},
		{
			MethodName: "TestCommand",
			Handler:    _Fmt_TestCommand_Handler,
		},
		{
			MethodName: "BakeMacaroon",
			Handler:    _Fmt_BakeMacaroon_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "fmt.proto",
}
