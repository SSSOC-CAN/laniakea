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

// TestPlanExecutorClient is the client API for TestPlanExecutor service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type TestPlanExecutorClient interface {
	// fmtcli: `load-test`
	//LoadTestPlan will load the inputted test plan and ensure the FMT can currently run it.
	LoadTestPlan(ctx context.Context, in *LoadTestPlanRequest, opts ...grpc.CallOption) (*LoadTestPlanResponse, error)
	// fmtcli: `start-test`
	//StartTestPlan will start the loaded test plan. This will automatically start recording from services as need be.
	StartTestPlan(ctx context.Context, in *StartTestPlanRequest, opts ...grpc.CallOption) (*StartTestPlanResponse, error)
	// fmtcli: `stop-test`
	//StopTestPlan will stop the currently running test plan. This will automatically stop all data recording.
	StopTestPlan(ctx context.Context, in *StopTestPlanRequest, opts ...grpc.CallOption) (*StopTestPlanResponse, error)
	// fmtcli: `insert-roi`
	//InsertROIMarker will insert a Region of Interest (ROI) marker within the test report based on given parameters.
	InsertROIMarker(ctx context.Context, in *InsertROIRequest, opts ...grpc.CallOption) (*InsertROIResponse, error)
}

type testPlanExecutorClient struct {
	cc grpc.ClientConnInterface
}

func NewTestPlanExecutorClient(cc grpc.ClientConnInterface) TestPlanExecutorClient {
	return &testPlanExecutorClient{cc}
}

func (c *testPlanExecutorClient) LoadTestPlan(ctx context.Context, in *LoadTestPlanRequest, opts ...grpc.CallOption) (*LoadTestPlanResponse, error) {
	out := new(LoadTestPlanResponse)
	err := c.cc.Invoke(ctx, "/fmtrpc.TestPlanExecutor/LoadTestPlan", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *testPlanExecutorClient) StartTestPlan(ctx context.Context, in *StartTestPlanRequest, opts ...grpc.CallOption) (*StartTestPlanResponse, error) {
	out := new(StartTestPlanResponse)
	err := c.cc.Invoke(ctx, "/fmtrpc.TestPlanExecutor/StartTestPlan", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *testPlanExecutorClient) StopTestPlan(ctx context.Context, in *StopTestPlanRequest, opts ...grpc.CallOption) (*StopTestPlanResponse, error) {
	out := new(StopTestPlanResponse)
	err := c.cc.Invoke(ctx, "/fmtrpc.TestPlanExecutor/StopTestPlan", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *testPlanExecutorClient) InsertROIMarker(ctx context.Context, in *InsertROIRequest, opts ...grpc.CallOption) (*InsertROIResponse, error) {
	out := new(InsertROIResponse)
	err := c.cc.Invoke(ctx, "/fmtrpc.TestPlanExecutor/InsertROIMarker", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// TestPlanExecutorServer is the server API for TestPlanExecutor service.
// All implementations must embed UnimplementedTestPlanExecutorServer
// for forward compatibility
type TestPlanExecutorServer interface {
	// fmtcli: `load-test`
	//LoadTestPlan will load the inputted test plan and ensure the FMT can currently run it.
	LoadTestPlan(context.Context, *LoadTestPlanRequest) (*LoadTestPlanResponse, error)
	// fmtcli: `start-test`
	//StartTestPlan will start the loaded test plan. This will automatically start recording from services as need be.
	StartTestPlan(context.Context, *StartTestPlanRequest) (*StartTestPlanResponse, error)
	// fmtcli: `stop-test`
	//StopTestPlan will stop the currently running test plan. This will automatically stop all data recording.
	StopTestPlan(context.Context, *StopTestPlanRequest) (*StopTestPlanResponse, error)
	// fmtcli: `insert-roi`
	//InsertROIMarker will insert a Region of Interest (ROI) marker within the test report based on given parameters.
	InsertROIMarker(context.Context, *InsertROIRequest) (*InsertROIResponse, error)
	mustEmbedUnimplementedTestPlanExecutorServer()
}

// UnimplementedTestPlanExecutorServer must be embedded to have forward compatible implementations.
type UnimplementedTestPlanExecutorServer struct {
}

func (UnimplementedTestPlanExecutorServer) LoadTestPlan(context.Context, *LoadTestPlanRequest) (*LoadTestPlanResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method LoadTestPlan not implemented")
}
func (UnimplementedTestPlanExecutorServer) StartTestPlan(context.Context, *StartTestPlanRequest) (*StartTestPlanResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method StartTestPlan not implemented")
}
func (UnimplementedTestPlanExecutorServer) StopTestPlan(context.Context, *StopTestPlanRequest) (*StopTestPlanResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method StopTestPlan not implemented")
}
func (UnimplementedTestPlanExecutorServer) InsertROIMarker(context.Context, *InsertROIRequest) (*InsertROIResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method InsertROIMarker not implemented")
}
func (UnimplementedTestPlanExecutorServer) mustEmbedUnimplementedTestPlanExecutorServer() {}

// UnsafeTestPlanExecutorServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to TestPlanExecutorServer will
// result in compilation errors.
type UnsafeTestPlanExecutorServer interface {
	mustEmbedUnimplementedTestPlanExecutorServer()
}

func RegisterTestPlanExecutorServer(s grpc.ServiceRegistrar, srv TestPlanExecutorServer) {
	s.RegisterService(&TestPlanExecutor_ServiceDesc, srv)
}

func _TestPlanExecutor_LoadTestPlan_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LoadTestPlanRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TestPlanExecutorServer).LoadTestPlan(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/fmtrpc.TestPlanExecutor/LoadTestPlan",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TestPlanExecutorServer).LoadTestPlan(ctx, req.(*LoadTestPlanRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TestPlanExecutor_StartTestPlan_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StartTestPlanRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TestPlanExecutorServer).StartTestPlan(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/fmtrpc.TestPlanExecutor/StartTestPlan",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TestPlanExecutorServer).StartTestPlan(ctx, req.(*StartTestPlanRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TestPlanExecutor_StopTestPlan_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StopTestPlanRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TestPlanExecutorServer).StopTestPlan(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/fmtrpc.TestPlanExecutor/StopTestPlan",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TestPlanExecutorServer).StopTestPlan(ctx, req.(*StopTestPlanRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TestPlanExecutor_InsertROIMarker_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(InsertROIRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TestPlanExecutorServer).InsertROIMarker(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/fmtrpc.TestPlanExecutor/InsertROIMarker",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TestPlanExecutorServer).InsertROIMarker(ctx, req.(*InsertROIRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// TestPlanExecutor_ServiceDesc is the grpc.ServiceDesc for TestPlanExecutor service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var TestPlanExecutor_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "fmtrpc.TestPlanExecutor",
	HandlerType: (*TestPlanExecutorServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "LoadTestPlan",
			Handler:    _TestPlanExecutor_LoadTestPlan_Handler,
		},
		{
			MethodName: "StartTestPlan",
			Handler:    _TestPlanExecutor_StartTestPlan_Handler,
		},
		{
			MethodName: "StopTestPlan",
			Handler:    _TestPlanExecutor_StopTestPlan_Handler,
		},
		{
			MethodName: "InsertROIMarker",
			Handler:    _TestPlanExecutor_InsertROIMarker_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "executor.proto",
}
