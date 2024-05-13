// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        (unknown)
// source: foyle/runme/ai.proto

package runme

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

type RunmeGenerateRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Notebook *Notebook `protobuf:"bytes,1,opt,name=notebook,proto3" json:"notebook,omitempty"`
}

func (x *RunmeGenerateRequest) Reset() {
	*x = RunmeGenerateRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_foyle_runme_ai_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RunmeGenerateRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RunmeGenerateRequest) ProtoMessage() {}

func (x *RunmeGenerateRequest) ProtoReflect() protoreflect.Message {
	mi := &file_foyle_runme_ai_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RunmeGenerateRequest.ProtoReflect.Descriptor instead.
func (*RunmeGenerateRequest) Descriptor() ([]byte, []int) {
	return file_foyle_runme_ai_proto_rawDescGZIP(), []int{0}
}

func (x *RunmeGenerateRequest) GetNotebook() *Notebook {
	if x != nil {
		return x.Notebook
	}
	return nil
}

type RunmeGenerateResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Cells []*Cell `protobuf:"bytes,1,rep,name=cells,proto3" json:"cells,omitempty"`
}

func (x *RunmeGenerateResponse) Reset() {
	*x = RunmeGenerateResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_foyle_runme_ai_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RunmeGenerateResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RunmeGenerateResponse) ProtoMessage() {}

