// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.19.4
// source: demorpc/controller.proto

package demorpc

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type SetTempRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The desired temperature set point in degrees Celsius
	TempSetPoint float64 `protobuf:"fixed64,1,opt,name=temp_set_point,json=tempSetPoint,proto3" json:"temp_set_point,omitempty"`
	// The desired temperature change rate in degrees Celsius per minute
	TempChangeRate float64 `protobuf:"fixed64,2,opt,name=temp_change_rate,json=tempChangeRate,proto3" json:"temp_change_rate,omitempty"`
}

func (x *SetTempRequest) Reset() {
	*x = SetTempRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_demorpc_controller_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SetTempRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SetTempRequest) ProtoMessage() {}

func (x *SetTempRequest) ProtoReflect() protoreflect.Message {
	mi := &file_demorpc_controller_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SetTempRequest.ProtoReflect.Descriptor instead.
func (*SetTempRequest) Descriptor() ([]byte, []int) {
	return file_demorpc_controller_proto_rawDescGZIP(), []int{0}
}

func (x *SetTempRequest) GetTempSetPoint() float64 {
	if x != nil {
		return x.TempSetPoint
	}
	return 0
}

func (x *SetTempRequest) GetTempChangeRate() float64 {
	if x != nil {
		return x.TempChangeRate
	}
	return 0
}

type SetTempResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The current temperature set point in degrees Celsius
	CurrentTempSetPoint float64 `protobuf:"fixed64,1,opt,name=current_temp_set_point,json=currentTempSetPoint,proto3" json:"current_temp_set_point,omitempty"`
	// The current temperature change rate in degrees Celsius per minute
	CurrentTempChangeRate float64 `protobuf:"fixed64,2,opt,name=current_temp_change_rate,json=currentTempChangeRate,proto3" json:"current_temp_change_rate,omitempty"`
	// The current average temperature of all temperature readings in degrees Celsius
	CurrentAvgTemp float64 `protobuf:"fixed64,3,opt,name=current_avg_temp,json=currentAvgTemp,proto3" json:"current_avg_temp,omitempty"`
}

func (x *SetTempResponse) Reset() {
	*x = SetTempResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_demorpc_controller_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SetTempResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SetTempResponse) ProtoMessage() {}

func (x *SetTempResponse) ProtoReflect() protoreflect.Message {
	mi := &file_demorpc_controller_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SetTempResponse.ProtoReflect.Descriptor instead.
func (*SetTempResponse) Descriptor() ([]byte, []int) {
	return file_demorpc_controller_proto_rawDescGZIP(), []int{1}
}

func (x *SetTempResponse) GetCurrentTempSetPoint() float64 {
	if x != nil {
		return x.CurrentTempSetPoint
	}
	return 0
}

func (x *SetTempResponse) GetCurrentTempChangeRate() float64 {
	if x != nil {
		return x.CurrentTempChangeRate
	}
	return 0
}

func (x *SetTempResponse) GetCurrentAvgTemp() float64 {
	if x != nil {
		return x.CurrentAvgTemp
	}
	return 0
}

type SetPresRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The desired pressure set point in Torr
	PressureSetPoint float64 `protobuf:"fixed64,1,opt,name=pressure_set_point,json=pressureSetPoint,proto3" json:"pressure_set_point,omitempty"`
	// The desired pressure change rate in Torr per minute
	PressureChangeRate float64 `protobuf:"fixed64,2,opt,name=pressure_change_rate,json=pressureChangeRate,proto3" json:"pressure_change_rate,omitempty"`
}

func (x *SetPresRequest) Reset() {
	*x = SetPresRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_demorpc_controller_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SetPresRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SetPresRequest) ProtoMessage() {}

