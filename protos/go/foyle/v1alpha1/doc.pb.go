// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        (unknown)
// source: foyle/v1alpha1/doc.proto

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

type BlockKind int32

const (
	BlockKind_UNKNOWN_BLOCK_KIND BlockKind = 0
	BlockKind_MARKUP             BlockKind = 1
	BlockKind_CODE               BlockKind = 2
)

// Enum value maps for BlockKind.
var (
	BlockKind_name = map[int32]string{
		0: "UNKNOWN_BLOCK_KIND",
		1: "MARKUP",
		2: "CODE",
	}
	BlockKind_value = map[string]int32{
		"UNKNOWN_BLOCK_KIND": 0,
		"MARKUP":             1,
		"CODE":               2,
	}
)

func (x BlockKind) Enum() *BlockKind {
	p := new(BlockKind)
	*p = x
	return p
}

func (x BlockKind) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (BlockKind) Descriptor() protoreflect.EnumDescriptor {
	return file_foyle_v1alpha1_doc_proto_enumTypes[0].Descriptor()
}

func (BlockKind) Type() protoreflect.EnumType {
	return &file_foyle_v1alpha1_doc_proto_enumTypes[0]
}

func (x BlockKind) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use BlockKind.Descriptor instead.
func (BlockKind) EnumDescriptor() ([]byte, []int) {
	return file_foyle_v1alpha1_doc_proto_rawDescGZIP(), []int{0}
}

// Doc represents a document in the editor.
type Doc struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Blocks []*Block `protobuf:"bytes,1,rep,name=blocks,proto3" json:"blocks,omitempty"`
}

func (x *Doc) Reset() {
	*x = Doc{}
	if protoimpl.UnsafeEnabled {
		mi := &file_foyle_v1alpha1_doc_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Doc) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Doc) ProtoMessage() {}

