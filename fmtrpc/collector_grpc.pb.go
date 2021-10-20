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

// DataCollectorClient is the client API for DataCollector service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type DataCollectorClient interface {
	// fmtcli: `start-record`
	//StartRecording will begin recording data from specified service and writing it.
	StartRecording(ctx context.Context, in *RecordRequest, opts ...grpc.CallOption) (*RecordResponse, error)
	// fmtcli: `stop-record`
	//StopRecording will end the recording of data from specified service.
	StopRecording(ctx context.Context, in *StopRecRequest, opts ...grpc.CallOption) (*StopRecResponse, error)
	// fmtcli: `subscribe-data-stream`
	//SubscribeDataStream returns a uni-directional stream (server -> client) of data being recorded from the Fluke DAQ.
	SubscribeDataStream(ctx context.Context, in *SubscribeDataRequest, opts ...grpc.CallOption) (DataCollector_SubscribeDataStreamClient, error)
}

type dataCollectorClient struct {
	cc grpc.ClientConnInterface
}

func NewDataCollectorClient(cc grpc.ClientConnInterface) DataCollectorClient {
	return &dataCollectorClient{cc}
}

func (c *dataCollectorClient) StartRecording(ctx context.Context, in *RecordRequest, opts ...grpc.CallOption) (*RecordResponse, error) {
	out := new(RecordResponse)
	err := c.cc.Invoke(ctx, "/fmtrpc.DataCollector/StartRecording", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *dataCollectorClient) StopRecording(ctx context.Context, in *StopRecRequest, opts ...grpc.CallOption) (*StopRecResponse, error) {
	out := new(StopRecResponse)
	err := c.cc.Invoke(ctx, "/fmtrpc.DataCollector/StopRecording", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *dataCollectorClient) SubscribeDataStream(ctx context.Context, in *SubscribeDataRequest, opts ...grpc.CallOption) (DataCollector_SubscribeDataStreamClient, error) {
	stream, err := c.cc.NewStream(ctx, &DataCollector_ServiceDesc.Streams[0], "/fmtrpc.DataCollector/SubscribeDataStream", opts...)
	if err != nil {
		return nil, err
	}
	x := &dataCollectorSubscribeDataStreamClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type DataCollector_SubscribeDataStreamClient interface {
	Recv() (*RealTimeData, error)
	grpc.ClientStream
}

type dataCollectorSubscribeDataStreamClient struct {
	grpc.ClientStream
}

func (x *dataCollectorSubscribeDataStreamClient) Recv() (*RealTimeData, error) {
	m := new(RealTimeData)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// DataCollectorServer is the server API for DataCollector service.
// All implementations must embed UnimplementedDataCollectorServer
// for forward compatibility
type DataCollectorServer interface {
	// fmtcli: `start-record`
	//StartRecording will begin recording data from specified service and writing it.
	StartRecording(context.Context, *RecordRequest) (*RecordResponse, error)
	// fmtcli: `stop-record`
	//StopRecording will end the recording of data from specified service.
	StopRecording(context.Context, *StopRecRequest) (*StopRecResponse, error)
	// fmtcli: `subscribe-data-stream`
	//SubscribeDataStream returns a uni-directional stream (server -> client) of data being recorded from the Fluke DAQ.
	SubscribeDataStream(*SubscribeDataRequest, DataCollector_SubscribeDataStreamServer) error
	mustEmbedUnimplementedDataCollectorServer()
}

// UnimplementedDataCollectorServer must be embedded to have forward compatible implementations.
type UnimplementedDataCollectorServer struct {
}

func (UnimplementedDataCollectorServer) StartRecording(context.Context, *RecordRequest) (*RecordResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method StartRecording not implemented")
}
func (UnimplementedDataCollectorServer) StopRecording(context.Context, *StopRecRequest) (*StopRecResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method StopRecording not implemented")
}
func (UnimplementedDataCollectorServer) SubscribeDataStream(*SubscribeDataRequest, DataCollector_SubscribeDataStreamServer) error {
	return status.Errorf(codes.Unimplemented, "method SubscribeDataStream not implemented")
}
func (UnimplementedDataCollectorServer) mustEmbedUnimplementedDataCollectorServer() {}

// UnsafeDataCollectorServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to DataCollectorServer will
// result in compilation errors.
type UnsafeDataCollectorServer interface {
	mustEmbedUnimplementedDataCollectorServer()
}

func RegisterDataCollectorServer(s grpc.ServiceRegistrar, srv DataCollectorServer) {
	s.RegisterService(&DataCollector_ServiceDesc, srv)
}

func _DataCollector_StartRecording_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RecordRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DataCollectorServer).StartRecording(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/fmtrpc.DataCollector/StartRecording",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DataCollectorServer).StartRecording(ctx, req.(*RecordRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _DataCollector_StopRecording_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StopRecRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DataCollectorServer).StopRecording(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/fmtrpc.DataCollector/StopRecording",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DataCollectorServer).StopRecording(ctx, req.(*StopRecRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _DataCollector_SubscribeDataStream_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(SubscribeDataRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(DataCollectorServer).SubscribeDataStream(m, &dataCollectorSubscribeDataStreamServer{stream})
}

type DataCollector_SubscribeDataStreamServer interface {
	Send(*RealTimeData) error
	grpc.ServerStream
}

type dataCollectorSubscribeDataStreamServer struct {
	grpc.ServerStream
}

func (x *dataCollectorSubscribeDataStreamServer) Send(m *RealTimeData) error {
	return x.ServerStream.SendMsg(m)
}

// DataCollector_ServiceDesc is the grpc.ServiceDesc for DataCollector service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var DataCollector_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "fmtrpc.DataCollector",
	HandlerType: (*DataCollectorServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "StartRecording",
			Handler:    _DataCollector_StartRecording_Handler,
		},
		{
			MethodName: "StopRecording",
			Handler:    _DataCollector_StopRecording_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "SubscribeDataStream",
			Handler:       _DataCollector_SubscribeDataStream_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "collector.proto",
}
