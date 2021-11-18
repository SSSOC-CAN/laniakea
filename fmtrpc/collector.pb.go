// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.18.1
// source: collector.proto

package fmtrpc

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

type RecordService int32

const (
	RecordService_FLUKE RecordService = 0
	RecordService_RGA   RecordService = 1
)

// Enum value maps for RecordService.
var (
	RecordService_name = map[int32]string{
		0: "FLUKE",
		1: "RGA",
	}
	RecordService_value = map[string]int32{
		"FLUKE": 0,
		"RGA":   1,
	}
)

func (x RecordService) Enum() *RecordService {
	p := new(RecordService)
	*p = x
	return p
}

func (x RecordService) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (RecordService) Descriptor() protoreflect.EnumDescriptor {
	return file_collector_proto_enumTypes[0].Descriptor()
}

func (RecordService) Type() protoreflect.EnumType {
	return &file_collector_proto_enumTypes[0]
}

func (x RecordService) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use RecordService.Descriptor instead.
func (RecordService) EnumDescriptor() ([]byte, []int) {
	return file_collector_proto_rawDescGZIP(), []int{0}
}

type RecordRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	PollingInterval int64         `protobuf:"varint,1,opt,name=polling_interval,json=pollingInterval,proto3" json:"polling_interval,omitempty"`
	Type            RecordService `protobuf:"varint,2,opt,name=type,proto3,enum=fmtrpc.RecordService" json:"type,omitempty"`
}

func (x *RecordRequest) Reset() {
	*x = RecordRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_collector_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RecordRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RecordRequest) ProtoMessage() {}

func (x *RecordRequest) ProtoReflect() protoreflect.Message {
	mi := &file_collector_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RecordRequest.ProtoReflect.Descriptor instead.
func (*RecordRequest) Descriptor() ([]byte, []int) {
	return file_collector_proto_rawDescGZIP(), []int{0}
}

func (x *RecordRequest) GetPollingInterval() int64 {
	if x != nil {
		return x.PollingInterval
	}
	return 0
}

func (x *RecordRequest) GetType() RecordService {
	if x != nil {
		return x.Type
	}
	return RecordService_FLUKE
}

type RecordResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Msg string `protobuf:"bytes,1,opt,name=msg,proto3" json:"msg,omitempty"`
}

func (x *RecordResponse) Reset() {
	*x = RecordResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_collector_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RecordResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RecordResponse) ProtoMessage() {}

func (x *RecordResponse) ProtoReflect() protoreflect.Message {
	mi := &file_collector_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RecordResponse.ProtoReflect.Descriptor instead.
func (*RecordResponse) Descriptor() ([]byte, []int) {
	return file_collector_proto_rawDescGZIP(), []int{1}
}

func (x *RecordResponse) GetMsg() string {
	if x != nil {
		return x.Msg
	}
	return ""
}

type StopRecRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Type RecordService `protobuf:"varint,1,opt,name=type,proto3,enum=fmtrpc.RecordService" json:"type,omitempty"`
}

func (x *StopRecRequest) Reset() {
	*x = StopRecRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_collector_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StopRecRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StopRecRequest) ProtoMessage() {}

func (x *StopRecRequest) ProtoReflect() protoreflect.Message {
	mi := &file_collector_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StopRecRequest.ProtoReflect.Descriptor instead.
func (*StopRecRequest) Descriptor() ([]byte, []int) {
	return file_collector_proto_rawDescGZIP(), []int{2}
}

func (x *StopRecRequest) GetType() RecordService {
	if x != nil {
		return x.Type
	}
	return RecordService_FLUKE
}

type StopRecResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Msg string `protobuf:"bytes,1,opt,name=msg,proto3" json:"msg,omitempty"`
}

func (x *StopRecResponse) Reset() {
	*x = StopRecResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_collector_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StopRecResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StopRecResponse) ProtoMessage() {}

func (x *StopRecResponse) ProtoReflect() protoreflect.Message {
	mi := &file_collector_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StopRecResponse.ProtoReflect.Descriptor instead.
func (*StopRecResponse) Descriptor() ([]byte, []int) {
	return file_collector_proto_rawDescGZIP(), []int{3}
}

func (x *StopRecResponse) GetMsg() string {
	if x != nil {
		return x.Msg
	}
	return ""
}

type SubscribeDataRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *SubscribeDataRequest) Reset() {
	*x = SubscribeDataRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_collector_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SubscribeDataRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SubscribeDataRequest) ProtoMessage() {}

