// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        (unknown)
// source: foyle/v1alpha1/trainer.proto

package v1alpha1

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	_ "google.golang.org/protobuf/types/known/structpb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// Example represents an example to be used in few shot learning
type Example struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id        string    `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Embedding []float32 `protobuf:"fixed32,2,rep,packed,name=embedding,proto3" json:"embedding,omitempty"`
	Query     *Doc      `protobuf:"bytes,3,opt,name=query,proto3" json:"query,omitempty"`
	Answer    []*Block  `protobuf:"bytes,4,rep,name=answer,proto3" json:"answer,omitempty"`
}

func (x *Example) Reset() {
	*x = Example{}
	if protoimpl.UnsafeEnabled {
		mi := &file_foyle_v1alpha1_trainer_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Example) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Example) ProtoMessage() {}

func (x *Example) ProtoReflect() protoreflect.Message {
	mi := &file_foyle_v1alpha1_trainer_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Example.ProtoReflect.Descriptor instead.
func (*Example) Descriptor() ([]byte, []int) {
	return file_foyle_v1alpha1_trainer_proto_rawDescGZIP(), []int{0}
}

func (x *Example) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Example) GetEmbedding() []float32 {
	if x != nil {
		return x.Embedding
	}
	return nil
}

func (x *Example) GetQuery() *Doc {
	if x != nil {
		return x.Query
	}
	return nil
}

func (x *Example) GetAnswer() []*Block {
	if x != nil {
		return x.Answer
	}
	return nil
}

var File_foyle_v1alpha1_trainer_proto protoreflect.FileDescriptor

var file_foyle_v1alpha1_trainer_proto_rawDesc = []byte{
	0x0a, 0x1c, 0x66, 0x6f, 0x79, 0x6c, 0x65, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31,
	0x2f, 0x74, 0x72, 0x61, 0x69, 0x6e, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x18,
	0x66, 0x6f, 0x79, 0x6c, 0x65, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2f, 0x64,
	0x6f, 0x63, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x73, 0x74, 0x72, 0x75, 0x63, 0x74,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x73, 0x0a, 0x07, 0x45, 0x78, 0x61, 0x6d, 0x70, 0x6c,
	0x65, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69,
	0x64, 0x12, 0x1c, 0x0a, 0x09, 0x65, 0x6d, 0x62, 0x65, 0x64, 0x64, 0x69, 0x6e, 0x67, 0x18, 0x02,
	0x20, 0x03, 0x28, 0x02, 0x52, 0x09, 0x65, 0x6d, 0x62, 0x65, 0x64, 0x64, 0x69, 0x6e, 0x67, 0x12,
	0x1a, 0x0a, 0x05, 0x71, 0x75, 0x65, 0x72, 0x79, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x04,
	0x2e, 0x44, 0x6f, 0x63, 0x52, 0x05, 0x71, 0x75, 0x65, 0x72, 0x79, 0x12, 0x1e, 0x0a, 0x06, 0x61,
	0x6e, 0x73, 0x77, 0x65, 0x72, 0x18, 0x04, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x06, 0x2e, 0x42, 0x6c,
	0x6f, 0x63, 0x6b, 0x52, 0x06, 0x61, 0x6e, 0x73, 0x77, 0x65, 0x72, 0x42, 0x2b, 0x5a, 0x29, 0x67,
	0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6a, 0x6c, 0x65, 0x77, 0x69, 0x2f,
	0x66, 0x6f, 0x79, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2f, 0x67, 0x6f, 0x2f,
	0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_foyle_v1alpha1_trainer_proto_rawDescOnce sync.Once
	file_foyle_v1alpha1_trainer_proto_rawDescData = file_foyle_v1alpha1_trainer_proto_rawDesc
)

func file_foyle_v1alpha1_trainer_proto_rawDescGZIP() []byte {
	file_foyle_v1alpha1_trainer_proto_rawDescOnce.Do(func() {
		file_foyle_v1alpha1_trainer_proto_rawDescData = protoimpl.X.CompressGZIP(file_foyle_v1alpha1_trainer_proto_rawDescData)
	})
	return file_foyle_v1alpha1_trainer_proto_rawDescData
}

var file_foyle_v1alpha1_trainer_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_foyle_v1alpha1_trainer_proto_goTypes = []interface{}{
	(*Example)(nil), // 0: Example
	(*Doc)(nil),     // 1: Doc
	(*Block)(nil),   // 2: Block
}
var file_foyle_v1alpha1_trainer_proto_depIdxs = []int32{
	1, // 0: Example.query:type_name -> Doc
	2, // 1: Example.answer:type_name -> Block
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_foyle_v1alpha1_trainer_proto_init() }
func file_foyle_v1alpha1_trainer_proto_init() {
	if File_foyle_v1alpha1_trainer_proto != nil {
		return
	}
	file_foyle_v1alpha1_doc_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_foyle_v1alpha1_trainer_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Example); i {
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
			RawDescriptor: file_foyle_v1alpha1_trainer_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_foyle_v1alpha1_trainer_proto_goTypes,
		DependencyIndexes: file_foyle_v1alpha1_trainer_proto_depIdxs,
		MessageInfos:      file_foyle_v1alpha1_trainer_proto_msgTypes,
	}.Build()
	File_foyle_v1alpha1_trainer_proto = out.File
	file_foyle_v1alpha1_trainer_proto_rawDesc = nil
	file_foyle_v1alpha1_trainer_proto_goTypes = nil
	file_foyle_v1alpha1_trainer_proto_depIdxs = nil
}
