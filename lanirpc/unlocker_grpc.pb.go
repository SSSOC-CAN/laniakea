// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package lanirpc

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

// UnlockerClient is the client API for Unlocker service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type UnlockerClient interface {
	// lanicli: `login`
	//Login will prompt the user to provide a password and send the response to the unlocker service for authentication.
	Login(ctx context.Context, in *LoginRequest, opts ...grpc.CallOption) (*LoginResponse, error)
	// lanicli: `setpassword`
	//SetPassword prompts the user to set a password on first startup if no password has already been set.
	SetPassword(ctx context.Context, in *SetPwdRequest, opts ...grpc.CallOption) (*SetPwdResponse, error)
	// lanicli: `changepassword`
	//ChangePassword prompts the user to enter the current password and enter a new password.
	ChangePassword(ctx context.Context, in *ChangePwdRequest, opts ...grpc.CallOption) (*ChangePwdResponse, error)
}

type unlockerClient struct {
	cc grpc.ClientConnInterface
}

func NewUnlockerClient(cc grpc.ClientConnInterface) UnlockerClient {
	return &unlockerClient{cc}
}

func (c *unlockerClient) Login(ctx context.Context, in *LoginRequest, opts ...grpc.CallOption) (*LoginResponse, error) {
	out := new(LoginResponse)
	err := c.cc.Invoke(ctx, "/lanirpc.Unlocker/Login", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *unlockerClient) SetPassword(ctx context.Context, in *SetPwdRequest, opts ...grpc.CallOption) (*SetPwdResponse, error) {
	out := new(SetPwdResponse)
	err := c.cc.Invoke(ctx, "/lanirpc.Unlocker/SetPassword", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *unlockerClient) ChangePassword(ctx context.Context, in *ChangePwdRequest, opts ...grpc.CallOption) (*ChangePwdResponse, error) {
	out := new(ChangePwdResponse)
	err := c.cc.Invoke(ctx, "/lanirpc.Unlocker/ChangePassword", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// UnlockerServer is the server API for Unlocker service.
// All implementations must embed UnimplementedUnlockerServer
// for forward compatibility
type UnlockerServer interface {
	// lanicli: `login`
	//Login will prompt the user to provide a password and send the response to the unlocker service for authentication.
	Login(context.Context, *LoginRequest) (*LoginResponse, error)
	// lanicli: `setpassword`
	//SetPassword prompts the user to set a password on first startup if no password has already been set.
	SetPassword(context.Context, *SetPwdRequest) (*SetPwdResponse, error)
	// lanicli: `changepassword`
	//ChangePassword prompts the user to enter the current password and enter a new password.
	ChangePassword(context.Context, *ChangePwdRequest) (*ChangePwdResponse, error)
	mustEmbedUnimplementedUnlockerServer()
}

// UnimplementedUnlockerServer must be embedded to have forward compatible implementations.
type UnimplementedUnlockerServer struct {
}

func (UnimplementedUnlockerServer) Login(context.Context, *LoginRequest) (*LoginResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Login not implemented")
}
func (UnimplementedUnlockerServer) SetPassword(context.Context, *SetPwdRequest) (*SetPwdResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetPassword not implemented")
}
func (UnimplementedUnlockerServer) ChangePassword(context.Context, *ChangePwdRequest) (*ChangePwdResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ChangePassword not implemented")
}
func (UnimplementedUnlockerServer) mustEmbedUnimplementedUnlockerServer() {}

// UnsafeUnlockerServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to UnlockerServer will
// result in compilation errors.
type UnsafeUnlockerServer interface {
	mustEmbedUnimplementedUnlockerServer()
}

func RegisterUnlockerServer(s grpc.ServiceRegistrar, srv UnlockerServer) {
	s.RegisterService(&Unlocker_ServiceDesc, srv)
}

func _Unlocker_Login_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LoginRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UnlockerServer).Login(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/lanirpc.Unlocker/Login",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UnlockerServer).Login(ctx, req.(*LoginRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Unlocker_SetPassword_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SetPwdRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UnlockerServer).SetPassword(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/lanirpc.Unlocker/SetPassword",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UnlockerServer).SetPassword(ctx, req.(*SetPwdRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Unlocker_ChangePassword_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ChangePwdRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UnlockerServer).ChangePassword(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/lanirpc.Unlocker/ChangePassword",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UnlockerServer).ChangePassword(ctx, req.(*ChangePwdRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Unlocker_ServiceDesc is the grpc.ServiceDesc for Unlocker service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Unlocker_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "lanirpc.Unlocker",
	HandlerType: (*UnlockerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Login",
			Handler:    _Unlocker_Login_Handler,
		},
		{
			MethodName: "SetPassword",
			Handler:    _Unlocker_SetPassword_Handler,
		},
		{
			MethodName: "ChangePassword",
			Handler:    _Unlocker_ChangePassword_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "unlocker.proto",
}