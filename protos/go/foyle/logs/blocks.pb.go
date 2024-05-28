// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        (unknown)
// source: foyle/logs/blocks.proto

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

// BlockLog is the log of what happened to a block. It includes information about how a block was generated (if it
// was generated by the AI) and how it was executed if it was.
type BlockLog struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	// gen_trace_id is the trace ID of the generation request
	GenTraceId string `protobuf:"bytes,2,opt,name=gen_trace_id,json=genTraceId,proto3" json:"gen_trace_id,omitempty"`
	// exec_trace_ids are the trace IDs of the execution requests
	// Doc is the doc that triggered the generated block
	ExecTraceIds []string `protobuf:"bytes,3,rep,name=exec_trace_ids,json=execTraceIds,proto3" json:"exec_trace_ids,omitempty"`
	// doc is the doc that triggered the generated block
	Doc *v1alpha1.Doc `protobuf:"bytes,4,opt,name=doc,proto3" json:"doc,omitempty"`
	// generatedBlock is the block generated by the AI
	GeneratedBlock *v1alpha1.Block `protobuf:"bytes,5,opt,name=generated_block,json=generatedBlock,proto3" json:"generated_block,omitempty"`
	// executed_block is the final block that was actually executed
	// nil if the block was not executed
	ExecutedBlock *v1alpha1.Block `protobuf:"bytes,6,opt,name=executed_block,json=executedBlock,proto3" json:"executed_block,omitempty"`
	// exit_code is the exit code of the executed block
	ExitCode int32 `protobuf:"varint,7,opt,name=exit_code,json=exitCode,proto3" json:"exit_code,omitempty"`
	// eval_mode is true if the block was generated as part of an evaluation and shouldn't be used for learning
	EvalMode bool `protobuf:"varint,8,opt,name=eval_mode,json=evalMode,proto3" json:"eval_mode,omitempty"`
}

func (x *BlockLog) Reset() {
	*x = BlockLog{}
	if protoimpl.UnsafeEnabled {
		mi := &file_foyle_logs_blocks_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BlockLog) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BlockLog) ProtoMessage() {}

func (x *BlockLog) ProtoReflect() protoreflect.Message {
	mi := &file_foyle_logs_blocks_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BlockLog.ProtoReflect.Descriptor instead.
func (*BlockLog) Descriptor() ([]byte, []int) {
	return file_foyle_logs_blocks_proto_rawDescGZIP(), []int{0}
}

func (x *BlockLog) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *BlockLog) GetGenTraceId() string {
	if x != nil {
		return x.GenTraceId
	}
	return ""
}

func (x *BlockLog) GetExecTraceIds() []string {
	if x != nil {
		return x.ExecTraceIds
	}
	return nil
}

func (x *BlockLog) GetDoc() *v1alpha1.Doc {
	if x != nil {
		return x.Doc
	}
	return nil
}

func (x *BlockLog) GetGeneratedBlock() *v1alpha1.Block {
	if x != nil {
		return x.GeneratedBlock
	}
	return nil
}

func (x *BlockLog) GetExecutedBlock() *v1alpha1.Block {
	if x != nil {
		return x.ExecutedBlock
	}
	return nil
}

func (x *BlockLog) GetExitCode() int32 {
	if x != nil {
		return x.ExitCode
	}
	return 0
}

func (x *BlockLog) GetEvalMode() bool {
	if x != nil {
		return x.EvalMode
	}
	return false
}

var File_foyle_logs_blocks_proto protoreflect.FileDescriptor

