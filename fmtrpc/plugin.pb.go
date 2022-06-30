//
//Author: Paul Côté
//Last Change Author: Paul Côté
//Last Date Changed: 2022/06/30

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.19.4
// source: plugin.proto

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

type Plugin_PluginType int32

const (
	Plugin_DATASOURCE Plugin_PluginType = 0
	Plugin_CONTROLLER Plugin_PluginType = 1
)

// Enum value maps for Plugin_PluginType.
var (
	Plugin_PluginType_name = map[int32]string{
		0: "DATASOURCE",
		1: "CONTROLLER",
	}
	Plugin_PluginType_value = map[string]int32{
		"DATASOURCE": 0,
		"CONTROLLER": 1,
	}
)

func (x Plugin_PluginType) Enum() *Plugin_PluginType {
	p := new(Plugin_PluginType)
	*p = x
	return p
}

func (x Plugin_PluginType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Plugin_PluginType) Descriptor() protoreflect.EnumDescriptor {
	return file_plugin_proto_enumTypes[0].Descriptor()
}

func (Plugin_PluginType) Type() protoreflect.EnumType {
	return &file_plugin_proto_enumTypes[0]
}

func (x Plugin_PluginType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Plugin_PluginType.Descriptor instead.
func (Plugin_PluginType) EnumDescriptor() ([]byte, []int) {
	return file_plugin_proto_rawDescGZIP(), []int{3, 0}
}

type Plugin_PluginState int32

const (
	Plugin_READY        Plugin_PluginState = 0
	Plugin_BUSY         Plugin_PluginState = 1
	Plugin_STOPPING     Plugin_PluginState = 2
	Plugin_STOPPED      Plugin_PluginState = 3
	Plugin_UNKNOWN      Plugin_PluginState = 4
	Plugin_UNRESPONSIVE Plugin_PluginState = 5
	Plugin_KILLED       Plugin_PluginState = 6
)

// Enum value maps for Plugin_PluginState.
var (
	Plugin_PluginState_name = map[int32]string{
		0: "READY",
		1: "BUSY",
		2: "STOPPING",
		3: "STOPPED",
		4: "UNKNOWN",
		5: "UNRESPONSIVE",
		6: "KILLED",
	}
	Plugin_PluginState_value = map[string]int32{
		"READY":        0,
		"BUSY":         1,
		"STOPPING":     2,
		"STOPPED":      3,
		"UNKNOWN":      4,
		"UNRESPONSIVE": 5,
		"KILLED":       6,
	}
)

func (x Plugin_PluginState) Enum() *Plugin_PluginState {
	p := new(Plugin_PluginState)
	*p = x
	return p
}

func (x Plugin_PluginState) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Plugin_PluginState) Descriptor() protoreflect.EnumDescriptor {
	return file_plugin_proto_enumTypes[1].Descriptor()
}

func (Plugin_PluginState) Type() protoreflect.EnumType {
	return &file_plugin_proto_enumTypes[1]
}

func (x Plugin_PluginState) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Plugin_PluginState.Descriptor instead.
func (Plugin_PluginState) EnumDescriptor() ([]byte, []int) {
	return file_plugin_proto_rawDescGZIP(), []int{3, 1}
}

type Empty struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *Empty) Reset() {
	*x = Empty{}
	if protoimpl.UnsafeEnabled {
		mi := &file_plugin_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Empty) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Empty) ProtoMessage() {}