func (x *SubscribeDataRequest) ProtoReflect() protoreflect.Message {
	mi := &file_collector_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SubscribeDataRequest.ProtoReflect.Descriptor instead.
func (*SubscribeDataRequest) Descriptor() ([]byte, []int) {
	return file_collector_proto_rawDescGZIP(), []int{4}
}

type DataField struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name  string  `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Value float64 `protobuf:"fixed64,2,opt,name=value,proto3" json:"value,omitempty"`
}

func (x *DataField) Reset() {
	*x = DataField{}
	if protoimpl.UnsafeEnabled {
		mi := &file_collector_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DataField) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DataField) ProtoMessage() {}

func (x *DataField) ProtoReflect() protoreflect.Message {
	mi := &file_collector_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DataField.ProtoReflect.Descriptor instead.
func (*DataField) Descriptor() ([]byte, []int) {
	return file_collector_proto_rawDescGZIP(), []int{5}
}

func (x *DataField) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *DataField) GetValue() float64 {
	if x != nil {
		return x.Value
	}
	return 0
}

type RealTimeData struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Source     string               `protobuf:"bytes,1,opt,name=source,proto3" json:"source,omitempty"`
	IsScanning bool                 `protobuf:"varint,2,opt,name=is_scanning,json=isScanning,proto3" json:"is_scanning,omitempty"`
	Timestamp  int64                `protobuf:"varint,3,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	Data       map[int64]*DataField `protobuf:"bytes,4,rep,name=data,proto3" json:"data,omitempty" protobuf_key:"varint,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *RealTimeData) Reset() {
	*x = RealTimeData{}
	if protoimpl.UnsafeEnabled {
		mi := &file_collector_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RealTimeData) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RealTimeData) ProtoMessage() {}

func (x *RealTimeData) ProtoReflect() protoreflect.Message {
	mi := &file_collector_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RealTimeData.ProtoReflect.Descriptor instead.
func (*RealTimeData) Descriptor() ([]byte, []int) {
	return file_collector_proto_rawDescGZIP(), []int{6}
}

func (x *RealTimeData) GetSource() string {
	if x != nil {
		return x.Source
	}
	return ""
}

func (x *RealTimeData) GetIsScanning() bool {
	if x != nil {
		return x.IsScanning
	}
	return false
}

func (x *RealTimeData) GetTimestamp() int64 {
	if x != nil {
		return x.Timestamp
	}
	return 0
}

func (x *RealTimeData) GetData() map[int64]*DataField {
	if x != nil {
		return x.Data
	}
	return nil
}

type HistoricalDataRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Source RecordService `protobuf:"varint,1,opt,name=source,proto3,enum=fmtrpc.RecordService" json:"source,omitempty"`
}

func (x *HistoricalDataRequest) Reset() {
	*x = HistoricalDataRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_collector_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *HistoricalDataRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HistoricalDataRequest) ProtoMessage() {}

func (x *HistoricalDataRequest) ProtoReflect() protoreflect.Message {
	mi := &file_collector_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HistoricalDataRequest.ProtoReflect.Descriptor instead.
func (*HistoricalDataRequest) Descriptor() ([]byte, []int) {
	return file_collector_proto_rawDescGZIP(), []int{7}
}

func (x *HistoricalDataRequest) GetSource() RecordService {
	if x != nil {
		return x.Source
	}
	return RecordService_FLUKE
}

type HistoricalDataResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ServerPort int64 `protobuf:"varint,1,opt,name=server_port,json=serverPort,proto3" json:"server_port,omitempty"`
	BufferSize int64 `protobuf:"varint,2,opt,name=buffer_size,json=bufferSize,proto3" json:"buffer_size,omitempty"`
}

func (x *HistoricalDataResponse) Reset() {
	*x = HistoricalDataResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_collector_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *HistoricalDataResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HistoricalDataResponse) ProtoMessage() {}

