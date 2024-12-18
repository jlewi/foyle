// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.2
// 	protoc        (unknown)
// source: foyle/logs/conversion.proto

package logspb

import (
	v1alpha1 "github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	_ "google.golang.org/protobuf/types/known/structpb"
	_ "google.golang.org/protobuf/types/known/timestamppb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type ConvertDocRequest_Format int32

const (
	ConvertDocRequest_UNKNOWN ConvertDocRequest_Format = 0
	// Convert to markdown
	ConvertDocRequest_MARKDOWN ConvertDocRequest_Format = 1
	// Convert to HTML
	ConvertDocRequest_HTML ConvertDocRequest_Format = 2
)

// Enum value maps for ConvertDocRequest_Format.
var (
	ConvertDocRequest_Format_name = map[int32]string{
		0: "UNKNOWN",
		1: "MARKDOWN",
		2: "HTML",
	}
	ConvertDocRequest_Format_value = map[string]int32{
		"UNKNOWN":  0,
		"MARKDOWN": 1,
		"HTML":     2,
	}
)

func (x ConvertDocRequest_Format) Enum() *ConvertDocRequest_Format {
	p := new(ConvertDocRequest_Format)
	*p = x
	return p
}

func (x ConvertDocRequest_Format) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ConvertDocRequest_Format) Descriptor() protoreflect.EnumDescriptor {
	return file_foyle_logs_conversion_proto_enumTypes[0].Descriptor()
}

func (ConvertDocRequest_Format) Type() protoreflect.EnumType {
	return &file_foyle_logs_conversion_proto_enumTypes[0]
}

func (x ConvertDocRequest_Format) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ConvertDocRequest_Format.Descriptor instead.
func (ConvertDocRequest_Format) EnumDescriptor() ([]byte, []int) {
	return file_foyle_logs_conversion_proto_rawDescGZIP(), []int{0, 0}
}

type ConvertDocRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The doc to convert
	Doc    *v1alpha1.Doc            `protobuf:"bytes,1,opt,name=doc,proto3" json:"doc,omitempty"`
	Format ConvertDocRequest_Format `protobuf:"varint,2,opt,name=format,proto3,enum=foyle.logs.ConvertDocRequest_Format" json:"format,omitempty"`
}

func (x *ConvertDocRequest) Reset() {
	*x = ConvertDocRequest{}
	mi := &file_foyle_logs_conversion_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ConvertDocRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ConvertDocRequest) ProtoMessage() {}

func (x *ConvertDocRequest) ProtoReflect() protoreflect.Message {
	mi := &file_foyle_logs_conversion_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ConvertDocRequest.ProtoReflect.Descriptor instead.
func (*ConvertDocRequest) Descriptor() ([]byte, []int) {
	return file_foyle_logs_conversion_proto_rawDescGZIP(), []int{0}
}

func (x *ConvertDocRequest) GetDoc() *v1alpha1.Doc {
	if x != nil {
		return x.Doc
	}
	return nil
}

func (x *ConvertDocRequest) GetFormat() ConvertDocRequest_Format {
	if x != nil {
		return x.Format
	}
	return ConvertDocRequest_UNKNOWN
}

type ConvertDocResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The converted doc
	Text string `protobuf:"bytes,1,opt,name=text,proto3" json:"text,omitempty"`
}

func (x *ConvertDocResponse) Reset() {
	*x = ConvertDocResponse{}
	mi := &file_foyle_logs_conversion_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ConvertDocResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ConvertDocResponse) ProtoMessage() {}