func (x *SetPresRequest) ProtoReflect() protoreflect.Message {
	mi := &file_demorpc_controller_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SetPresRequest.ProtoReflect.Descriptor instead.
func (*SetPresRequest) Descriptor() ([]byte, []int) {
	return file_demorpc_controller_proto_rawDescGZIP(), []int{2}
}

func (x *SetPresRequest) GetPressureSetPoint() float64 {
	if x != nil {
		return x.PressureSetPoint
	}
	return 0
}

func (x *SetPresRequest) GetPressureChangeRate() float64 {
	if x != nil {
		return x.PressureChangeRate
	}
	return 0
}

type SetPresResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The current pressure set point in Torr
	CurrentPressureSetPoint float64 `protobuf:"fixed64,1,opt,name=current_pressure_set_point,json=currentPressureSetPoint,proto3" json:"current_pressure_set_point,omitempty"`
	// The current pressure change rate in Torr per minute
	CurrentPressureChangeRate float64 `protobuf:"fixed64,2,opt,name=current_pressure_change_rate,json=currentPressureChangeRate,proto3" json:"current_pressure_change_rate,omitempty"`
	// The current average pressure of all pressure readings in Torr
	CurrentAvgPressure float64 `protobuf:"fixed64,3,opt,name=current_avg_pressure,json=currentAvgPressure,proto3" json:"current_avg_pressure,omitempty"`
}

func (x *SetPresResponse) Reset() {
	*x = SetPresResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_demorpc_controller_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SetPresResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SetPresResponse) ProtoMessage() {}

