//
//Author: Paul Côté
//Last Change Author: Paul Côté
//Last Date Changed: 2022/06/10
//
//Copyright (C) 2015-2018 Lightning Labs and The Lightning Network Developers
//
//Permission is hereby granted, free of charge, to any person obtaining a copy
//of this software and associated documentation files (the "Software"), to deal
//in the Software without restriction, including without limitation the rights
//to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
//copies of the Software, and to permit persons to whom the Software is
//furnished to do so, subject to the following conditions:
//
//The above copyright notice and this permission notice shall be included in
//all copies or substantial portions of the Software.
//
//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
//IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
//AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
//LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
//OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
//THE SOFTWARE.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.19.4
// source: fmt.proto

package fmtrpc

import (
	proto "github.com/SSSOC-CAN/laniakea-plugin-sdk/proto"
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

type TimeoutType int32

const (
	TimeoutType_SECOND TimeoutType = 0
	TimeoutType_MINUTE TimeoutType = 1
	TimeoutType_HOUR   TimeoutType = 2
	TimeoutType_DAY    TimeoutType = 3
)

// Enum value maps for TimeoutType.
var (
	TimeoutType_name = map[int32]string{
		0: "SECOND",
		1: "MINUTE",
		2: "HOUR",
		3: "DAY",
	}
	TimeoutType_value = map[string]int32{
		"SECOND": 0,
		"MINUTE": 1,
		"HOUR":   2,
		"DAY":    3,
	}
)

func (x TimeoutType) Enum() *TimeoutType {
	p := new(TimeoutType)
	*p = x
	return p
}

func (x TimeoutType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (TimeoutType) Descriptor() protoreflect.EnumDescriptor {
	return file_fmt_proto_enumTypes[0].Descriptor()
}

func (TimeoutType) Type() protoreflect.EnumType {
	return &file_fmt_proto_enumTypes[0]
}

func (x TimeoutType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use TimeoutType.Descriptor instead.
func (TimeoutType) EnumDescriptor() ([]byte, []int) {
	return file_fmt_proto_rawDescGZIP(), []int{0}
}

type StopRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *StopRequest) Reset() {
	*x = StopRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_fmt_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StopRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StopRequest) ProtoMessage() {}

func (x *StopRequest) ProtoReflect() protoreflect.Message {
	mi := &file_fmt_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StopRequest.ProtoReflect.Descriptor instead.
func (*StopRequest) Descriptor() ([]byte, []int) {
	return file_fmt_proto_rawDescGZIP(), []int{0}
}

type StopResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *StopResponse) Reset() {
	*x = StopResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_fmt_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StopResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StopResponse) ProtoMessage() {}

func (x *StopResponse) ProtoReflect() protoreflect.Message {
	mi := &file_fmt_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StopResponse.ProtoReflect.Descriptor instead.
func (*StopResponse) Descriptor() ([]byte, []int) {
	return file_fmt_proto_rawDescGZIP(), []int{1}
}

type AdminTestRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *AdminTestRequest) Reset() {
	*x = AdminTestRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_fmt_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AdminTestRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AdminTestRequest) ProtoMessage() {}

func (x *AdminTestRequest) ProtoReflect() protoreflect.Message {
	mi := &file_fmt_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AdminTestRequest.ProtoReflect.Descriptor instead.
func (*AdminTestRequest) Descriptor() ([]byte, []int) {
	return file_fmt_proto_rawDescGZIP(), []int{2}
}

type AdminTestResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// A short message indicating success or failure
	Msg string `protobuf:"bytes,1,opt,name=msg,proto3" json:"msg,omitempty"`
}

func (x *AdminTestResponse) Reset() {
	*x = AdminTestResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_fmt_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AdminTestResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AdminTestResponse) ProtoMessage() {}

func (x *AdminTestResponse) ProtoReflect() protoreflect.Message {
	mi := &file_fmt_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AdminTestResponse.ProtoReflect.Descriptor instead.
func (*AdminTestResponse) Descriptor() ([]byte, []int) {
	return file_fmt_proto_rawDescGZIP(), []int{3}
}

func (x *AdminTestResponse) GetMsg() string {
	if x != nil {
		return x.Msg
	}
	return ""
}

type TestRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *TestRequest) Reset() {
	*x = TestRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_fmt_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TestRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TestRequest) ProtoMessage() {}

func (x *TestRequest) ProtoReflect() protoreflect.Message {
	mi := &file_fmt_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TestRequest.ProtoReflect.Descriptor instead.
func (*TestRequest) Descriptor() ([]byte, []int) {
	return file_fmt_proto_rawDescGZIP(), []int{4}
}

type TestResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// A short message indicating success or failure
	Msg string `protobuf:"bytes,1,opt,name=msg,proto3" json:"msg,omitempty"`
}

func (x *TestResponse) Reset() {
	*x = TestResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_fmt_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TestResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TestResponse) ProtoMessage() {}

func (x *TestResponse) ProtoReflect() protoreflect.Message {
	mi := &file_fmt_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TestResponse.ProtoReflect.Descriptor instead.
func (*TestResponse) Descriptor() ([]byte, []int) {
	return file_fmt_proto_rawDescGZIP(), []int{5}
}

func (x *TestResponse) GetMsg() string {
	if x != nil {
		return x.Msg
	}
	return ""
}

type MacaroonPermission struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The entity a permission grants access to
	Entity string `protobuf:"bytes,1,opt,name=entity,proto3" json:"entity,omitempty"`
	// The action that is granted
	Action string `protobuf:"bytes,2,opt,name=action,proto3" json:"action,omitempty"`
}

func (x *MacaroonPermission) Reset() {
	*x = MacaroonPermission{}
	if protoimpl.UnsafeEnabled {
		mi := &file_fmt_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MacaroonPermission) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MacaroonPermission) ProtoMessage() {}

func (x *MacaroonPermission) ProtoReflect() protoreflect.Message {
	mi := &file_fmt_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MacaroonPermission.ProtoReflect.Descriptor instead.
func (*MacaroonPermission) Descriptor() ([]byte, []int) {
	return file_fmt_proto_rawDescGZIP(), []int{6}
}

func (x *MacaroonPermission) GetEntity() string {
	if x != nil {
		return x.Entity
	}
	return ""
}

func (x *MacaroonPermission) GetAction() string {
	if x != nil {
		return x.Action
	}
	return ""
}

type BakeMacaroonRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The length of time for which this macaroon is valid
	Timeout int64 `protobuf:"varint,1,opt,name=timeout,proto3" json:"timeout,omitempty"`
	// The unit for the macaroon timeout. Choose from `SECOND`, `MINUTE`, `HOUR` or `DAY`
	TimeoutType TimeoutType `protobuf:"varint,2,opt,name=timeout_type,json=timeoutType,proto3,enum=fmtrpc.TimeoutType" json:"timeout_type,omitempty"`
	// The list of permissions the new macaroon should grant
	Permissions []*MacaroonPermission `protobuf:"bytes,3,rep,name=permissions,proto3" json:"permissions,omitempty"`
	// The list of plugin names to be included in the macaroon
	Plugins []string `protobuf:"bytes,4,rep,name=plugins,proto3" json:"plugins,omitempty"`
}

func (x *BakeMacaroonRequest) Reset() {
	*x = BakeMacaroonRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_fmt_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BakeMacaroonRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BakeMacaroonRequest) ProtoMessage() {}

func (x *BakeMacaroonRequest) ProtoReflect() protoreflect.Message {
	mi := &file_fmt_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BakeMacaroonRequest.ProtoReflect.Descriptor instead.
func (*BakeMacaroonRequest) Descriptor() ([]byte, []int) {
	return file_fmt_proto_rawDescGZIP(), []int{7}
}

func (x *BakeMacaroonRequest) GetTimeout() int64 {
	if x != nil {
		return x.Timeout
	}
	return 0
}

func (x *BakeMacaroonRequest) GetTimeoutType() TimeoutType {
	if x != nil {
		return x.TimeoutType
	}
	return TimeoutType_SECOND
}

func (x *BakeMacaroonRequest) GetPermissions() []*MacaroonPermission {
	if x != nil {
		return x.Permissions
	}
	return nil
}

func (x *BakeMacaroonRequest) GetPlugins() []string {
	if x != nil {
		return x.Plugins
	}
	return nil
}

type BakeMacaroonResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The hex-encoded macaroon, serialized in binary format
	Macaroon string `protobuf:"bytes,1,opt,name=macaroon,proto3" json:"macaroon,omitempty"`
}

func (x *BakeMacaroonResponse) Reset() {
	*x = BakeMacaroonResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_fmt_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BakeMacaroonResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BakeMacaroonResponse) ProtoMessage() {}

func (x *BakeMacaroonResponse) ProtoReflect() protoreflect.Message {
	mi := &file_fmt_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BakeMacaroonResponse.ProtoReflect.Descriptor instead.
func (*BakeMacaroonResponse) Descriptor() ([]byte, []int) {
	return file_fmt_proto_rawDescGZIP(), []int{8}
}

func (x *BakeMacaroonResponse) GetMacaroon() string {
	if x != nil {
		return x.Macaroon
	}
	return ""
}

type DeprecationResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The message indicating the API endpoint has been deprecated
	Msg string `protobuf:"bytes,1,opt,name=msg,proto3" json:"msg,omitempty"`
}

func (x *DeprecationResponse) Reset() {
	*x = DeprecationResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_fmt_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DeprecationResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeprecationResponse) ProtoMessage() {}

func (x *DeprecationResponse) ProtoReflect() protoreflect.Message {
	mi := &file_fmt_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeprecationResponse.ProtoReflect.Descriptor instead.
func (*DeprecationResponse) Descriptor() ([]byte, []int) {
	return file_fmt_proto_rawDescGZIP(), []int{9}
}

func (x *DeprecationResponse) GetMsg() string {
	if x != nil {
		return x.Msg
	}
	return ""
}

var File_fmt_proto protoreflect.FileDescriptor

var file_fmt_proto_rawDesc = []byte{
	0x0a, 0x09, 0x66, 0x6d, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x06, 0x66, 0x6d, 0x74,
	0x72, 0x70, 0x63, 0x1a, 0x0c, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x22, 0x0d, 0x0a, 0x0b, 0x53, 0x74, 0x6f, 0x70, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x22, 0x0e, 0x0a, 0x0c, 0x53, 0x74, 0x6f, 0x70, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x22, 0x12, 0x0a, 0x10, 0x41, 0x64, 0x6d, 0x69, 0x6e, 0x54, 0x65, 0x73, 0x74, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x22, 0x25, 0x0a, 0x11, 0x41, 0x64, 0x6d, 0x69, 0x6e, 0x54, 0x65, 0x73,
	0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x6d, 0x73, 0x67,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6d, 0x73, 0x67, 0x22, 0x0d, 0x0a, 0x0b, 0x54,
	0x65, 0x73, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0x20, 0x0a, 0x0c, 0x54, 0x65,
	0x73, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x6d, 0x73,
	0x67, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6d, 0x73, 0x67, 0x22, 0x44, 0x0a, 0x12,
	0x4d, 0x61, 0x63, 0x61, 0x72, 0x6f, 0x6f, 0x6e, 0x50, 0x65, 0x72, 0x6d, 0x69, 0x73, 0x73, 0x69,
	0x6f, 0x6e, 0x12, 0x16, 0x0a, 0x06, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x06, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x12, 0x16, 0x0a, 0x06, 0x61, 0x63,
	0x74, 0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x61, 0x63, 0x74, 0x69,
	0x6f, 0x6e, 0x22, 0xbf, 0x01, 0x0a, 0x13, 0x42, 0x61, 0x6b, 0x65, 0x4d, 0x61, 0x63, 0x61, 0x72,
	0x6f, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x18, 0x0a, 0x07, 0x74, 0x69,
	0x6d, 0x65, 0x6f, 0x75, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x07, 0x74, 0x69, 0x6d,
	0x65, 0x6f, 0x75, 0x74, 0x12, 0x36, 0x0a, 0x0c, 0x74, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74, 0x5f,
	0x74, 0x79, 0x70, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x13, 0x2e, 0x66, 0x6d, 0x74,
	0x72, 0x70, 0x63, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74, 0x54, 0x79, 0x70, 0x65, 0x52,
	0x0b, 0x74, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74, 0x54, 0x79, 0x70, 0x65, 0x12, 0x3c, 0x0a, 0x0b,
	0x70, 0x65, 0x72, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x1a, 0x2e, 0x66, 0x6d, 0x74, 0x72, 0x70, 0x63, 0x2e, 0x4d, 0x61, 0x63, 0x61, 0x72,
	0x6f, 0x6f, 0x6e, 0x50, 0x65, 0x72, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x0b, 0x70,
	0x65, 0x72, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x18, 0x0a, 0x07, 0x70, 0x6c,
	0x75, 0x67, 0x69, 0x6e, 0x73, 0x18, 0x04, 0x20, 0x03, 0x28, 0x09, 0x52, 0x07, 0x70, 0x6c, 0x75,
	0x67, 0x69, 0x6e, 0x73, 0x22, 0x32, 0x0a, 0x14, 0x42, 0x61, 0x6b, 0x65, 0x4d, 0x61, 0x63, 0x61,
	0x72, 0x6f, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x1a, 0x0a, 0x08,
	0x6d, 0x61, 0x63, 0x61, 0x72, 0x6f, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08,
	0x6d, 0x61, 0x63, 0x61, 0x72, 0x6f, 0x6f, 0x6e, 0x22, 0x27, 0x0a, 0x13, 0x44, 0x65, 0x70, 0x72,
	0x65, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12,
	0x10, 0x0a, 0x03, 0x6d, 0x73, 0x67, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6d, 0x73,
	0x67, 0x2a, 0x38, 0x0a, 0x0b, 0x54, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74, 0x54, 0x79, 0x70, 0x65,
	0x12, 0x0a, 0x0a, 0x06, 0x53, 0x45, 0x43, 0x4f, 0x4e, 0x44, 0x10, 0x00, 0x12, 0x0a, 0x0a, 0x06,
	0x4d, 0x49, 0x4e, 0x55, 0x54, 0x45, 0x10, 0x01, 0x12, 0x08, 0x0a, 0x04, 0x48, 0x4f, 0x55, 0x52,
	0x10, 0x02, 0x12, 0x07, 0x0a, 0x03, 0x44, 0x41, 0x59, 0x10, 0x03, 0x32, 0xa7, 0x06, 0x0a, 0x03,
	0x46, 0x6d, 0x74, 0x12, 0x37, 0x0a, 0x0a, 0x53, 0x74, 0x6f, 0x70, 0x44, 0x61, 0x65, 0x6d, 0x6f,
	0x6e, 0x12, 0x13, 0x2e, 0x66, 0x6d, 0x74, 0x72, 0x70, 0x63, 0x2e, 0x53, 0x74, 0x6f, 0x70, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x14, 0x2e, 0x66, 0x6d, 0x74, 0x72, 0x70, 0x63, 0x2e,
	0x53, 0x74, 0x6f, 0x70, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x40, 0x0a, 0x09,
	0x41, 0x64, 0x6d, 0x69, 0x6e, 0x54, 0x65, 0x73, 0x74, 0x12, 0x18, 0x2e, 0x66, 0x6d, 0x74, 0x72,
	0x70, 0x63, 0x2e, 0x41, 0x64, 0x6d, 0x69, 0x6e, 0x54, 0x65, 0x73, 0x74, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x19, 0x2e, 0x66, 0x6d, 0x74, 0x72, 0x70, 0x63, 0x2e, 0x41, 0x64, 0x6d,
	0x69, 0x6e, 0x54, 0x65, 0x73, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x38,
	0x0a, 0x0b, 0x54, 0x65, 0x73, 0x74, 0x43, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x12, 0x13, 0x2e,
	0x66, 0x6d, 0x74, 0x72, 0x70, 0x63, 0x2e, 0x54, 0x65, 0x73, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x14, 0x2e, 0x66, 0x6d, 0x74, 0x72, 0x70, 0x63, 0x2e, 0x54, 0x65, 0x73, 0x74,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x49, 0x0a, 0x0c, 0x42, 0x61, 0x6b, 0x65,
	0x4d, 0x61, 0x63, 0x61, 0x72, 0x6f, 0x6f, 0x6e, 0x12, 0x1b, 0x2e, 0x66, 0x6d, 0x74, 0x72, 0x70,
	0x63, 0x2e, 0x42, 0x61, 0x6b, 0x65, 0x4d, 0x61, 0x63, 0x61, 0x72, 0x6f, 0x6f, 0x6e, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1c, 0x2e, 0x66, 0x6d, 0x74, 0x72, 0x70, 0x63, 0x2e, 0x42,
	0x61, 0x6b, 0x65, 0x4d, 0x61, 0x63, 0x61, 0x72, 0x6f, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x12, 0x3b, 0x0a, 0x0e, 0x53, 0x65, 0x74, 0x54, 0x65, 0x6d, 0x70, 0x65, 0x72,
	0x61, 0x74, 0x75, 0x72, 0x65, 0x12, 0x0c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x45, 0x6d,
	0x70, 0x74, 0x79, 0x1a, 0x1b, 0x2e, 0x66, 0x6d, 0x74, 0x72, 0x70, 0x63, 0x2e, 0x44, 0x65, 0x70,
	0x72, 0x65, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x38, 0x0a, 0x0b, 0x53, 0x65, 0x74, 0x50, 0x72, 0x65, 0x73, 0x73, 0x75, 0x72, 0x65, 0x12,
	0x0c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x1b, 0x2e,
	0x66, 0x6d, 0x74, 0x72, 0x70, 0x63, 0x2e, 0x44, 0x65, 0x70, 0x72, 0x65, 0x63, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x3b, 0x0a, 0x0e, 0x53, 0x74,
	0x61, 0x72, 0x74, 0x52, 0x65, 0x63, 0x6f, 0x72, 0x64, 0x69, 0x6e, 0x67, 0x12, 0x0c, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x1b, 0x2e, 0x66, 0x6d, 0x74,
	0x72, 0x70, 0x63, 0x2e, 0x44, 0x65, 0x70, 0x72, 0x65, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x3a, 0x0a, 0x0d, 0x53, 0x74, 0x6f, 0x70, 0x52,
	0x65, 0x63, 0x6f, 0x72, 0x64, 0x69, 0x6e, 0x67, 0x12, 0x0c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x1b, 0x2e, 0x66, 0x6d, 0x74, 0x72, 0x70, 0x63, 0x2e,
	0x44, 0x65, 0x70, 0x72, 0x65, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x12, 0x40, 0x0a, 0x13, 0x53, 0x75, 0x62, 0x73, 0x63, 0x72, 0x69, 0x62, 0x65,
	0x44, 0x61, 0x74, 0x61, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x12, 0x0c, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x1b, 0x2e, 0x66, 0x6d, 0x74, 0x72, 0x70,
	0x63, 0x2e, 0x44, 0x65, 0x70, 0x72, 0x65, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x39, 0x0a, 0x0c, 0x4c, 0x6f, 0x61, 0x64, 0x54, 0x65, 0x73,
	0x74, 0x50, 0x6c, 0x61, 0x6e, 0x12, 0x0c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x45, 0x6d,
	0x70, 0x74, 0x79, 0x1a, 0x1b, 0x2e, 0x66, 0x6d, 0x74, 0x72, 0x70, 0x63, 0x2e, 0x44, 0x65, 0x70,
	0x72, 0x65, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x3a, 0x0a, 0x0d, 0x53, 0x74, 0x61, 0x72, 0x74, 0x54, 0x65, 0x73, 0x74, 0x50, 0x6c, 0x61,
	0x6e, 0x12, 0x0c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a,
	0x1b, 0x2e, 0x66, 0x6d, 0x74, 0x72, 0x70, 0x63, 0x2e, 0x44, 0x65, 0x70, 0x72, 0x65, 0x63, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x39, 0x0a, 0x0c,
	0x53, 0x74, 0x6f, 0x70, 0x54, 0x65, 0x73, 0x74, 0x50, 0x6c, 0x61, 0x6e, 0x12, 0x0c, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x1b, 0x2e, 0x66, 0x6d, 0x74,
	0x72, 0x70, 0x63, 0x2e, 0x44, 0x65, 0x70, 0x72, 0x65, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x3c, 0x0a, 0x0f, 0x49, 0x6e, 0x73, 0x65, 0x72,
	0x74, 0x52, 0x4f, 0x49, 0x4d, 0x61, 0x72, 0x6b, 0x65, 0x72, 0x12, 0x0c, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x1b, 0x2e, 0x66, 0x6d, 0x74, 0x72, 0x70,
	0x63, 0x2e, 0x44, 0x65, 0x70, 0x72, 0x65, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42, 0x22, 0x5a, 0x20, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e,
	0x63, 0x6f, 0x6d, 0x2f, 0x53, 0x53, 0x53, 0x4f, 0x43, 0x2d, 0x43, 0x41, 0x4e, 0x2f, 0x66, 0x6d,
	0x74, 0x64, 0x2f, 0x66, 0x6d, 0x74, 0x72, 0x70, 0x63, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_fmt_proto_rawDescOnce sync.Once
	file_fmt_proto_rawDescData = file_fmt_proto_rawDesc
)

func file_fmt_proto_rawDescGZIP() []byte {
	file_fmt_proto_rawDescOnce.Do(func() {
		file_fmt_proto_rawDescData = protoimpl.X.CompressGZIP(file_fmt_proto_rawDescData)
	})
	return file_fmt_proto_rawDescData
}

var file_fmt_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_fmt_proto_msgTypes = make([]protoimpl.MessageInfo, 10)
var file_fmt_proto_goTypes = []interface{}{
	(TimeoutType)(0),             // 0: fmtrpc.TimeoutType
	(*StopRequest)(nil),          // 1: fmtrpc.StopRequest
	(*StopResponse)(nil),         // 2: fmtrpc.StopResponse
	(*AdminTestRequest)(nil),     // 3: fmtrpc.AdminTestRequest
	(*AdminTestResponse)(nil),    // 4: fmtrpc.AdminTestResponse
	(*TestRequest)(nil),          // 5: fmtrpc.TestRequest
	(*TestResponse)(nil),         // 6: fmtrpc.TestResponse
	(*MacaroonPermission)(nil),   // 7: fmtrpc.MacaroonPermission
	(*BakeMacaroonRequest)(nil),  // 8: fmtrpc.BakeMacaroonRequest
	(*BakeMacaroonResponse)(nil), // 9: fmtrpc.BakeMacaroonResponse
	(*DeprecationResponse)(nil),  // 10: fmtrpc.DeprecationResponse
	(*proto.Empty)(nil),          // 11: proto.Empty
}
var file_fmt_proto_depIdxs = []int32{
	0,  // 0: fmtrpc.BakeMacaroonRequest.timeout_type:type_name -> fmtrpc.TimeoutType
	7,  // 1: fmtrpc.BakeMacaroonRequest.permissions:type_name -> fmtrpc.MacaroonPermission
	1,  // 2: fmtrpc.Fmt.StopDaemon:input_type -> fmtrpc.StopRequest
	3,  // 3: fmtrpc.Fmt.AdminTest:input_type -> fmtrpc.AdminTestRequest
	5,  // 4: fmtrpc.Fmt.TestCommand:input_type -> fmtrpc.TestRequest
	8,  // 5: fmtrpc.Fmt.BakeMacaroon:input_type -> fmtrpc.BakeMacaroonRequest
	11, // 6: fmtrpc.Fmt.SetTemperature:input_type -> proto.Empty
	11, // 7: fmtrpc.Fmt.SetPressure:input_type -> proto.Empty
	11, // 8: fmtrpc.Fmt.StartRecording:input_type -> proto.Empty
	11, // 9: fmtrpc.Fmt.StopRecording:input_type -> proto.Empty
	11, // 10: fmtrpc.Fmt.SubscribeDataStream:input_type -> proto.Empty
	11, // 11: fmtrpc.Fmt.LoadTestPlan:input_type -> proto.Empty
	11, // 12: fmtrpc.Fmt.StartTestPlan:input_type -> proto.Empty
	11, // 13: fmtrpc.Fmt.StopTestPlan:input_type -> proto.Empty
	11, // 14: fmtrpc.Fmt.InsertROIMarker:input_type -> proto.Empty
	2,  // 15: fmtrpc.Fmt.StopDaemon:output_type -> fmtrpc.StopResponse
	4,  // 16: fmtrpc.Fmt.AdminTest:output_type -> fmtrpc.AdminTestResponse
	6,  // 17: fmtrpc.Fmt.TestCommand:output_type -> fmtrpc.TestResponse
	9,  // 18: fmtrpc.Fmt.BakeMacaroon:output_type -> fmtrpc.BakeMacaroonResponse
	10, // 19: fmtrpc.Fmt.SetTemperature:output_type -> fmtrpc.DeprecationResponse
	10, // 20: fmtrpc.Fmt.SetPressure:output_type -> fmtrpc.DeprecationResponse
	10, // 21: fmtrpc.Fmt.StartRecording:output_type -> fmtrpc.DeprecationResponse
	10, // 22: fmtrpc.Fmt.StopRecording:output_type -> fmtrpc.DeprecationResponse
	10, // 23: fmtrpc.Fmt.SubscribeDataStream:output_type -> fmtrpc.DeprecationResponse
	10, // 24: fmtrpc.Fmt.LoadTestPlan:output_type -> fmtrpc.DeprecationResponse
	10, // 25: fmtrpc.Fmt.StartTestPlan:output_type -> fmtrpc.DeprecationResponse
	10, // 26: fmtrpc.Fmt.StopTestPlan:output_type -> fmtrpc.DeprecationResponse
	10, // 27: fmtrpc.Fmt.InsertROIMarker:output_type -> fmtrpc.DeprecationResponse
	15, // [15:28] is the sub-list for method output_type
	2,  // [2:15] is the sub-list for method input_type
	2,  // [2:2] is the sub-list for extension type_name
	2,  // [2:2] is the sub-list for extension extendee
	0,  // [0:2] is the sub-list for field type_name
}

func init() { file_fmt_proto_init() }
func file_fmt_proto_init() {
	if File_fmt_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_fmt_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StopRequest); i {
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
		file_fmt_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StopResponse); i {
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
		file_fmt_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AdminTestRequest); i {
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
		file_fmt_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AdminTestResponse); i {
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
		file_fmt_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TestRequest); i {
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
		file_fmt_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TestResponse); i {
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
		file_fmt_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MacaroonPermission); i {
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
		file_fmt_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BakeMacaroonRequest); i {
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
		file_fmt_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BakeMacaroonResponse); i {
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
		file_fmt_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DeprecationResponse); i {
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
			RawDescriptor: file_fmt_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   10,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_fmt_proto_goTypes,
		DependencyIndexes: file_fmt_proto_depIdxs,
		EnumInfos:         file_fmt_proto_enumTypes,
		MessageInfos:      file_fmt_proto_msgTypes,
	}.Build()
	File_fmt_proto = out.File
	file_fmt_proto_rawDesc = nil
	file_fmt_proto_goTypes = nil
	file_fmt_proto_depIdxs = nil
}