func (x *Empty) ProtoReflect() protoreflect.Message {
	mi := &file_plugin_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Empty.ProtoReflect.Descriptor instead.
func (*Empty) Descriptor() ([]byte, []int) {
	return file_plugin_proto_rawDescGZIP(), []int{0}
}

type Frame struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The name of the source of this Frame
	Source string `protobuf:"bytes,1,opt,name=source,proto3" json:"source,omitempty"`
	// A MIME-like type indicating the kind of content within the payload field
	Type string `protobuf:"bytes,2,opt,name=type,proto3" json:"type,omitempty"`
	// The UNIX millisecond timestamp of this frame
	Timestamp int64 `protobuf:"varint,3,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	// The actual payload data in bytes. Limit is 2^32
	Payload []byte `protobuf:"bytes,4,opt,name=payload,proto3" json:"payload,omitempty"`
}

func (x *Frame) Reset() {
	*x = Frame{}
	if protoimpl.UnsafeEnabled {
		mi := &file_plugin_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Frame) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Frame) ProtoMessage() {}

func (x *Frame) ProtoReflect() protoreflect.Message {
	mi := &file_plugin_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Frame.ProtoReflect.Descriptor instead.
func (*Frame) Descriptor() ([]byte, []int) {
	return file_plugin_proto_rawDescGZIP(), []int{1}
}

func (x *Frame) GetSource() string {
	if x != nil {
		return x.Source
	}
	return ""
}

func (x *Frame) GetType() string {
	if x != nil {
		return x.Type
	}
	return ""
}

func (x *Frame) GetTimestamp() int64 {
	if x != nil {
		return x.Timestamp
	}
	return 0
}

func (x *Frame) GetPayload() []byte {
	if x != nil {
		return x.Payload
	}
	return nil
}

type PluginRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The name of the plugin we want to interact with
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
}

func (x *PluginRequest) Reset() {
	*x = PluginRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_plugin_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PluginRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PluginRequest) ProtoMessage() {}

func (x *PluginRequest) ProtoReflect() protoreflect.Message {
	mi := &file_plugin_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PluginRequest.ProtoReflect.Descriptor instead.
func (*PluginRequest) Descriptor() ([]byte, []int) {
	return file_plugin_proto_rawDescGZIP(), []int{2}
}

func (x *PluginRequest) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

type Plugin struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The name of the plugin
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// The plugin type (either Datasource or Controller)
	Type Plugin_PluginType `protobuf:"varint,2,opt,name=type,proto3,enum=fmtrpc.Plugin_PluginType" json:"type,omitempty"`
	// the current state of the plugin
	State Plugin_PluginState `protobuf:"varint,3,opt,name=state,proto3,enum=fmtrpc.Plugin_PluginState" json:"state,omitempty"`
	// Unix milli timestamp of when the plugin was started
	StartedAt int64 `protobuf:"varint,4,opt,name=started_at,json=startedAt,proto3" json:"started_at,omitempty"`
	// Unix milli timestamp of when the plugin was stopped or killed. Value is 0 if it's not stopped or killed
	StoppedAt int64 `protobuf:"varint,5,opt,name=stopped_at,json=stoppedAt,proto3" json:"stopped_at,omitempty"`
}

func (x *Plugin) Reset() {
	*x = Plugin{}
	if protoimpl.UnsafeEnabled {
		mi := &file_plugin_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Plugin) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Plugin) ProtoMessage() {}

func (x *Plugin) ProtoReflect() protoreflect.Message {
	mi := &file_plugin_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Plugin.ProtoReflect.Descriptor instead.
func (*Plugin) Descriptor() ([]byte, []int) {
	return file_plugin_proto_rawDescGZIP(), []int{3}
}

func (x *Plugin) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Plugin) GetType() Plugin_PluginType {
	if x != nil {
		return x.Type
	}
	return Plugin_DATASOURCE
}

func (x *Plugin) GetState() Plugin_PluginState {
	if x != nil {
		return x.State
	}
	return Plugin_READY
}

func (x *Plugin) GetStartedAt() int64 {
	if x != nil {
		return x.StartedAt
	}
	return 0
}

func (x *Plugin) GetStoppedAt() int64 {
	if x != nil {
		return x.StoppedAt
	}
	return 0
}

type PluginsList struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// List of all currently registered plugins
	Plugins []*Plugin `protobuf:"bytes,1,rep,name=plugins,proto3" json:"plugins,omitempty"`
}

func (x *PluginsList) Reset() {
	*x = PluginsList{}
	if protoimpl.UnsafeEnabled {
		mi := &file_plugin_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PluginsList) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PluginsList) ProtoMessage() {}

func (x *PluginsList) ProtoReflect() protoreflect.Message {
	mi := &file_plugin_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PluginsList.ProtoReflect.Descriptor instead.
func (*PluginsList) Descriptor() ([]byte, []int) {
	return file_plugin_proto_rawDescGZIP(), []int{4}
}

func (x *PluginsList) GetPlugins() []*Plugin {
	if x != nil {
		return x.Plugins
	}
	return nil
}

type ControllerPluginRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// the name of the plugin we wish to send the command to
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// the data we are sending to the plugin
	Frame *Frame `protobuf:"bytes,2,opt,name=frame,proto3" json:"frame,omitempty"`
}

func (x *ControllerPluginRequest) Reset() {
	*x = ControllerPluginRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_plugin_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ControllerPluginRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ControllerPluginRequest) ProtoMessage() {}

func (x *ControllerPluginRequest) ProtoReflect() protoreflect.Message {
	mi := &file_plugin_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ControllerPluginRequest.ProtoReflect.Descriptor instead.
func (*ControllerPluginRequest) Descriptor() ([]byte, []int) {
	return file_plugin_proto_rawDescGZIP(), []int{5}
}

func (x *ControllerPluginRequest) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *ControllerPluginRequest) GetFrame() *Frame {
	if x != nil {
		return x.Frame
	}
	return nil
}

type AddPluginRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The plugin information string in format unique_name:type(datasource or controller):executable_file_name.ext
	PluginString string `protobuf:"bytes,1,opt,name=plugin_string,json=pluginString,proto3" json:"plugin_string,omitempty"`
}

func (x *AddPluginRequest) Reset() {
	*x = AddPluginRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_plugin_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AddPluginRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AddPluginRequest) ProtoMessage() {}

func (x *AddPluginRequest) ProtoReflect() protoreflect.Message {
	mi := &file_plugin_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AddPluginRequest.ProtoReflect.Descriptor instead.
func (*AddPluginRequest) Descriptor() ([]byte, []int) {
	return file_plugin_proto_rawDescGZIP(), []int{6}
}

func (x *AddPluginRequest) GetPluginString() string {
	if x != nil {
		return x.PluginString
	}
	return ""
}

var File_plugin_proto protoreflect.FileDescriptor

var file_plugin_proto_rawDesc = []byte{
	0x0a, 0x0c, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x06,
	0x66, 0x6d, 0x74, 0x72, 0x70, 0x63, 0x22, 0x07, 0x0a, 0x05, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x22,
	0x6b, 0x0a, 0x05, 0x46, 0x72, 0x61, 0x6d, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x6f, 0x75, 0x72,
	0x63, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65,
	0x12, 0x12, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04,
	0x74, 0x79, 0x70, 0x65, 0x12, 0x1c, 0x0a, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d,
	0x70, 0x18, 0x03, 0x20, 0x01, 0x28, 0x03, 0x52, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61,
	0x6d, 0x70, 0x12, 0x18, 0x0a, 0x07, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x18, 0x04, 0x20,
	0x01, 0x28, 0x0c, 0x52, 0x07, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x22, 0x23, 0x0a, 0x0d,
	0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x12, 0x0a,
	0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d,
	0x65, 0x22, 0xd3, 0x02, 0x0a, 0x06, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x12, 0x12, 0x0a, 0x04,
	0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65,
	0x12, 0x2d, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x19,
	0x2e, 0x66, 0x6d, 0x74, 0x72, 0x70, 0x63, 0x2e, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x2e, 0x50,
	0x6c, 0x75, 0x67, 0x69, 0x6e, 0x54, 0x79, 0x70, 0x65, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12,
	0x30, 0x0a, 0x05, 0x73, 0x74, 0x61, 0x74, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x1a,
	0x2e, 0x66, 0x6d, 0x74, 0x72, 0x70, 0x63, 0x2e, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x2e, 0x50,
	0x6c, 0x75, 0x67, 0x69, 0x6e, 0x53, 0x74, 0x61, 0x74, 0x65, 0x52, 0x05, 0x73, 0x74, 0x61, 0x74,
	0x65, 0x12, 0x1d, 0x0a, 0x0a, 0x73, 0x74, 0x61, 0x72, 0x74, 0x65, 0x64, 0x5f, 0x61, 0x74, 0x18,
	0x04, 0x20, 0x01, 0x28, 0x03, 0x52, 0x09, 0x73, 0x74, 0x61, 0x72, 0x74, 0x65, 0x64, 0x41, 0x74,
	0x12, 0x1d, 0x0a, 0x0a, 0x73, 0x74, 0x6f, 0x70, 0x70, 0x65, 0x64, 0x5f, 0x61, 0x74, 0x18, 0x05,
	0x20, 0x01, 0x28, 0x03, 0x52, 0x09, 0x73, 0x74, 0x6f, 0x70, 0x70, 0x65, 0x64, 0x41, 0x74, 0x22,
	0x2c, 0x0a, 0x0a, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x54, 0x79, 0x70, 0x65, 0x12, 0x0e, 0x0a,
	0x0a, 0x44, 0x41, 0x54, 0x41, 0x53, 0x4f, 0x55, 0x52, 0x43, 0x45, 0x10, 0x00, 0x12, 0x0e, 0x0a,
	0x0a, 0x43, 0x4f, 0x4e, 0x54, 0x52, 0x4f, 0x4c, 0x4c, 0x45, 0x52, 0x10, 0x01, 0x22, 0x68, 0x0a,
	0x0b, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x53, 0x74, 0x61, 0x74, 0x65, 0x12, 0x09, 0x0a, 0x05,
	0x52, 0x45, 0x41, 0x44, 0x59, 0x10, 0x00, 0x12, 0x08, 0x0a, 0x04, 0x42, 0x55, 0x53, 0x59, 0x10,
	0x01, 0x12, 0x0c, 0x0a, 0x08, 0x53, 0x54, 0x4f, 0x50, 0x50, 0x49, 0x4e, 0x47, 0x10, 0x02, 0x12,
	0x0b, 0x0a, 0x07, 0x53, 0x54, 0x4f, 0x50, 0x50, 0x45, 0x44, 0x10, 0x03, 0x12, 0x0b, 0x0a, 0x07,
	0x55, 0x4e, 0x4b, 0x4e, 0x4f, 0x57, 0x4e, 0x10, 0x04, 0x12, 0x10, 0x0a, 0x0c, 0x55, 0x4e, 0x52,
	0x45, 0x53, 0x50, 0x4f, 0x4e, 0x53, 0x49, 0x56, 0x45, 0x10, 0x05, 0x12, 0x0a, 0x0a, 0x06, 0x4b,
	0x49, 0x4c, 0x4c, 0x45, 0x44, 0x10, 0x06, 0x22, 0x37, 0x0a, 0x0b, 0x50, 0x6c, 0x75, 0x67, 0x69,
	0x6e, 0x73, 0x4c, 0x69, 0x73, 0x74, 0x12, 0x28, 0x0a, 0x07, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e,
	0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x66, 0x6d, 0x74, 0x72, 0x70, 0x63,
	0x2e, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x52, 0x07, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73,
	0x22, 0x52, 0x0a, 0x17, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x50, 0x6c,
	0x75, 0x67, 0x69, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x6e,
	0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12,
	0x23, 0x0a, 0x05, 0x66, 0x72, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0d,
	0x2e, 0x66, 0x6d, 0x74, 0x72, 0x70, 0x63, 0x2e, 0x46, 0x72, 0x61, 0x6d, 0x65, 0x52, 0x05, 0x66,
	0x72, 0x61, 0x6d, 0x65, 0x22, 0x37, 0x0a, 0x10, 0x41, 0x64, 0x64, 0x50, 0x6c, 0x75, 0x67, 0x69,
	0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x23, 0x0a, 0x0d, 0x70, 0x6c, 0x75, 0x67,
	0x69, 0x6e, 0x5f, 0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x0c, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x32, 0x8d, 0x01,
	0x0a, 0x0a, 0x44, 0x61, 0x74, 0x61, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x12, 0x2d, 0x0a, 0x0b,
	0x53, 0x74, 0x61, 0x72, 0x74, 0x52, 0x65, 0x63, 0x6f, 0x72, 0x64, 0x12, 0x0d, 0x2e, 0x66, 0x6d,
	0x74, 0x72, 0x70, 0x63, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x0d, 0x2e, 0x66, 0x6d, 0x74,
	0x72, 0x70, 0x63, 0x2e, 0x46, 0x72, 0x61, 0x6d, 0x65, 0x30, 0x01, 0x12, 0x2a, 0x0a, 0x0a, 0x53,
	0x74, 0x6f, 0x70, 0x52, 0x65, 0x63, 0x6f, 0x72, 0x64, 0x12, 0x0d, 0x2e, 0x66, 0x6d, 0x74, 0x72,
	0x70, 0x63, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x0d, 0x2e, 0x66, 0x6d, 0x74, 0x72, 0x70,
	0x63, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x12, 0x24, 0x0a, 0x04, 0x53, 0x74, 0x6f, 0x70, 0x12,
	0x0d, 0x2e, 0x66, 0x6d, 0x74, 0x72, 0x70, 0x63, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x0d,
	0x2e, 0x66, 0x6d, 0x74, 0x72, 0x70, 0x63, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x32, 0x5d, 0x0a,
	0x0a, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x12, 0x24, 0x0a, 0x04, 0x53,
	0x74, 0x6f, 0x70, 0x12, 0x0d, 0x2e, 0x66, 0x6d, 0x74, 0x72, 0x70, 0x63, 0x2e, 0x45, 0x6d, 0x70,
	0x74, 0x79, 0x1a, 0x0d, 0x2e, 0x66, 0x6d, 0x74, 0x72, 0x70, 0x63, 0x2e, 0x45, 0x6d, 0x70, 0x74,
	0x79, 0x12, 0x29, 0x0a, 0x07, 0x43, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x12, 0x0d, 0x2e, 0x66,
	0x6d, 0x74, 0x72, 0x70, 0x63, 0x2e, 0x46, 0x72, 0x61, 0x6d, 0x65, 0x1a, 0x0d, 0x2e, 0x66, 0x6d,
	0x74, 0x72, 0x70, 0x63, 0x2e, 0x46, 0x72, 0x61, 0x6d, 0x65, 0x30, 0x01, 0x32, 0xb8, 0x03, 0x0a,
	0x09, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x41, 0x50, 0x49, 0x12, 0x33, 0x0a, 0x0b, 0x53, 0x74,
	0x61, 0x72, 0x74, 0x52, 0x65, 0x63, 0x6f, 0x72, 0x64, 0x12, 0x15, 0x2e, 0x66, 0x6d, 0x74, 0x72,
	0x70, 0x63, 0x2e, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x1a, 0x0d, 0x2e, 0x66, 0x6d, 0x74, 0x72, 0x70, 0x63, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x12,
	0x32, 0x0a, 0x0a, 0x53, 0x74, 0x6f, 0x70, 0x52, 0x65, 0x63, 0x6f, 0x72, 0x64, 0x12, 0x15, 0x2e,
	0x66, 0x6d, 0x74, 0x72, 0x70, 0x63, 0x2e, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x0d, 0x2e, 0x66, 0x6d, 0x74, 0x72, 0x70, 0x63, 0x2e, 0x45, 0x6d,
	0x70, 0x74, 0x79, 0x12, 0x33, 0x0a, 0x09, 0x53, 0x75, 0x62, 0x73, 0x63, 0x72, 0x69, 0x62, 0x65,
	0x12, 0x15, 0x2e, 0x66, 0x6d, 0x74, 0x72, 0x70, 0x63, 0x2e, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x0d, 0x2e, 0x66, 0x6d, 0x74, 0x72, 0x70, 0x63,
	0x2e, 0x46, 0x72, 0x61, 0x6d, 0x65, 0x30, 0x01, 0x12, 0x33, 0x0a, 0x0b, 0x53, 0x74, 0x61, 0x72,
	0x74, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x12, 0x15, 0x2e, 0x66, 0x6d, 0x74, 0x72, 0x70, 0x63,
	0x2e, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x0d,
	0x2e, 0x66, 0x6d, 0x74, 0x72, 0x70, 0x63, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x12, 0x32, 0x0a,
	0x0a, 0x53, 0x74, 0x6f, 0x70, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x12, 0x15, 0x2e, 0x66, 0x6d,
	0x74, 0x72, 0x70, 0x63, 0x2e, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x0d, 0x2e, 0x66, 0x6d, 0x74, 0x72, 0x70, 0x63, 0x2e, 0x45, 0x6d, 0x70, 0x74,
	0x79, 0x12, 0x3b, 0x0a, 0x07, 0x43, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x12, 0x1f, 0x2e, 0x66,
	0x6d, 0x74, 0x72, 0x70, 0x63, 0x2e, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72,
	0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x0d, 0x2e,
	0x66, 0x6d, 0x74, 0x72, 0x70, 0x63, 0x2e, 0x46, 0x72, 0x61, 0x6d, 0x65, 0x30, 0x01, 0x12, 0x31,
	0x0a, 0x0b, 0x4c, 0x69, 0x73, 0x74, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73, 0x12, 0x0d, 0x2e,
	0x66, 0x6d, 0x74, 0x72, 0x70, 0x63, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x13, 0x2e, 0x66,
	0x6d, 0x74, 0x72, 0x70, 0x63, 0x2e, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x73, 0x4c, 0x69, 0x73,
	0x74, 0x12, 0x34, 0x0a, 0x09, 0x41, 0x64, 0x64, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x12, 0x18,
	0x2e, 0x66, 0x6d, 0x74, 0x72, 0x70, 0x63, 0x2e, 0x41, 0x64, 0x64, 0x50, 0x6c, 0x75, 0x67, 0x69,
	0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x0d, 0x2e, 0x66, 0x6d, 0x74, 0x72, 0x70,
	0x63, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x42, 0x22, 0x5a, 0x20, 0x67, 0x69, 0x74, 0x68, 0x75,
	0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x53, 0x53, 0x53, 0x4f, 0x43, 0x2d, 0x43, 0x41, 0x4e, 0x2f,
	0x66, 0x6d, 0x74, 0x64, 0x2f, 0x66, 0x6d, 0x74, 0x72, 0x70, 0x63, 0x62, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x33,
}

var (
	file_plugin_proto_rawDescOnce sync.Once
	file_plugin_proto_rawDescData = file_plugin_proto_rawDesc
)

func file_plugin_proto_rawDescGZIP() []byte {
	file_plugin_proto_rawDescOnce.Do(func() {
		file_plugin_proto_rawDescData = protoimpl.X.CompressGZIP(file_plugin_proto_rawDescData)
	})
	return file_plugin_proto_rawDescData
}

var file_plugin_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_plugin_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_plugin_proto_goTypes = []interface{}{
	(Plugin_PluginType)(0),          // 0: fmtrpc.Plugin.PluginType
	(Plugin_PluginState)(0),         // 1: fmtrpc.Plugin.PluginState
	(*Empty)(nil),                   // 2: fmtrpc.Empty
	(*Frame)(nil),                   // 3: fmtrpc.Frame
	(*PluginRequest)(nil),           // 4: fmtrpc.PluginRequest
	(*Plugin)(nil),                  // 5: fmtrpc.Plugin
	(*PluginsList)(nil),             // 6: fmtrpc.PluginsList
	(*ControllerPluginRequest)(nil), // 7: fmtrpc.ControllerPluginRequest
	(*AddPluginRequest)(nil),        // 8: fmtrpc.AddPluginRequest
}
var file_plugin_proto_depIdxs = []int32{
	0,  // 0: fmtrpc.Plugin.type:type_name -> fmtrpc.Plugin.PluginType
	1,  // 1: fmtrpc.Plugin.state:type_name -> fmtrpc.Plugin.PluginState
	5,  // 2: fmtrpc.PluginsList.plugins:type_name -> fmtrpc.Plugin
	3,  // 3: fmtrpc.ControllerPluginRequest.frame:type_name -> fmtrpc.Frame
	2,  // 4: fmtrpc.Datasource.StartRecord:input_type -> fmtrpc.Empty
	2,  // 5: fmtrpc.Datasource.StopRecord:input_type -> fmtrpc.Empty
	2,  // 6: fmtrpc.Datasource.Stop:input_type -> fmtrpc.Empty
	2,  // 7: fmtrpc.Controller.Stop:input_type -> fmtrpc.Empty
	3,  // 8: fmtrpc.Controller.Command:input_type -> fmtrpc.Frame
	4,  // 9: fmtrpc.PluginAPI.StartRecord:input_type -> fmtrpc.PluginRequest
	4,  // 10: fmtrpc.PluginAPI.StopRecord:input_type -> fmtrpc.PluginRequest
	4,  // 11: fmtrpc.PluginAPI.Subscribe:input_type -> fmtrpc.PluginRequest
	4,  // 12: fmtrpc.PluginAPI.StartPlugin:input_type -> fmtrpc.PluginRequest
	4,  // 13: fmtrpc.PluginAPI.StopPlugin:input_type -> fmtrpc.PluginRequest
	7,  // 14: fmtrpc.PluginAPI.Command:input_type -> fmtrpc.ControllerPluginRequest
	2,  // 15: fmtrpc.PluginAPI.ListPlugins:input_type -> fmtrpc.Empty
	8,  // 16: fmtrpc.PluginAPI.AddPlugin:input_type -> fmtrpc.AddPluginRequest
	3,  // 17: fmtrpc.Datasource.StartRecord:output_type -> fmtrpc.Frame
	2,  // 18: fmtrpc.Datasource.StopRecord:output_type -> fmtrpc.Empty
	2,  // 19: fmtrpc.Datasource.Stop:output_type -> fmtrpc.Empty
	2,  // 20: fmtrpc.Controller.Stop:output_type -> fmtrpc.Empty
	3,  // 21: fmtrpc.Controller.Command:output_type -> fmtrpc.Frame
	2,  // 22: fmtrpc.PluginAPI.StartRecord:output_type -> fmtrpc.Empty
	2,  // 23: fmtrpc.PluginAPI.StopRecord:output_type -> fmtrpc.Empty
	3,  // 24: fmtrpc.PluginAPI.Subscribe:output_type -> fmtrpc.Frame
	2,  // 25: fmtrpc.PluginAPI.StartPlugin:output_type -> fmtrpc.Empty
	2,  // 26: fmtrpc.PluginAPI.StopPlugin:output_type -> fmtrpc.Empty
	3,  // 27: fmtrpc.PluginAPI.Command:output_type -> fmtrpc.Frame
	6,  // 28: fmtrpc.PluginAPI.ListPlugins:output_type -> fmtrpc.PluginsList
	2,  // 29: fmtrpc.PluginAPI.AddPlugin:output_type -> fmtrpc.Empty
	17, // [17:30] is the sub-list for method output_type
	4,  // [4:17] is the sub-list for method input_type
	4,  // [4:4] is the sub-list for extension type_name
	4,  // [4:4] is the sub-list for extension extendee
	0,  // [0:4] is the sub-list for field type_name
}

func init() { file_plugin_proto_init() }
func file_plugin_proto_init() {
	if File_plugin_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_plugin_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Empty); i {
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
		file_plugin_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Frame); i {
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
		file_plugin_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PluginRequest); i {
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
		file_plugin_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Plugin); i {
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
		file_plugin_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PluginsList); i {
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
		file_plugin_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ControllerPluginRequest); i {
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
		file_plugin_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AddPluginRequest); i {
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
			RawDescriptor: file_plugin_proto_rawDesc,
			NumEnums:      2,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   3,
		},
		GoTypes:           file_plugin_proto_goTypes,
		DependencyIndexes: file_plugin_proto_depIdxs,
		EnumInfos:         file_plugin_proto_enumTypes,
		MessageInfos:      file_plugin_proto_msgTypes,
	}.Build()
	File_plugin_proto = out.File
	file_plugin_proto_rawDesc = nil
	file_plugin_proto_goTypes = nil
	file_plugin_proto_depIdxs = nil
}
