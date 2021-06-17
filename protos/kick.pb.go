// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.15.6
// source: kick.proto

package protos

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

type KickMsg struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	UserId string `protobuf:"bytes,1,opt,name=userId,proto3" json:"userId,omitempty"`
}

func (x *KickMsg) Reset() {
	*x = KickMsg{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kick_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *KickMsg) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*KickMsg) ProtoMessage() {}

func (x *KickMsg) ProtoReflect() protoreflect.Message {
	mi := &file_kick_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use KickMsg.ProtoReflect.Descriptor instead.
func (*KickMsg) Descriptor() ([]byte, []int) {
	return file_kick_proto_rawDescGZIP(), []int{0}
}

func (x *KickMsg) GetUserId() string {
	if x != nil {
		return x.UserId
	}
	return ""
}

type KickAnswer struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Kicked bool `protobuf:"varint,1,opt,name=kicked,proto3" json:"kicked,omitempty"`
}

func (x *KickAnswer) Reset() {
	*x = KickAnswer{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kick_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *KickAnswer) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*KickAnswer) ProtoMessage() {}

func (x *KickAnswer) ProtoReflect() protoreflect.Message {
	mi := &file_kick_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use KickAnswer.ProtoReflect.Descriptor instead.
func (*KickAnswer) Descriptor() ([]byte, []int) {
	return file_kick_proto_rawDescGZIP(), []int{1}
}

func (x *KickAnswer) GetKicked() bool {
	if x != nil {
		return x.Kicked
	}
	return false
}

var File_kick_proto protoreflect.FileDescriptor

var file_kick_proto_rawDesc = []byte{
	0x0a, 0x0a, 0x6b, 0x69, 0x63, 0x6b, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x73, 0x22, 0x21, 0x0a, 0x07, 0x4b, 0x69, 0x63, 0x6b, 0x4d, 0x73, 0x67, 0x12,
	0x16, 0x0a, 0x06, 0x75, 0x73, 0x65, 0x72, 0x49, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x06, 0x75, 0x73, 0x65, 0x72, 0x49, 0x64, 0x22, 0x24, 0x0a, 0x0a, 0x4b, 0x69, 0x63, 0x6b, 0x41,
	0x6e, 0x73, 0x77, 0x65, 0x72, 0x12, 0x16, 0x0a, 0x06, 0x6b, 0x69, 0x63, 0x6b, 0x65, 0x64, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x06, 0x6b, 0x69, 0x63, 0x6b, 0x65, 0x64, 0x42, 0x1b, 0x5a,
	0x08, 0x2e, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0xaa, 0x02, 0x0e, 0x4e, 0x50, 0x69, 0x74,
	0x61, 0x79, 0x61, 0x2e, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
}

var (
	file_kick_proto_rawDescOnce sync.Once
	file_kick_proto_rawDescData = file_kick_proto_rawDesc
)

func file_kick_proto_rawDescGZIP() []byte {
	file_kick_proto_rawDescOnce.Do(func() {
		file_kick_proto_rawDescData = protoimpl.X.CompressGZIP(file_kick_proto_rawDescData)
	})
	return file_kick_proto_rawDescData
}

var file_kick_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_kick_proto_goTypes = []interface{}{
	(*KickMsg)(nil),    // 0: protos.KickMsg
	(*KickAnswer)(nil), // 1: protos.KickAnswer
}
var file_kick_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_kick_proto_init() }
func file_kick_proto_init() {
	if File_kick_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_kick_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*KickMsg); i {
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
		file_kick_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*KickAnswer); i {
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
			RawDescriptor: file_kick_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_kick_proto_goTypes,
		DependencyIndexes: file_kick_proto_depIdxs,
		MessageInfos:      file_kick_proto_msgTypes,
	}.Build()
	File_kick_proto = out.File
	file_kick_proto_rawDesc = nil
	file_kick_proto_goTypes = nil
	file_kick_proto_depIdxs = nil
}