func (x *RunmeGenerateResponse) ProtoReflect() protoreflect.Message {
	mi := &file_foyle_runme_ai_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RunmeGenerateResponse.ProtoReflect.Descriptor instead.
func (*RunmeGenerateResponse) Descriptor() ([]byte, []int) {
	return file_foyle_runme_ai_proto_rawDescGZIP(), []int{1}
}

func (x *RunmeGenerateResponse) GetCells() []*Cell {
	if x != nil {
		return x.Cells
	}
	return nil
}

var File_foyle_runme_ai_proto protoreflect.FileDescriptor

var file_foyle_runme_ai_proto_rawDesc = []byte{
	0x0a, 0x14, 0x66, 0x6f, 0x79, 0x6c, 0x65, 0x2f, 0x72, 0x75, 0x6e, 0x6d, 0x65, 0x2f, 0x61, 0x69,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0b, 0x66, 0x6f, 0x79, 0x6c, 0x65, 0x2e, 0x72, 0x75,
	0x6e, 0x6d, 0x65, 0x1a, 0x18, 0x66, 0x6f, 0x79, 0x6c, 0x65, 0x2f, 0x72, 0x75, 0x6e, 0x6d, 0x65,
	0x2f, 0x70, 0x61, 0x72, 0x73, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x49, 0x0a,
	0x14, 0x52, 0x75, 0x6e, 0x6d, 0x65, 0x47, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x65, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x31, 0x0a, 0x08, 0x6e, 0x6f, 0x74, 0x65, 0x62, 0x6f, 0x6f,
	0x6b, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x66, 0x6f, 0x79, 0x6c, 0x65, 0x2e,
	0x72, 0x75, 0x6e, 0x6d, 0x65, 0x2e, 0x4e, 0x6f, 0x74, 0x65, 0x62, 0x6f, 0x6f, 0x6b, 0x52, 0x08,
	0x6e, 0x6f, 0x74, 0x65, 0x62, 0x6f, 0x6f, 0x6b, 0x22, 0x40, 0x0a, 0x15, 0x52, 0x75, 0x6e, 0x6d,
	0x65, 0x47, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x12, 0x27, 0x0a, 0x05, 0x63, 0x65, 0x6c, 0x6c, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x11, 0x2e, 0x66, 0x6f, 0x79, 0x6c, 0x65, 0x2e, 0x72, 0x75, 0x6e, 0x6d, 0x65, 0x2e, 0x43,
	0x65, 0x6c, 0x6c, 0x52, 0x05, 0x63, 0x65, 0x6c, 0x6c, 0x73, 0x32, 0x6b, 0x0a, 0x14, 0x52, 0x75,
	0x6e, 0x6d, 0x65, 0x47, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x65, 0x53, 0x65, 0x72, 0x76, 0x69,
	0x63, 0x65, 0x12, 0x53, 0x0a, 0x08, 0x47, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x65, 0x12, 0x21,
	0x2e, 0x66, 0x6f, 0x79, 0x6c, 0x65, 0x2e, 0x72, 0x75, 0x6e, 0x6d, 0x65, 0x2e, 0x52, 0x75, 0x6e,
	0x6d, 0x65, 0x47, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x22, 0x2e, 0x66, 0x6f, 0x79, 0x6c, 0x65, 0x2e, 0x72, 0x75, 0x6e, 0x6d, 0x65, 0x2e,
	0x52, 0x75, 0x6e, 0x6d, 0x65, 0x47, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x65, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x42, 0x2e, 0x5a, 0x2c, 0x67, 0x69, 0x74, 0x68, 0x75,
	0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6a, 0x6c, 0x65, 0x77, 0x69, 0x2f, 0x66, 0x6f, 0x79, 0x6c,
	0x65, 0x2f, 0x72, 0x75, 0x6e, 0x6d, 0x65, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x66, 0x6f, 0x79, 0x6c,
	0x65, 0x2f, 0x72, 0x75, 0x6e, 0x6d, 0x65, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_foyle_runme_ai_proto_rawDescOnce sync.Once
	file_foyle_runme_ai_proto_rawDescData = file_foyle_runme_ai_proto_rawDesc
)

func file_foyle_runme_ai_proto_rawDescGZIP() []byte {
	file_foyle_runme_ai_proto_rawDescOnce.Do(func() {
		file_foyle_runme_ai_proto_rawDescData = protoimpl.X.CompressGZIP(file_foyle_runme_ai_proto_rawDescData)
	})
	return file_foyle_runme_ai_proto_rawDescData
}

var file_foyle_runme_ai_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_foyle_runme_ai_proto_goTypes = []interface{}{
	(*RunmeGenerateRequest)(nil),  // 0: foyle.runme.RunmeGenerateRequest
	(*RunmeGenerateResponse)(nil), // 1: foyle.runme.RunmeGenerateResponse
	(*Notebook)(nil),              // 2: foyle.runme.Notebook
	(*Cell)(nil),                  // 3: foyle.runme.Cell
}
var file_foyle_runme_ai_proto_depIdxs = []int32{
	2, // 0: foyle.runme.RunmeGenerateRequest.notebook:type_name -> foyle.runme.Notebook
	3, // 1: foyle.runme.RunmeGenerateResponse.cells:type_name -> foyle.runme.Cell
	0, // 2: foyle.runme.RunmeGenerateService.Generate:input_type -> foyle.runme.RunmeGenerateRequest
	1, // 3: foyle.runme.RunmeGenerateService.Generate:output_type -> foyle.runme.RunmeGenerateResponse
	3, // [3:4] is the sub-list for method output_type
	2, // [2:3] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_foyle_runme_ai_proto_init() }
func file_foyle_runme_ai_proto_init() {
	if File_foyle_runme_ai_proto != nil {
		return
	}
	file_foyle_runme_parser_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_foyle_runme_ai_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RunmeGenerateRequest); i {
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
		file_foyle_runme_ai_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RunmeGenerateResponse); i {
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
			RawDescriptor: file_foyle_runme_ai_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_foyle_runme_ai_proto_goTypes,
		DependencyIndexes: file_foyle_runme_ai_proto_depIdxs,
		MessageInfos:      file_foyle_runme_ai_proto_msgTypes,
	}.Build()
	File_foyle_runme_ai_proto = out.File
	file_foyle_runme_ai_proto_rawDesc = nil
	file_foyle_runme_ai_proto_goTypes = nil
	file_foyle_runme_ai_proto_depIdxs = nil
}