func (x *HistoricalDataResponse) ProtoReflect() protoreflect.Message {
	mi := &file_collector_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HistoricalDataResponse.ProtoReflect.Descriptor instead.
func (*HistoricalDataResponse) Descriptor() ([]byte, []int) {
	return file_collector_proto_rawDescGZIP(), []int{8}
}

func (x *HistoricalDataResponse) GetServerPort() int64 {
	if x != nil {
		return x.ServerPort
	}
	return 0
}

func (x *HistoricalDataResponse) GetBufferSize() int64 {
	if x != nil {
		return x.BufferSize
	}
	return 0
}

var File_collector_proto protoreflect.FileDescriptor

var file_collector_proto_rawDesc = []byte{
	0x0a, 0x0f, 0x63, 0x6f, 0x6c, 0x6c, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x06, 0x66, 0x6d, 0x74, 0x72, 0x70, 0x63, 0x22, 0x65, 0x0a, 0x0d, 0x52, 0x65, 0x63,
	0x6f, 0x72, 0x64, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x29, 0x0a, 0x10, 0x70, 0x6f,
	0x6c, 0x6c, 0x69, 0x6e, 0x67, 0x5f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x76, 0x61, 0x6c, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x03, 0x52, 0x0f, 0x70, 0x6f, 0x6c, 0x6c, 0x69, 0x6e, 0x67, 0x49, 0x6e, 0x74,
	0x65, 0x72, 0x76, 0x61, 0x6c, 0x12, 0x29, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x0e, 0x32, 0x15, 0x2e, 0x66, 0x6d, 0x74, 0x72, 0x70, 0x63, 0x2e, 0x52, 0x65, 0x63,
	0x6f, 0x72, 0x64, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65,
	0x22, 0x22, 0x0a, 0x0e, 0x52, 0x65, 0x63, 0x6f, 0x72, 0x64, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x6d, 0x73, 0x67, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x03, 0x6d, 0x73, 0x67, 0x22, 0x3b, 0x0a, 0x0e, 0x53, 0x74, 0x6f, 0x70, 0x52, 0x65, 0x63, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x29, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0e, 0x32, 0x15, 0x2e, 0x66, 0x6d, 0x74, 0x72, 0x70, 0x63, 0x2e, 0x52, 0x65,
	0x63, 0x6f, 0x72, 0x64, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x52, 0x04, 0x74, 0x79, 0x70,
	0x65, 0x22, 0x23, 0x0a, 0x0f, 0x53, 0x74, 0x6f, 0x70, 0x52, 0x65, 0x63, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x6d, 0x73, 0x67, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x03, 0x6d, 0x73, 0x67, 0x22, 0x16, 0x0a, 0x14, 0x53, 0x75, 0x62, 0x73, 0x63, 0x72,
	0x69, 0x62, 0x65, 0x44, 0x61, 0x74, 0x61, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0x35,
	0x0a, 0x09, 0x44, 0x61, 0x74, 0x61, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x6e,
	0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12,
	0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x01, 0x52, 0x05,
	0x76, 0x61, 0x6c, 0x75, 0x65, 0x22, 0xe5, 0x01, 0x0a, 0x0c, 0x52, 0x65, 0x61, 0x6c, 0x54, 0x69,
	0x6d, 0x65, 0x44, 0x61, 0x74, 0x61, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x12, 0x1f,
	0x0a, 0x0b, 0x69, 0x73, 0x5f, 0x73, 0x63, 0x61, 0x6e, 0x6e, 0x69, 0x6e, 0x67, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x08, 0x52, 0x0a, 0x69, 0x73, 0x53, 0x63, 0x61, 0x6e, 0x6e, 0x69, 0x6e, 0x67, 0x12,
	0x1c, 0x0a, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x03, 0x52, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x12, 0x32, 0x0a,
	0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x04, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1e, 0x2e, 0x66, 0x6d,
	0x74, 0x72, 0x70, 0x63, 0x2e, 0x52, 0x65, 0x61, 0x6c, 0x54, 0x69, 0x6d, 0x65, 0x44, 0x61, 0x74,
	0x61, 0x2e, 0x44, 0x61, 0x74, 0x61, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x04, 0x64, 0x61, 0x74,
	0x61, 0x1a, 0x4a, 0x0a, 0x09, 0x44, 0x61, 0x74, 0x61, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10,
	0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x03, 0x6b, 0x65, 0x79,
	0x12, 0x27, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x11, 0x2e, 0x66, 0x6d, 0x74, 0x72, 0x70, 0x63, 0x2e, 0x44, 0x61, 0x74, 0x61, 0x46, 0x69, 0x65,
	0x6c, 0x64, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0x46, 0x0a,
	0x15, 0x48, 0x69, 0x73, 0x74, 0x6f, 0x72, 0x69, 0x63, 0x61, 0x6c, 0x44, 0x61, 0x74, 0x61, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x2d, 0x0a, 0x06, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x15, 0x2e, 0x66, 0x6d, 0x74, 0x72, 0x70, 0x63, 0x2e,
	0x52, 0x65, 0x63, 0x6f, 0x72, 0x64, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x52, 0x06, 0x73,
	0x6f, 0x75, 0x72, 0x63, 0x65, 0x22, 0x5a, 0x0a, 0x16, 0x48, 0x69, 0x73, 0x74, 0x6f, 0x72, 0x69,
	0x63, 0x61, 0x6c, 0x44, 0x61, 0x74, 0x61, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12,
	0x1f, 0x0a, 0x0b, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x5f, 0x70, 0x6f, 0x72, 0x74, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x03, 0x52, 0x0a, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x50, 0x6f, 0x72, 0x74,
	0x12, 0x1f, 0x0a, 0x0b, 0x62, 0x75, 0x66, 0x66, 0x65, 0x72, 0x5f, 0x73, 0x69, 0x7a, 0x65, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0a, 0x62, 0x75, 0x66, 0x66, 0x65, 0x72, 0x53, 0x69, 0x7a,
	0x65, 0x2a, 0x23, 0x0a, 0x0d, 0x52, 0x65, 0x63, 0x6f, 0x72, 0x64, 0x53, 0x65, 0x72, 0x76, 0x69,
	0x63, 0x65, 0x12, 0x09, 0x0a, 0x05, 0x46, 0x4c, 0x55, 0x4b, 0x45, 0x10, 0x00, 0x12, 0x07, 0x0a,
	0x03, 0x52, 0x47, 0x41, 0x10, 0x01, 0x32, 0xb8, 0x02, 0x0a, 0x0d, 0x44, 0x61, 0x74, 0x61, 0x43,
	0x6f, 0x6c, 0x6c, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x12, 0x3f, 0x0a, 0x0e, 0x53, 0x74, 0x61, 0x72,
	0x74, 0x52, 0x65, 0x63, 0x6f, 0x72, 0x64, 0x69, 0x6e, 0x67, 0x12, 0x15, 0x2e, 0x66, 0x6d, 0x74,
	0x72, 0x70, 0x63, 0x2e, 0x52, 0x65, 0x63, 0x6f, 0x72, 0x64, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x16, 0x2e, 0x66, 0x6d, 0x74, 0x72, 0x70, 0x63, 0x2e, 0x52, 0x65, 0x63, 0x6f, 0x72,
	0x64, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x40, 0x0a, 0x0d, 0x53, 0x74, 0x6f,
	0x70, 0x52, 0x65, 0x63, 0x6f, 0x72, 0x64, 0x69, 0x6e, 0x67, 0x12, 0x16, 0x2e, 0x66, 0x6d, 0x74,
	0x72, 0x70, 0x63, 0x2e, 0x53, 0x74, 0x6f, 0x70, 0x52, 0x65, 0x63, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x17, 0x2e, 0x66, 0x6d, 0x74, 0x72, 0x70, 0x63, 0x2e, 0x53, 0x74, 0x6f, 0x70,
	0x52, 0x65, 0x63, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x4b, 0x0a, 0x13, 0x53,
	0x75, 0x62, 0x73, 0x63, 0x72, 0x69, 0x62, 0x65, 0x44, 0x61, 0x74, 0x61, 0x53, 0x74, 0x72, 0x65,
	0x61, 0x6d, 0x12, 0x1c, 0x2e, 0x66, 0x6d, 0x74, 0x72, 0x70, 0x63, 0x2e, 0x53, 0x75, 0x62, 0x73,
	0x63, 0x72, 0x69, 0x62, 0x65, 0x44, 0x61, 0x74, 0x61, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x1a, 0x14, 0x2e, 0x66, 0x6d, 0x74, 0x72, 0x70, 0x63, 0x2e, 0x52, 0x65, 0x61, 0x6c, 0x54, 0x69,
	0x6d, 0x65, 0x44, 0x61, 0x74, 0x61, 0x30, 0x01, 0x12, 0x57, 0x0a, 0x16, 0x44, 0x6f, 0x77, 0x6e,
	0x6c, 0x6f, 0x61, 0x64, 0x48, 0x69, 0x73, 0x74, 0x6f, 0x72, 0x69, 0x63, 0x61, 0x6c, 0x44, 0x61,
	0x74, 0x61, 0x12, 0x1d, 0x2e, 0x66, 0x6d, 0x74, 0x72, 0x70, 0x63, 0x2e, 0x48, 0x69, 0x73, 0x74,
	0x6f, 0x72, 0x69, 0x63, 0x61, 0x6c, 0x44, 0x61, 0x74, 0x61, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x1e, 0x2e, 0x66, 0x6d, 0x74, 0x72, 0x70, 0x63, 0x2e, 0x48, 0x69, 0x73, 0x74, 0x6f,
	0x72, 0x69, 0x63, 0x61, 0x6c, 0x44, 0x61, 0x74, 0x61, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x42, 0x22, 0x5a, 0x20, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f,
	0x53, 0x53, 0x53, 0x4f, 0x43, 0x2d, 0x43, 0x41, 0x4e, 0x2f, 0x66, 0x6d, 0x74, 0x64, 0x2f, 0x66,
	0x6d, 0x74, 0x72, 0x70, 0x63, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_collector_proto_rawDescOnce sync.Once
	file_collector_proto_rawDescData = file_collector_proto_rawDesc
)

func file_collector_proto_rawDescGZIP() []byte {
	file_collector_proto_rawDescOnce.Do(func() {
		file_collector_proto_rawDescData = protoimpl.X.CompressGZIP(file_collector_proto_rawDescData)
	})
	return file_collector_proto_rawDescData
}

var file_collector_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_collector_proto_msgTypes = make([]protoimpl.MessageInfo, 10)
var file_collector_proto_goTypes = []interface{}{
	(RecordService)(0),             // 0: fmtrpc.RecordService
	(*RecordRequest)(nil),          // 1: fmtrpc.RecordRequest
	(*RecordResponse)(nil),         // 2: fmtrpc.RecordResponse
	(*StopRecRequest)(nil),         // 3: fmtrpc.StopRecRequest
	(*StopRecResponse)(nil),        // 4: fmtrpc.StopRecResponse
	(*SubscribeDataRequest)(nil),   // 5: fmtrpc.SubscribeDataRequest
	(*DataField)(nil),              // 6: fmtrpc.DataField
	(*RealTimeData)(nil),           // 7: fmtrpc.RealTimeData
	(*HistoricalDataRequest)(nil),  // 8: fmtrpc.HistoricalDataRequest
	(*HistoricalDataResponse)(nil), // 9: fmtrpc.HistoricalDataResponse
	nil,                            // 10: fmtrpc.RealTimeData.DataEntry
}
var file_collector_proto_depIdxs = []int32{
	0,  // 0: fmtrpc.RecordRequest.type:type_name -> fmtrpc.RecordService
	0,  // 1: fmtrpc.StopRecRequest.type:type_name -> fmtrpc.RecordService
	10, // 2: fmtrpc.RealTimeData.data:type_name -> fmtrpc.RealTimeData.DataEntry
	0,  // 3: fmtrpc.HistoricalDataRequest.source:type_name -> fmtrpc.RecordService
	6,  // 4: fmtrpc.RealTimeData.DataEntry.value:type_name -> fmtrpc.DataField
	1,  // 5: fmtrpc.DataCollector.StartRecording:input_type -> fmtrpc.RecordRequest
	3,  // 6: fmtrpc.DataCollector.StopRecording:input_type -> fmtrpc.StopRecRequest
	5,  // 7: fmtrpc.DataCollector.SubscribeDataStream:input_type -> fmtrpc.SubscribeDataRequest
	8,  // 8: fmtrpc.DataCollector.DownloadHistoricalData:input_type -> fmtrpc.HistoricalDataRequest
	2,  // 9: fmtrpc.DataCollector.StartRecording:output_type -> fmtrpc.RecordResponse
	4,  // 10: fmtrpc.DataCollector.StopRecording:output_type -> fmtrpc.StopRecResponse
	7,  // 11: fmtrpc.DataCollector.SubscribeDataStream:output_type -> fmtrpc.RealTimeData
	9,  // 12: fmtrpc.DataCollector.DownloadHistoricalData:output_type -> fmtrpc.HistoricalDataResponse
	9,  // [9:13] is the sub-list for method output_type
	5,  // [5:9] is the sub-list for method input_type
	5,  // [5:5] is the sub-list for extension type_name
	5,  // [5:5] is the sub-list for extension extendee
	0,  // [0:5] is the sub-list for field type_name
}

func init() { file_collector_proto_init() }
func file_collector_proto_init() {
	if File_collector_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_collector_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RecordRequest); i {
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
		file_collector_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RecordResponse); i {
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
		file_collector_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StopRecRequest); i {
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
		file_collector_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StopRecResponse); i {
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
		file_collector_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SubscribeDataRequest); i {
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
		file_collector_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DataField); i {
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
		file_collector_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RealTimeData); i {
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
		file_collector_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*HistoricalDataRequest); i {
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
		file_collector_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*HistoricalDataResponse); i {
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
			RawDescriptor: file_collector_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   10,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_collector_proto_goTypes,
		DependencyIndexes: file_collector_proto_depIdxs,
		EnumInfos:         file_collector_proto_enumTypes,
		MessageInfos:      file_collector_proto_msgTypes,
	}.Build()
	File_collector_proto = out.File
	file_collector_proto_rawDesc = nil
	file_collector_proto_goTypes = nil
	file_collector_proto_depIdxs = nil
}