func (x *SetPresResponse) ProtoReflect() protoreflect.Message {
	mi := &file_demorpc_controller_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SetPresResponse.ProtoReflect.Descriptor instead.
func (*SetPresResponse) Descriptor() ([]byte, []int) {
	return file_demorpc_controller_proto_rawDescGZIP(), []int{3}
}

func (x *SetPresResponse) GetCurrentPressureSetPoint() float64 {
	if x != nil {
		return x.CurrentPressureSetPoint
	}
	return 0
}

func (x *SetPresResponse) GetCurrentPressureChangeRate() float64 {
	if x != nil {
		return x.CurrentPressureChangeRate
	}
	return 0
}

func (x *SetPresResponse) GetCurrentAvgPressure() float64 {
	if x != nil {
		return x.CurrentAvgPressure
	}
	return 0
}

var File_demorpc_controller_proto protoreflect.FileDescriptor

var file_demorpc_controller_proto_rawDesc = []byte{
	0x0a, 0x18, 0x64, 0x65, 0x6d, 0x6f, 0x72, 0x70, 0x63, 0x2f, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f,
	0x6c, 0x6c, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x07, 0x64, 0x65, 0x6d, 0x6f,
	0x72, 0x70, 0x63, 0x22, 0x60, 0x0a, 0x0e, 0x53, 0x65, 0x74, 0x54, 0x65, 0x6d, 0x70, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x24, 0x0a, 0x0e, 0x74, 0x65, 0x6d, 0x70, 0x5f, 0x73, 0x65,
	0x74, 0x5f, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x01, 0x52, 0x0c, 0x74,
	0x65, 0x6d, 0x70, 0x53, 0x65, 0x74, 0x50, 0x6f, 0x69, 0x6e, 0x74, 0x12, 0x28, 0x0a, 0x10, 0x74,
	0x65, 0x6d, 0x70, 0x5f, 0x63, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x5f, 0x72, 0x61, 0x74, 0x65, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x01, 0x52, 0x0e, 0x74, 0x65, 0x6d, 0x70, 0x43, 0x68, 0x61, 0x6e, 0x67,
	0x65, 0x52, 0x61, 0x74, 0x65, 0x22, 0xa9, 0x01, 0x0a, 0x0f, 0x53, 0x65, 0x74, 0x54, 0x65, 0x6d,
	0x70, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x33, 0x0a, 0x16, 0x63, 0x75, 0x72,
	0x72, 0x65, 0x6e, 0x74, 0x5f, 0x74, 0x65, 0x6d, 0x70, 0x5f, 0x73, 0x65, 0x74, 0x5f, 0x70, 0x6f,
	0x69, 0x6e, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x01, 0x52, 0x13, 0x63, 0x75, 0x72, 0x72, 0x65,
	0x6e, 0x74, 0x54, 0x65, 0x6d, 0x70, 0x53, 0x65, 0x74, 0x50, 0x6f, 0x69, 0x6e, 0x74, 0x12, 0x37,
	0x0a, 0x18, 0x63, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x74, 0x5f, 0x74, 0x65, 0x6d, 0x70, 0x5f, 0x63,
	0x68, 0x61, 0x6e, 0x67, 0x65, 0x5f, 0x72, 0x61, 0x74, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x01,
	0x52, 0x15, 0x63, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x74, 0x54, 0x65, 0x6d, 0x70, 0x43, 0x68, 0x61,
	0x6e, 0x67, 0x65, 0x52, 0x61, 0x74, 0x65, 0x12, 0x28, 0x0a, 0x10, 0x63, 0x75, 0x72, 0x72, 0x65,
	0x6e, 0x74, 0x5f, 0x61, 0x76, 0x67, 0x5f, 0x74, 0x65, 0x6d, 0x70, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x01, 0x52, 0x0e, 0x63, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x74, 0x41, 0x76, 0x67, 0x54, 0x65, 0x6d,
	0x70, 0x22, 0x70, 0x0a, 0x0e, 0x53, 0x65, 0x74, 0x50, 0x72, 0x65, 0x73, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x12, 0x2c, 0x0a, 0x12, 0x70, 0x72, 0x65, 0x73, 0x73, 0x75, 0x72, 0x65, 0x5f,
	0x73, 0x65, 0x74, 0x5f, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x01, 0x52,
	0x10, 0x70, 0x72, 0x65, 0x73, 0x73, 0x75, 0x72, 0x65, 0x53, 0x65, 0x74, 0x50, 0x6f, 0x69, 0x6e,
	0x74, 0x12, 0x30, 0x0a, 0x14, 0x70, 0x72, 0x65, 0x73, 0x73, 0x75, 0x72, 0x65, 0x5f, 0x63, 0x68,
	0x61, 0x6e, 0x67, 0x65, 0x5f, 0x72, 0x61, 0x74, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x01, 0x52,
	0x12, 0x70, 0x72, 0x65, 0x73, 0x73, 0x75, 0x72, 0x65, 0x43, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x52,
	0x61, 0x74, 0x65, 0x22, 0xc1, 0x01, 0x0a, 0x0f, 0x53, 0x65, 0x74, 0x50, 0x72, 0x65, 0x73, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x3b, 0x0a, 0x1a, 0x63, 0x75, 0x72, 0x72, 0x65,
	0x6e, 0x74, 0x5f, 0x70, 0x72, 0x65, 0x73, 0x73, 0x75, 0x72, 0x65, 0x5f, 0x73, 0x65, 0x74, 0x5f,
	0x70, 0x6f, 0x69, 0x6e, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x01, 0x52, 0x17, 0x63, 0x75, 0x72,
	0x72, 0x65, 0x6e, 0x74, 0x50, 0x72, 0x65, 0x73, 0x73, 0x75, 0x72, 0x65, 0x53, 0x65, 0x74, 0x50,
	0x6f, 0x69, 0x6e, 0x74, 0x12, 0x3f, 0x0a, 0x1c, 0x63, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x74, 0x5f,
	0x70, 0x72, 0x65, 0x73, 0x73, 0x75, 0x72, 0x65, 0x5f, 0x63, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x5f,
	0x72, 0x61, 0x74, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x01, 0x52, 0x19, 0x63, 0x75, 0x72, 0x72,
	0x65, 0x6e, 0x74, 0x50, 0x72, 0x65, 0x73, 0x73, 0x75, 0x72, 0x65, 0x43, 0x68, 0x61, 0x6e, 0x67,
	0x65, 0x52, 0x61, 0x74, 0x65, 0x12, 0x30, 0x0a, 0x14, 0x63, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x74,
	0x5f, 0x61, 0x76, 0x67, 0x5f, 0x70, 0x72, 0x65, 0x73, 0x73, 0x75, 0x72, 0x65, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x01, 0x52, 0x12, 0x63, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x74, 0x41, 0x76, 0x67, 0x50,
	0x72, 0x65, 0x73, 0x73, 0x75, 0x72, 0x65, 0x32, 0x97, 0x01, 0x0a, 0x0a, 0x43, 0x6f, 0x6e, 0x74,
	0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x12, 0x45, 0x0a, 0x0e, 0x53, 0x65, 0x74, 0x54, 0x65, 0x6d,
	0x70, 0x65, 0x72, 0x61, 0x74, 0x75, 0x72, 0x65, 0x12, 0x17, 0x2e, 0x64, 0x65, 0x6d, 0x6f, 0x72,
	0x70, 0x63, 0x2e, 0x53, 0x65, 0x74, 0x54, 0x65, 0x6d, 0x70, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x18, 0x2e, 0x64, 0x65, 0x6d, 0x6f, 0x72, 0x70, 0x63, 0x2e, 0x53, 0x65, 0x74, 0x54,
	0x65, 0x6d, 0x70, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x30, 0x01, 0x12, 0x42, 0x0a,
	0x0b, 0x53, 0x65, 0x74, 0x50, 0x72, 0x65, 0x73, 0x73, 0x75, 0x72, 0x65, 0x12, 0x17, 0x2e, 0x64,
	0x65, 0x6d, 0x6f, 0x72, 0x70, 0x63, 0x2e, 0x53, 0x65, 0x74, 0x50, 0x72, 0x65, 0x73, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x18, 0x2e, 0x64, 0x65, 0x6d, 0x6f, 0x72, 0x70, 0x63, 0x2e,
	0x53, 0x65, 0x74, 0x50, 0x72, 0x65, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x30,
	0x01, 0x42, 0x2a, 0x5a, 0x28, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f,
	0x53, 0x53, 0x53, 0x4f, 0x43, 0x2d, 0x43, 0x41, 0x4e, 0x2f, 0x66, 0x6d, 0x74, 0x64, 0x2f, 0x66,
	0x6d, 0x74, 0x72, 0x70, 0x63, 0x2f, 0x64, 0x65, 0x6d, 0x6f, 0x72, 0x70, 0x63, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_demorpc_controller_proto_rawDescOnce sync.Once
	file_demorpc_controller_proto_rawDescData = file_demorpc_controller_proto_rawDesc
)

func file_demorpc_controller_proto_rawDescGZIP() []byte {
	file_demorpc_controller_proto_rawDescOnce.Do(func() {
		file_demorpc_controller_proto_rawDescData = protoimpl.X.CompressGZIP(file_demorpc_controller_proto_rawDescData)
	})
	return file_demorpc_controller_proto_rawDescData
}

var file_demorpc_controller_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_demorpc_controller_proto_goTypes = []interface{}{
	(*SetTempRequest)(nil),  // 0: demorpc.SetTempRequest
	(*SetTempResponse)(nil), // 1: demorpc.SetTempResponse
	(*SetPresRequest)(nil),  // 2: demorpc.SetPresRequest
	(*SetPresResponse)(nil), // 3: demorpc.SetPresResponse
}
var file_demorpc_controller_proto_depIdxs = []int32{
	0, // 0: demorpc.Controller.SetTemperature:input_type -> demorpc.SetTempRequest
	2, // 1: demorpc.Controller.SetPressure:input_type -> demorpc.SetPresRequest
	1, // 2: demorpc.Controller.SetTemperature:output_type -> demorpc.SetTempResponse
	3, // 3: demorpc.Controller.SetPressure:output_type -> demorpc.SetPresResponse
	2, // [2:4] is the sub-list for method output_type
	0, // [0:2] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_demorpc_controller_proto_init() }
func file_demorpc_controller_proto_init() {
	if File_demorpc_controller_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_demorpc_controller_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SetTempRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_demorpc_controller_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SetTempResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_demorpc_controller_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SetPresRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_demorpc_controller_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SetPresResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_demorpc_controller_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_demorpc_controller_proto_goTypes,
		DependencyIndexes: file_demorpc_controller_proto_depIdxs,
		MessageInfos:      file_demorpc_controller_proto_msgTypes,
	}.Build()
	File_demorpc_controller_proto = out.File
	file_demorpc_controller_proto_rawDesc = nil
	file_demorpc_controller_proto_goTypes = nil
	file_demorpc_controller_proto_depIdxs = nil
}