func (x *Doc) ProtoReflect() protoreflect.Message {
	mi := &file_foyle_v1alpha1_doc_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Doc.ProtoReflect.Descriptor instead.
func (*Doc) Descriptor() ([]byte, []int) {
	return file_foyle_v1alpha1_doc_proto_rawDescGZIP(), []int{0}
}

func (x *Doc) GetBlocks() []*Block {
	if x != nil {
		return x.Blocks
	}
	return nil
}

// Block represents a block in a document.
// It is inspired by the VSCode NotebookCellData type
// https://github.com/microsoft/vscode/blob/98332892fd2cb3c948ced33f542698e20c6279b9/src/vs/workbench/api/common/extHostTypes.ts#L3598
type Block struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// BlockKind is an enum indicating what type of block it is e.g text or output
	// It maps to VSCode's NotebookCellKind
	// https://github.com/microsoft/vscode/blob/98332892fd2cb3c948ced33f542698e20c6279b9/src/vs/workbench/api/common/extHostTypes.ts#L3766
	Kind BlockKind `protobuf:"varint,1,opt,name=kind,proto3,enum=BlockKind" json:"kind,omitempty"`
	// language is a string identifying the language.
	// It maps to languageId
	// https://github.com/microsoft/vscode/blob/98332892fd2cb3c948ced33f542698e20c6279b9/src/vs/workbench/api/common/extHostTypes.ts#L3623
	Language string `protobuf:"bytes,2,opt,name=language,proto3" json:"language,omitempty"`
	// contents is the actual contents of the block.
	// Not the outputs of the block.
	// It corresponds to the value in NotebookCellData
	Contents string `protobuf:"bytes,3,opt,name=contents,proto3" json:"contents,omitempty"`
	// outputs are the output of a block if any.
	Outputs []*BlockOutput `protobuf:"bytes,4,rep,name=outputs,proto3" json:"outputs,omitempty"`
	// IDs of any traces associated with this block.
	// TODO(jeremy): Can we deprecate this field? The trace is a property of the request not the individual block.
	TraceIds []string `protobuf:"bytes,6,rep,name=trace_ids,json=traceIds,proto3" json:"trace_ids,omitempty"`
	// ID of the block.
	Id string `protobuf:"bytes,7,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *Block) Reset() {
	*x = Block{}
	if protoimpl.UnsafeEnabled {
		mi := &file_foyle_v1alpha1_doc_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Block) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Block) ProtoMessage() {}

func (x *Block) ProtoReflect() protoreflect.Message {
	mi := &file_foyle_v1alpha1_doc_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Block.ProtoReflect.Descriptor instead.
func (*Block) Descriptor() ([]byte, []int) {
	return file_foyle_v1alpha1_doc_proto_rawDescGZIP(), []int{1}
}

func (x *Block) GetKind() BlockKind {
	if x != nil {
		return x.Kind
	}
	return BlockKind_UNKNOWN_BLOCK_KIND
}

func (x *Block) GetLanguage() string {
	if x != nil {
		return x.Language
	}
	return ""
}

func (x *Block) GetContents() string {
	if x != nil {
		return x.Contents
	}
	return ""
}

func (x *Block) GetOutputs() []*BlockOutput {
	if x != nil {
		return x.Outputs
	}
	return nil
}

func (x *Block) GetTraceIds() []string {
	if x != nil {
		return x.TraceIds
	}
	return nil
}

func (x *Block) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

// BlockOutput represents the output of a block.
// It corresponds to a VSCode NotebookCellOutput
// https://github.com/microsoft/vscode/blob/98332892fd2cb3c948ced33f542698e20c6279b9/src/vscode-dts/vscode.d.ts#L14835
type BlockOutput struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// items is the output items. Each item is the different representation of the same output data
	Items []*BlockOutputItem `protobuf:"bytes,1,rep,name=items,proto3" json:"items,omitempty"`
}

func (x *BlockOutput) Reset() {
	*x = BlockOutput{}
	if protoimpl.UnsafeEnabled {
		mi := &file_foyle_v1alpha1_doc_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BlockOutput) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BlockOutput) ProtoMessage() {}

func (x *BlockOutput) ProtoReflect() protoreflect.Message {
	mi := &file_foyle_v1alpha1_doc_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BlockOutput.ProtoReflect.Descriptor instead.
func (*BlockOutput) Descriptor() ([]byte, []int) {
	return file_foyle_v1alpha1_doc_proto_rawDescGZIP(), []int{2}
}

func (x *BlockOutput) GetItems() []*BlockOutputItem {
	if x != nil {
		return x.Items
	}
	return nil
}

// BlockOutputItem represents an item in a block output.
// It corresponds to a VSCode NotebookCellOutputItem
// https://github.com/microsoft/vscode/blob/98332892fd2cb3c948ced33f542698e20c6279b9/src/vscode-dts/vscode.d.ts#L14753
type BlockOutputItem struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// mime is the mime type of the output item.
	Mime string `protobuf:"bytes,1,opt,name=mime,proto3" json:"mime,omitempty"`
	// value of the output item.
	// We use string data type and not bytes because the JSON representation of bytes is a base64 string.
	// vscode data uses a byte. We may need to add support for bytes to support non text data data in the future.
	TextData string `protobuf:"bytes,2,opt,name=text_data,json=textData,proto3" json:"text_data,omitempty"`
}

func (x *BlockOutputItem) Reset() {
	*x = BlockOutputItem{}
	if protoimpl.UnsafeEnabled {
		mi := &file_foyle_v1alpha1_doc_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BlockOutputItem) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BlockOutputItem) ProtoMessage() {}

func (x *BlockOutputItem) ProtoReflect() protoreflect.Message {
	mi := &file_foyle_v1alpha1_doc_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BlockOutputItem.ProtoReflect.Descriptor instead.
func (*BlockOutputItem) Descriptor() ([]byte, []int) {
	return file_foyle_v1alpha1_doc_proto_rawDescGZIP(), []int{3}
}

func (x *BlockOutputItem) GetMime() string {
	if x != nil {
		return x.Mime
	}
	return ""
}

func (x *BlockOutputItem) GetTextData() string {
	if x != nil {
		return x.TextData
	}
	return ""
}

var File_foyle_v1alpha1_doc_proto protoreflect.FileDescriptor

var file_foyle_v1alpha1_doc_proto_rawDesc = []byte{
	0x0a, 0x18, 0x66, 0x6f, 0x79, 0x6c, 0x65, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31,
	0x2f, 0x64, 0x6f, 0x63, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x73, 0x74, 0x72, 0x75,
	0x63, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x25, 0x0a, 0x03, 0x44, 0x6f, 0x63, 0x12,
	0x1e, 0x0a, 0x06, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x06, 0x2e, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x52, 0x06, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x73, 0x22,
	0xb4, 0x01, 0x0a, 0x05, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x12, 0x1e, 0x0a, 0x04, 0x6b, 0x69, 0x6e,
	0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x0a, 0x2e, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x4b,
	0x69, 0x6e, 0x64, 0x52, 0x04, 0x6b, 0x69, 0x6e, 0x64, 0x12, 0x1a, 0x0a, 0x08, 0x6c, 0x61, 0x6e,
	0x67, 0x75, 0x61, 0x67, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x6c, 0x61, 0x6e,
	0x67, 0x75, 0x61, 0x67, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74,
	0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74,
	0x73, 0x12, 0x26, 0x0a, 0x07, 0x6f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x73, 0x18, 0x04, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x0c, 0x2e, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x4f, 0x75, 0x74, 0x70, 0x75, 0x74,
	0x52, 0x07, 0x6f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x73, 0x12, 0x1b, 0x0a, 0x09, 0x74, 0x72, 0x61,
	0x63, 0x65, 0x5f, 0x69, 0x64, 0x73, 0x18, 0x06, 0x20, 0x03, 0x28, 0x09, 0x52, 0x08, 0x74, 0x72,
	0x61, 0x63, 0x65, 0x49, 0x64, 0x73, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x07, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x22, 0x35, 0x0a, 0x0b, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x4f,
	0x75, 0x74, 0x70, 0x75, 0x74, 0x12, 0x26, 0x0a, 0x05, 0x69, 0x74, 0x65, 0x6d, 0x73, 0x18, 0x01,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x4f, 0x75, 0x74, 0x70,
	0x75, 0x74, 0x49, 0x74, 0x65, 0x6d, 0x52, 0x05, 0x69, 0x74, 0x65, 0x6d, 0x73, 0x22, 0x42, 0x0a,
	0x0f, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x4f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x49, 0x74, 0x65, 0x6d,
	0x12, 0x12, 0x0a, 0x04, 0x6d, 0x69, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04,
	0x6d, 0x69, 0x6d, 0x65, 0x12, 0x1b, 0x0a, 0x09, 0x74, 0x65, 0x78, 0x74, 0x5f, 0x64, 0x61, 0x74,
	0x61, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x74, 0x65, 0x78, 0x74, 0x44, 0x61, 0x74,
	0x61, 0x2a, 0x39, 0x0a, 0x09, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x4b, 0x69, 0x6e, 0x64, 0x12, 0x16,
	0x0a, 0x12, 0x55, 0x4e, 0x4b, 0x4e, 0x4f, 0x57, 0x4e, 0x5f, 0x42, 0x4c, 0x4f, 0x43, 0x4b, 0x5f,
	0x4b, 0x49, 0x4e, 0x44, 0x10, 0x00, 0x12, 0x0a, 0x0a, 0x06, 0x4d, 0x41, 0x52, 0x4b, 0x55, 0x50,
	0x10, 0x01, 0x12, 0x08, 0x0a, 0x04, 0x43, 0x4f, 0x44, 0x45, 0x10, 0x02, 0x42, 0x3d, 0x42, 0x08,
	0x44, 0x6f, 0x63, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x2f, 0x67, 0x69, 0x74, 0x68,
	0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6a, 0x6c, 0x65, 0x77, 0x69, 0x2f, 0x66, 0x6f, 0x79,
	0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2f, 0x67, 0x6f, 0x2f, 0x66, 0x6f, 0x79,
	0x6c, 0x65, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x33,
}

var (
	file_foyle_v1alpha1_doc_proto_rawDescOnce sync.Once
	file_foyle_v1alpha1_doc_proto_rawDescData = file_foyle_v1alpha1_doc_proto_rawDesc
)

func file_foyle_v1alpha1_doc_proto_rawDescGZIP() []byte {
	file_foyle_v1alpha1_doc_proto_rawDescOnce.Do(func() {
		file_foyle_v1alpha1_doc_proto_rawDescData = protoimpl.X.CompressGZIP(file_foyle_v1alpha1_doc_proto_rawDescData)
	})
	return file_foyle_v1alpha1_doc_proto_rawDescData
}

var file_foyle_v1alpha1_doc_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_foyle_v1alpha1_doc_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_foyle_v1alpha1_doc_proto_goTypes = []interface{}{
	(BlockKind)(0),          // 0: BlockKind
	(*Doc)(nil),             // 1: Doc
	(*Block)(nil),           // 2: Block
	(*BlockOutput)(nil),     // 3: BlockOutput
	(*BlockOutputItem)(nil), // 4: BlockOutputItem
}
var file_foyle_v1alpha1_doc_proto_depIdxs = []int32{
	2, // 0: Doc.blocks:type_name -> Block
	0, // 1: Block.kind:type_name -> BlockKind
	3, // 2: Block.outputs:type_name -> BlockOutput
	4, // 3: BlockOutput.items:type_name -> BlockOutputItem
	4, // [4:4] is the sub-list for method output_type
	4, // [4:4] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_foyle_v1alpha1_doc_proto_init() }
func file_foyle_v1alpha1_doc_proto_init() {
	if File_foyle_v1alpha1_doc_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_foyle_v1alpha1_doc_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Doc); i {
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
		file_foyle_v1alpha1_doc_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Block); i {
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
		file_foyle_v1alpha1_doc_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BlockOutput); i {
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
		file_foyle_v1alpha1_doc_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BlockOutputItem); i {
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
			RawDescriptor: file_foyle_v1alpha1_doc_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_foyle_v1alpha1_doc_proto_goTypes,
		DependencyIndexes: file_foyle_v1alpha1_doc_proto_depIdxs,
		EnumInfos:         file_foyle_v1alpha1_doc_proto_enumTypes,
		MessageInfos:      file_foyle_v1alpha1_doc_proto_msgTypes,
	}.Build()
	File_foyle_v1alpha1_doc_proto = out.File
	file_foyle_v1alpha1_doc_proto_rawDesc = nil
	file_foyle_v1alpha1_doc_proto_goTypes = nil
	file_foyle_v1alpha1_doc_proto_depIdxs = nil
}