func (x *ConvertDocResponse) ProtoReflect() protoreflect.Message {
	mi := &file_foyle_logs_conversion_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ConvertDocResponse.ProtoReflect.Descriptor instead.
func (*ConvertDocResponse) Descriptor() ([]byte, []int) {
	return file_foyle_logs_conversion_proto_rawDescGZIP(), []int{1}
}

func (x *ConvertDocResponse) GetText() string {
	if x != nil {
		return x.Text
	}
	return ""
}

var File_foyle_logs_conversion_proto protoreflect.FileDescriptor

var file_foyle_logs_conversion_proto_rawDesc = []byte{
	0x0a, 0x1b, 0x66, 0x6f, 0x79, 0x6c, 0x65, 0x2f, 0x6c, 0x6f, 0x67, 0x73, 0x2f, 0x63, 0x6f, 0x6e,
	0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0a, 0x66,
	0x6f, 0x79, 0x6c, 0x65, 0x2e, 0x6c, 0x6f, 0x67, 0x73, 0x1a, 0x1a, 0x66, 0x6f, 0x79, 0x6c, 0x65,
	0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2f, 0x61, 0x67, 0x65, 0x6e, 0x74, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x18, 0x66, 0x6f, 0x79, 0x6c, 0x65, 0x2f, 0x76, 0x31, 0x61,
	0x6c, 0x70, 0x68, 0x61, 0x31, 0x2f, 0x64, 0x6f, 0x63, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a,
	0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2f, 0x73, 0x74, 0x72, 0x75, 0x63, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1f, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74,
	0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x98,
	0x01, 0x0a, 0x11, 0x43, 0x6f, 0x6e, 0x76, 0x65, 0x72, 0x74, 0x44, 0x6f, 0x63, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x12, 0x16, 0x0a, 0x03, 0x64, 0x6f, 0x63, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x04, 0x2e, 0x44, 0x6f, 0x63, 0x52, 0x03, 0x64, 0x6f, 0x63, 0x12, 0x3c, 0x0a, 0x06,
	0x66, 0x6f, 0x72, 0x6d, 0x61, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x24, 0x2e, 0x66,
	0x6f, 0x79, 0x6c, 0x65, 0x2e, 0x6c, 0x6f, 0x67, 0x73, 0x2e, 0x43, 0x6f, 0x6e, 0x76, 0x65, 0x72,
	0x74, 0x44, 0x6f, 0x63, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x2e, 0x46, 0x6f, 0x72, 0x6d,
	0x61, 0x74, 0x52, 0x06, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x74, 0x22, 0x2d, 0x0a, 0x06, 0x46, 0x6f,
	0x72, 0x6d, 0x61, 0x74, 0x12, 0x0b, 0x0a, 0x07, 0x55, 0x4e, 0x4b, 0x4e, 0x4f, 0x57, 0x4e, 0x10,
	0x00, 0x12, 0x0c, 0x0a, 0x08, 0x4d, 0x41, 0x52, 0x4b, 0x44, 0x4f, 0x57, 0x4e, 0x10, 0x01, 0x12,
	0x08, 0x0a, 0x04, 0x48, 0x54, 0x4d, 0x4c, 0x10, 0x02, 0x22, 0x28, 0x0a, 0x12, 0x43, 0x6f, 0x6e,
	0x76, 0x65, 0x72, 0x74, 0x44, 0x6f, 0x63, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12,
	0x12, 0x0a, 0x04, 0x74, 0x65, 0x78, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x74,
	0x65, 0x78, 0x74, 0x32, 0x62, 0x0a, 0x11, 0x43, 0x6f, 0x6e, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f,
	0x6e, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x4d, 0x0a, 0x0a, 0x43, 0x6f, 0x6e, 0x76,
	0x65, 0x72, 0x74, 0x44, 0x6f, 0x63, 0x12, 0x1d, 0x2e, 0x66, 0x6f, 0x79, 0x6c, 0x65, 0x2e, 0x6c,
	0x6f, 0x67, 0x73, 0x2e, 0x43, 0x6f, 0x6e, 0x76, 0x65, 0x72, 0x74, 0x44, 0x6f, 0x63, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1e, 0x2e, 0x66, 0x6f, 0x79, 0x6c, 0x65, 0x2e, 0x6c, 0x6f,
	0x67, 0x73, 0x2e, 0x43, 0x6f, 0x6e, 0x76, 0x65, 0x72, 0x74, 0x44, 0x6f, 0x63, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x42, 0x9e, 0x01, 0x0a, 0x0e, 0x63, 0x6f, 0x6d, 0x2e,
	0x66, 0x6f, 0x79, 0x6c, 0x65, 0x2e, 0x6c, 0x6f, 0x67, 0x73, 0x42, 0x0f, 0x43, 0x6f, 0x6e, 0x76,
	0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x32, 0x67,
	0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6a, 0x6c, 0x65, 0x77, 0x69, 0x2f,
	0x66, 0x6f, 0x79, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2f, 0x67, 0x6f, 0x2f,
	0x66, 0x6f, 0x79, 0x6c, 0x65, 0x2f, 0x6c, 0x6f, 0x67, 0x73, 0x3b, 0x6c, 0x6f, 0x67, 0x73, 0x70,
	0x62, 0xa2, 0x02, 0x03, 0x46, 0x4c, 0x58, 0xaa, 0x02, 0x0a, 0x46, 0x6f, 0x79, 0x6c, 0x65, 0x2e,
	0x4c, 0x6f, 0x67, 0x73, 0xca, 0x02, 0x0a, 0x46, 0x6f, 0x79, 0x6c, 0x65, 0x5c, 0x4c, 0x6f, 0x67,
	0x73, 0xe2, 0x02, 0x16, 0x46, 0x6f, 0x79, 0x6c, 0x65, 0x5c, 0x4c, 0x6f, 0x67, 0x73, 0x5c, 0x47,
	0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02, 0x0b, 0x46, 0x6f, 0x79,
	0x6c, 0x65, 0x3a, 0x3a, 0x4c, 0x6f, 0x67, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_foyle_logs_conversion_proto_rawDescOnce sync.Once
	file_foyle_logs_conversion_proto_rawDescData = file_foyle_logs_conversion_proto_rawDesc
)

func file_foyle_logs_conversion_proto_rawDescGZIP() []byte {
	file_foyle_logs_conversion_proto_rawDescOnce.Do(func() {
		file_foyle_logs_conversion_proto_rawDescData = protoimpl.X.CompressGZIP(file_foyle_logs_conversion_proto_rawDescData)
	})
	return file_foyle_logs_conversion_proto_rawDescData
}

var file_foyle_logs_conversion_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_foyle_logs_conversion_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_foyle_logs_conversion_proto_goTypes = []any{
	(ConvertDocRequest_Format)(0), // 0: foyle.logs.ConvertDocRequest.Format
	(*ConvertDocRequest)(nil),     // 1: foyle.logs.ConvertDocRequest
	(*ConvertDocResponse)(nil),    // 2: foyle.logs.ConvertDocResponse
	(*v1alpha1.Doc)(nil),          // 3: Doc
}
var file_foyle_logs_conversion_proto_depIdxs = []int32{
	3, // 0: foyle.logs.ConvertDocRequest.doc:type_name -> Doc
	0, // 1: foyle.logs.ConvertDocRequest.format:type_name -> foyle.logs.ConvertDocRequest.Format
	1, // 2: foyle.logs.ConversionService.ConvertDoc:input_type -> foyle.logs.ConvertDocRequest
	2, // 3: foyle.logs.ConversionService.ConvertDoc:output_type -> foyle.logs.ConvertDocResponse
	3, // [3:4] is the sub-list for method output_type
	2, // [2:3] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_foyle_logs_conversion_proto_init() }
func file_foyle_logs_conversion_proto_init() {
	if File_foyle_logs_conversion_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_foyle_logs_conversion_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_foyle_logs_conversion_proto_goTypes,
		DependencyIndexes: file_foyle_logs_conversion_proto_depIdxs,
		EnumInfos:         file_foyle_logs_conversion_proto_enumTypes,
		MessageInfos:      file_foyle_logs_conversion_proto_msgTypes,
	}.Build()
	File_foyle_logs_conversion_proto = out.File
	file_foyle_logs_conversion_proto_rawDesc = nil
	file_foyle_logs_conversion_proto_goTypes = nil
	file_foyle_logs_conversion_proto_depIdxs = nil
}