var file_foyle_logs_blocks_proto_rawDesc = []byte{
	0x0a, 0x17, 0x66, 0x6f, 0x79, 0x6c, 0x65, 0x2f, 0x6c, 0x6f, 0x67, 0x73, 0x2f, 0x62, 0x6c, 0x6f,
	0x63, 0x6b, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0a, 0x66, 0x6f, 0x79, 0x6c, 0x65,
	0x2e, 0x6c, 0x6f, 0x67, 0x73, 0x1a, 0x1a, 0x66, 0x6f, 0x79, 0x6c, 0x65, 0x2f, 0x76, 0x31, 0x61,
	0x6c, 0x70, 0x68, 0x61, 0x31, 0x2f, 0x61, 0x67, 0x65, 0x6e, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x1a, 0x18, 0x66, 0x6f, 0x79, 0x6c, 0x65, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61,
	0x31, 0x2f, 0x64, 0x6f, 0x63, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1c, 0x67, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x73, 0x74, 0x72,
	0x75, 0x63, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73,
	0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x94, 0x02, 0x0a, 0x08, 0x42,
	0x6c, 0x6f, 0x63, 0x6b, 0x4c, 0x6f, 0x67, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x20, 0x0a, 0x0c, 0x67, 0x65, 0x6e, 0x5f, 0x74,
	0x72, 0x61, 0x63, 0x65, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x67,
	0x65, 0x6e, 0x54, 0x72, 0x61, 0x63, 0x65, 0x49, 0x64, 0x12, 0x24, 0x0a, 0x0e, 0x65, 0x78, 0x65,
	0x63, 0x5f, 0x74, 0x72, 0x61, 0x63, 0x65, 0x5f, 0x69, 0x64, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28,
	0x09, 0x52, 0x0c, 0x65, 0x78, 0x65, 0x63, 0x54, 0x72, 0x61, 0x63, 0x65, 0x49, 0x64, 0x73, 0x12,
	0x16, 0x0a, 0x03, 0x64, 0x6f, 0x63, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x04, 0x2e, 0x44,
	0x6f, 0x63, 0x52, 0x03, 0x64, 0x6f, 0x63, 0x12, 0x2f, 0x0a, 0x0f, 0x67, 0x65, 0x6e, 0x65, 0x72,
	0x61, 0x74, 0x65, 0x64, 0x5f, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x06, 0x2e, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x52, 0x0e, 0x67, 0x65, 0x6e, 0x65, 0x72, 0x61,
	0x74, 0x65, 0x64, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x12, 0x2d, 0x0a, 0x0e, 0x65, 0x78, 0x65, 0x63,
	0x75, 0x74, 0x65, 0x64, 0x5f, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x06, 0x2e, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x52, 0x0d, 0x65, 0x78, 0x65, 0x63, 0x75, 0x74,
	0x65, 0x64, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x12, 0x1b, 0x0a, 0x09, 0x65, 0x78, 0x69, 0x74, 0x5f,
	0x63, 0x6f, 0x64, 0x65, 0x18, 0x07, 0x20, 0x01, 0x28, 0x05, 0x52, 0x08, 0x65, 0x78, 0x69, 0x74,
	0x43, 0x6f, 0x64, 0x65, 0x12, 0x1b, 0x0a, 0x09, 0x65, 0x76, 0x61, 0x6c, 0x5f, 0x6d, 0x6f, 0x64,
	0x65, 0x18, 0x08, 0x20, 0x01, 0x28, 0x08, 0x52, 0x08, 0x65, 0x76, 0x61, 0x6c, 0x4d, 0x6f, 0x64,
	0x65, 0x42, 0x34, 0x5a, 0x32, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f,
	0x6a, 0x6c, 0x65, 0x77, 0x69, 0x2f, 0x66, 0x6f, 0x79, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x73, 0x2f, 0x67, 0x6f, 0x2f, 0x66, 0x6f, 0x79, 0x6c, 0x65, 0x2f, 0x6c, 0x6f, 0x67, 0x73,
	0x3b, 0x6c, 0x6f, 0x67, 0x73, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_foyle_logs_blocks_proto_rawDescOnce sync.Once
	file_foyle_logs_blocks_proto_rawDescData = file_foyle_logs_blocks_proto_rawDesc
)

func file_foyle_logs_blocks_proto_rawDescGZIP() []byte {
	file_foyle_logs_blocks_proto_rawDescOnce.Do(func() {
		file_foyle_logs_blocks_proto_rawDescData = protoimpl.X.CompressGZIP(file_foyle_logs_blocks_proto_rawDescData)
	})
	return file_foyle_logs_blocks_proto_rawDescData
}

var file_foyle_logs_blocks_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_foyle_logs_blocks_proto_goTypes = []interface{}{
	(*BlockLog)(nil),       // 0: foyle.logs.BlockLog
	(*v1alpha1.Doc)(nil),   // 1: Doc
	(*v1alpha1.Block)(nil), // 2: Block
}
var file_foyle_logs_blocks_proto_depIdxs = []int32{
	1, // 0: foyle.logs.BlockLog.doc:type_name -> Doc
	2, // 1: foyle.logs.BlockLog.generated_block:type_name -> Block
	2, // 2: foyle.logs.BlockLog.executed_block:type_name -> Block
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_foyle_logs_blocks_proto_init() }
func file_foyle_logs_blocks_proto_init() {
	if File_foyle_logs_blocks_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_foyle_logs_blocks_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BlockLog); i {
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
			RawDescriptor: file_foyle_logs_blocks_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_foyle_logs_blocks_proto_goTypes,
		DependencyIndexes: file_foyle_logs_blocks_proto_depIdxs,
		MessageInfos:      file_foyle_logs_blocks_proto_msgTypes,
	}.Build()
	File_foyle_logs_blocks_proto = out.File
	file_foyle_logs_blocks_proto_rawDesc = nil
	file_foyle_logs_blocks_proto_goTypes = nil
	file_foyle_logs_blocks_proto_depIdxs = nil
}
