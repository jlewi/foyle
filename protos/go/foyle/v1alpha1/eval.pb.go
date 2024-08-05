// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        (unknown)
// source: foyle/v1alpha1/eval.proto

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

type EvalResultStatus int32

const (
	EvalResultStatus_UNKNOWN_EVAL_RESULT_STATUS EvalResultStatus = 0
	EvalResultStatus_DONE                       EvalResultStatus = 1
	EvalResultStatus_ERROR                      EvalResultStatus = 2
)

// Enum value maps for EvalResultStatus.
var (
	EvalResultStatus_name = map[int32]string{
		0: "UNKNOWN_EVAL_RESULT_STATUS",
		1: "DONE",
		2: "ERROR",
	}
	EvalResultStatus_value = map[string]int32{
		"UNKNOWN_EVAL_RESULT_STATUS": 0,
		"DONE":                       1,
		"ERROR":                      2,
	}
)

func (x EvalResultStatus) Enum() *EvalResultStatus {
	p := new(EvalResultStatus)
	*p = x
	return p
}

func (x EvalResultStatus) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (EvalResultStatus) Descriptor() protoreflect.EnumDescriptor {
	return file_foyle_v1alpha1_eval_proto_enumTypes[0].Descriptor()
}

func (EvalResultStatus) Type() protoreflect.EnumType {
	return &file_foyle_v1alpha1_eval_proto_enumTypes[0]
}

func (x EvalResultStatus) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use EvalResultStatus.Descriptor instead.
func (EvalResultStatus) EnumDescriptor() ([]byte, []int) {
	return file_foyle_v1alpha1_eval_proto_rawDescGZIP(), []int{0}
}

type AssertResult int32

const (
	AssertResult_UNKNOWN_AssertResult AssertResult = 0
	AssertResult_PASSED               AssertResult = 1
	AssertResult_FAILED               AssertResult = 2
	AssertResult_SKIPPED              AssertResult = 3
)

// Enum value maps for AssertResult.
var (
	AssertResult_name = map[int32]string{
		0: "UNKNOWN_AssertResult",
		1: "PASSED",
		2: "FAILED",
		3: "SKIPPED",
	}
	AssertResult_value = map[string]int32{
		"UNKNOWN_AssertResult": 0,
		"PASSED":               1,
		"FAILED":               2,
		"SKIPPED":              3,
	}
)

func (x AssertResult) Enum() *AssertResult {
	p := new(AssertResult)
	*p = x
	return p
}

func (x AssertResult) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (AssertResult) Descriptor() protoreflect.EnumDescriptor {
	return file_foyle_v1alpha1_eval_proto_enumTypes[1].Descriptor()
}

func (AssertResult) Type() protoreflect.EnumType {
	return &file_foyle_v1alpha1_eval_proto_enumTypes[1]
}

func (x AssertResult) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use AssertResult.Descriptor instead.
func (AssertResult) EnumDescriptor() ([]byte, []int) {
	return file_foyle_v1alpha1_eval_proto_rawDescGZIP(), []int{1}
}

// EvalResult represents an evaluation result
type EvalResult struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Example is the answer and expected result
	Example *Example `protobuf:"bytes,1,opt,name=example,proto3" json:"example,omitempty"`
	// example_file is the file containing the example
	ExampleFile string `protobuf:"bytes,2,opt,name=example_file,json=exampleFile,proto3" json:"example_file,omitempty"`
	// Actual response
	Actual []*Block `protobuf:"bytes,3,rep,name=actual,proto3" json:"actual,omitempty"`
	// The distance between the actual and expected response
	Distance           int32   `protobuf:"varint,4,opt,name=distance,proto3" json:"distance,omitempty"`
	NormalizedDistance float32 `protobuf:"fixed32,7,opt,name=normalized_distance,json=normalizedDistance,proto3" json:"normalized_distance,omitempty"`
	Error              string  `protobuf:"bytes,5,opt,name=error,proto3" json:"error,omitempty"`
	// Status of the evaluation
	Status EvalResultStatus `protobuf:"varint,6,opt,name=status,proto3,enum=EvalResultStatus" json:"status,omitempty"`
	// The ID of the generate trace
	GenTraceId string `protobuf:"bytes,8,opt,name=gen_trace_id,json=genTraceId,proto3" json:"gen_trace_id,omitempty"`
	// Best matching RAG result
	BestRagResult *RAGResult   `protobuf:"bytes,9,opt,name=best_rag_result,json=bestRagResult,proto3" json:"best_rag_result,omitempty"`
	Assertions    []*Assertion `protobuf:"bytes,10,rep,name=assertions,proto3" json:"assertions,omitempty"`
}

func (x *EvalResult) Reset() {
	*x = EvalResult{}
	if protoimpl.UnsafeEnabled {
		mi := &file_foyle_v1alpha1_eval_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EvalResult) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EvalResult) ProtoMessage() {}

func (x *EvalResult) ProtoReflect() protoreflect.Message {
	mi := &file_foyle_v1alpha1_eval_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EvalResult.ProtoReflect.Descriptor instead.
func (*EvalResult) Descriptor() ([]byte, []int) {
	return file_foyle_v1alpha1_eval_proto_rawDescGZIP(), []int{0}
}

func (x *EvalResult) GetExample() *Example {
	if x != nil {
		return x.Example
	}
	return nil
}

func (x *EvalResult) GetExampleFile() string {
	if x != nil {
		return x.ExampleFile
	}
	return ""
}

func (x *EvalResult) GetActual() []*Block {
	if x != nil {
		return x.Actual
	}
	return nil
}

func (x *EvalResult) GetDistance() int32 {
	if x != nil {
		return x.Distance
	}
	return 0
}

func (x *EvalResult) GetNormalizedDistance() float32 {
	if x != nil {
		return x.NormalizedDistance
	}
	return 0
}

func (x *EvalResult) GetError() string {
	if x != nil {
		return x.Error
	}
	return ""
}

func (x *EvalResult) GetStatus() EvalResultStatus {
	if x != nil {
		return x.Status
	}
	return EvalResultStatus_UNKNOWN_EVAL_RESULT_STATUS
}

func (x *EvalResult) GetGenTraceId() string {
	if x != nil {
		return x.GenTraceId
	}
	return ""
}

func (x *EvalResult) GetBestRagResult() *RAGResult {
	if x != nil {
		return x.BestRagResult
	}
	return nil
}

func (x *EvalResult) GetAssertions() []*Assertion {
	if x != nil {
		return x.Assertions
	}
	return nil
}

type Assertion struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Name of the assertion
	Name   string       `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Result AssertResult `protobuf:"varint,2,opt,name=result,proto3,enum=AssertResult" json:"result,omitempty"`
	// Human readable detail of the assertion. If there was an error this should contain the error message.
	Detail string `protobuf:"bytes,3,opt,name=detail,proto3" json:"detail,omitempty"`
}

func (x *Assertion) Reset() {
	*x = Assertion{}
	if protoimpl.UnsafeEnabled {
		mi := &file_foyle_v1alpha1_eval_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Assertion) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Assertion) ProtoMessage() {}

func (x *Assertion) ProtoReflect() protoreflect.Message {
	mi := &file_foyle_v1alpha1_eval_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Assertion.ProtoReflect.Descriptor instead.
func (*Assertion) Descriptor() ([]byte, []int) {
	return file_foyle_v1alpha1_eval_proto_rawDescGZIP(), []int{1}
}

func (x *Assertion) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Assertion) GetResult() AssertResult {
	if x != nil {
		return x.Result
	}
	return AssertResult_UNKNOWN_AssertResult
}

func (x *Assertion) GetDetail() string {
	if x != nil {
		return x.Detail
	}
	return ""
}

type EvalResultListRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The path of the database to fetch results for
	Database string `protobuf:"bytes,1,opt,name=database,proto3" json:"database,omitempty"`
}

func (x *EvalResultListRequest) Reset() {
	*x = EvalResultListRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_foyle_v1alpha1_eval_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EvalResultListRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EvalResultListRequest) ProtoMessage() {}

func (x *EvalResultListRequest) ProtoReflect() protoreflect.Message {
	mi := &file_foyle_v1alpha1_eval_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EvalResultListRequest.ProtoReflect.Descriptor instead.
func (*EvalResultListRequest) Descriptor() ([]byte, []int) {
	return file_foyle_v1alpha1_eval_proto_rawDescGZIP(), []int{2}
}

func (x *EvalResultListRequest) GetDatabase() string {
	if x != nil {
		return x.Database
	}
	return ""
}

type EvalResultListResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Items []*EvalResult `protobuf:"bytes,1,rep,name=items,proto3" json:"items,omitempty"`
}

func (x *EvalResultListResponse) Reset() {
	*x = EvalResultListResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_foyle_v1alpha1_eval_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EvalResultListResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EvalResultListResponse) ProtoMessage() {}

func (x *EvalResultListResponse) ProtoReflect() protoreflect.Message {
	mi := &file_foyle_v1alpha1_eval_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EvalResultListResponse.ProtoReflect.Descriptor instead.
func (*EvalResultListResponse) Descriptor() ([]byte, []int) {
	return file_foyle_v1alpha1_eval_proto_rawDescGZIP(), []int{3}
}

func (x *EvalResultListResponse) GetItems() []*EvalResult {
	if x != nil {
		return x.Items
	}
	return nil
}

// AssertionRow represents a row in the assertion table.
// It is intended for returning the results of assertions. In a way that makes it easy to view the assertions
// in a table inside a RunMe notebook. So we need to flatten the data.
type AssertionRow struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// id of the example
	Id          string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	ExampleFile string `protobuf:"bytes,2,opt,name=exampleFile,proto3" json:"exampleFile,omitempty"`
	// Document markdown
	DocMd    string `protobuf:"bytes,3,opt,name=doc_md,json=docMd,proto3" json:"doc_md,omitempty"`
	AnswerMd string `protobuf:"bytes,4,opt,name=answer_md,json=answerMd,proto3" json:"answer_md,omitempty"`
	// TODO(jeremy): How can we avoid having to add each assertion here
	CodeAfterMarkdown AssertResult `protobuf:"varint,5,opt,name=code_after_markdown,json=codeAfterMarkdown,proto3,enum=AssertResult" json:"code_after_markdown,omitempty"`
	OneCodeCell       AssertResult `protobuf:"varint,6,opt,name=one_code_cell,json=oneCodeCell,proto3,enum=AssertResult" json:"one_code_cell,omitempty"`
	EndsWithCodeCell  AssertResult `protobuf:"varint,7,opt,name=ends_with_code_cell,json=endsWithCodeCell,proto3,enum=AssertResult" json:"ends_with_code_cell,omitempty"`
}

func (x *AssertionRow) Reset() {
	*x = AssertionRow{}
	if protoimpl.UnsafeEnabled {
		mi := &file_foyle_v1alpha1_eval_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AssertionRow) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AssertionRow) ProtoMessage() {}

func (x *AssertionRow) ProtoReflect() protoreflect.Message {
	mi := &file_foyle_v1alpha1_eval_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AssertionRow.ProtoReflect.Descriptor instead.
func (*AssertionRow) Descriptor() ([]byte, []int) {
	return file_foyle_v1alpha1_eval_proto_rawDescGZIP(), []int{4}
}

func (x *AssertionRow) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *AssertionRow) GetExampleFile() string {
	if x != nil {
		return x.ExampleFile
	}
	return ""
}

func (x *AssertionRow) GetDocMd() string {
	if x != nil {
		return x.DocMd
	}
	return ""
}

func (x *AssertionRow) GetAnswerMd() string {
	if x != nil {
		return x.AnswerMd
	}
	return ""
}

func (x *AssertionRow) GetCodeAfterMarkdown() AssertResult {
	if x != nil {
		return x.CodeAfterMarkdown
	}
	return AssertResult_UNKNOWN_AssertResult
}

func (x *AssertionRow) GetOneCodeCell() AssertResult {
	if x != nil {
		return x.OneCodeCell
	}
	return AssertResult_UNKNOWN_AssertResult
}

func (x *AssertionRow) GetEndsWithCodeCell() AssertResult {
	if x != nil {
		return x.EndsWithCodeCell
	}
	return AssertResult_UNKNOWN_AssertResult
}

type AssertionTableRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The path of the database to fetch results for
	Database string `protobuf:"bytes,1,opt,name=database,proto3" json:"database,omitempty"`
}

func (x *AssertionTableRequest) Reset() {
	*x = AssertionTableRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_foyle_v1alpha1_eval_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AssertionTableRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AssertionTableRequest) ProtoMessage() {}

func (x *AssertionTableRequest) ProtoReflect() protoreflect.Message {
	mi := &file_foyle_v1alpha1_eval_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AssertionTableRequest.ProtoReflect.Descriptor instead.
func (*AssertionTableRequest) Descriptor() ([]byte, []int) {
	return file_foyle_v1alpha1_eval_proto_rawDescGZIP(), []int{5}
}

func (x *AssertionTableRequest) GetDatabase() string {
	if x != nil {
		return x.Database
	}
	return ""
}

type AssertionTableResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Rows []*AssertionRow `protobuf:"bytes,1,rep,name=rows,proto3" json:"rows,omitempty"`
}

func (x *AssertionTableResponse) Reset() {
	*x = AssertionTableResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_foyle_v1alpha1_eval_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AssertionTableResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AssertionTableResponse) ProtoMessage() {}

func (x *AssertionTableResponse) ProtoReflect() protoreflect.Message {
	mi := &file_foyle_v1alpha1_eval_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AssertionTableResponse.ProtoReflect.Descriptor instead.
func (*AssertionTableResponse) Descriptor() ([]byte, []int) {
	return file_foyle_v1alpha1_eval_proto_rawDescGZIP(), []int{6}
}

func (x *AssertionTableResponse) GetRows() []*AssertionRow {
	if x != nil {
		return x.Rows
	}
	return nil
}

var File_foyle_v1alpha1_eval_proto protoreflect.FileDescriptor

var file_foyle_v1alpha1_eval_proto_rawDesc = []byte{
	0x0a, 0x19, 0x66, 0x6f, 0x79, 0x6c, 0x65, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31,
	0x2f, 0x65, 0x76, 0x61, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x18, 0x66, 0x6f, 0x79,
	0x6c, 0x65, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2f, 0x64, 0x6f, 0x63, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1c, 0x66, 0x6f, 0x79, 0x6c, 0x65, 0x2f, 0x76, 0x31, 0x61,
	0x6c, 0x70, 0x68, 0x61, 0x31, 0x2f, 0x74, 0x72, 0x61, 0x69, 0x6e, 0x65, 0x72, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x75, 0x66, 0x2f, 0x73, 0x74, 0x72, 0x75, 0x63, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x22, 0x83, 0x03, 0x0a, 0x0a, 0x45, 0x76, 0x61, 0x6c, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74,
	0x12, 0x22, 0x0a, 0x07, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x08, 0x2e, 0x45, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x52, 0x07, 0x65, 0x78, 0x61,
	0x6d, 0x70, 0x6c, 0x65, 0x12, 0x21, 0x0a, 0x0c, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x5f,
	0x66, 0x69, 0x6c, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x65, 0x78, 0x61, 0x6d,
	0x70, 0x6c, 0x65, 0x46, 0x69, 0x6c, 0x65, 0x12, 0x1e, 0x0a, 0x06, 0x61, 0x63, 0x74, 0x75, 0x61,
	0x6c, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x06, 0x2e, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x52,
	0x06, 0x61, 0x63, 0x74, 0x75, 0x61, 0x6c, 0x12, 0x1a, 0x0a, 0x08, 0x64, 0x69, 0x73, 0x74, 0x61,
	0x6e, 0x63, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x05, 0x52, 0x08, 0x64, 0x69, 0x73, 0x74, 0x61,
	0x6e, 0x63, 0x65, 0x12, 0x2f, 0x0a, 0x13, 0x6e, 0x6f, 0x72, 0x6d, 0x61, 0x6c, 0x69, 0x7a, 0x65,
	0x64, 0x5f, 0x64, 0x69, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x18, 0x07, 0x20, 0x01, 0x28, 0x02,
	0x52, 0x12, 0x6e, 0x6f, 0x72, 0x6d, 0x61, 0x6c, 0x69, 0x7a, 0x65, 0x64, 0x44, 0x69, 0x73, 0x74,
	0x61, 0x6e, 0x63, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x18, 0x05, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x12, 0x29, 0x0a, 0x06, 0x73, 0x74,
	0x61, 0x74, 0x75, 0x73, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x11, 0x2e, 0x45, 0x76, 0x61,
	0x6c, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x06, 0x73,
	0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x20, 0x0a, 0x0c, 0x67, 0x65, 0x6e, 0x5f, 0x74, 0x72, 0x61,
	0x63, 0x65, 0x5f, 0x69, 0x64, 0x18, 0x08, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x67, 0x65, 0x6e,
	0x54, 0x72, 0x61, 0x63, 0x65, 0x49, 0x64, 0x12, 0x32, 0x0a, 0x0f, 0x62, 0x65, 0x73, 0x74, 0x5f,
	0x72, 0x61, 0x67, 0x5f, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x18, 0x09, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x0a, 0x2e, 0x52, 0x41, 0x47, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x52, 0x0d, 0x62, 0x65,
	0x73, 0x74, 0x52, 0x61, 0x67, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x12, 0x2a, 0x0a, 0x0a, 0x61,
	0x73, 0x73, 0x65, 0x72, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x0a, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x0a, 0x2e, 0x41, 0x73, 0x73, 0x65, 0x72, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x0a, 0x61, 0x73, 0x73,
	0x65, 0x72, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x22, 0x5e, 0x0a, 0x09, 0x41, 0x73, 0x73, 0x65, 0x72,
	0x74, 0x69, 0x6f, 0x6e, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x25, 0x0a, 0x06, 0x72, 0x65, 0x73, 0x75,
	0x6c, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x0d, 0x2e, 0x41, 0x73, 0x73, 0x65, 0x72,
	0x74, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x52, 0x06, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x12,
	0x16, 0x0a, 0x06, 0x64, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x06, 0x64, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x22, 0x33, 0x0a, 0x15, 0x45, 0x76, 0x61, 0x6c, 0x52,
	0x65, 0x73, 0x75, 0x6c, 0x74, 0x4c, 0x69, 0x73, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x12, 0x1a, 0x0a, 0x08, 0x64, 0x61, 0x74, 0x61, 0x62, 0x61, 0x73, 0x65, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x08, 0x64, 0x61, 0x74, 0x61, 0x62, 0x61, 0x73, 0x65, 0x22, 0x3b, 0x0a, 0x16,
	0x45, 0x76, 0x61, 0x6c, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x4c, 0x69, 0x73, 0x74, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x21, 0x0a, 0x05, 0x69, 0x74, 0x65, 0x6d, 0x73, 0x18,
	0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0b, 0x2e, 0x45, 0x76, 0x61, 0x6c, 0x52, 0x65, 0x73, 0x75,
	0x6c, 0x74, 0x52, 0x05, 0x69, 0x74, 0x65, 0x6d, 0x73, 0x22, 0xa4, 0x02, 0x0a, 0x0c, 0x41, 0x73,
	0x73, 0x65, 0x72, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x6f, 0x77, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x20, 0x0a, 0x0b, 0x65, 0x78,
	0x61, 0x6d, 0x70, 0x6c, 0x65, 0x46, 0x69, 0x6c, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x0b, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x46, 0x69, 0x6c, 0x65, 0x12, 0x15, 0x0a, 0x06,
	0x64, 0x6f, 0x63, 0x5f, 0x6d, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x64, 0x6f,
	0x63, 0x4d, 0x64, 0x12, 0x1b, 0x0a, 0x09, 0x61, 0x6e, 0x73, 0x77, 0x65, 0x72, 0x5f, 0x6d, 0x64,
	0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x61, 0x6e, 0x73, 0x77, 0x65, 0x72, 0x4d, 0x64,
	0x12, 0x3d, 0x0a, 0x13, 0x63, 0x6f, 0x64, 0x65, 0x5f, 0x61, 0x66, 0x74, 0x65, 0x72, 0x5f, 0x6d,
	0x61, 0x72, 0x6b, 0x64, 0x6f, 0x77, 0x6e, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x0d, 0x2e,
	0x41, 0x73, 0x73, 0x65, 0x72, 0x74, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x52, 0x11, 0x63, 0x6f,
	0x64, 0x65, 0x41, 0x66, 0x74, 0x65, 0x72, 0x4d, 0x61, 0x72, 0x6b, 0x64, 0x6f, 0x77, 0x6e, 0x12,
	0x31, 0x0a, 0x0d, 0x6f, 0x6e, 0x65, 0x5f, 0x63, 0x6f, 0x64, 0x65, 0x5f, 0x63, 0x65, 0x6c, 0x6c,
	0x18, 0x06, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x0d, 0x2e, 0x41, 0x73, 0x73, 0x65, 0x72, 0x74, 0x52,
	0x65, 0x73, 0x75, 0x6c, 0x74, 0x52, 0x0b, 0x6f, 0x6e, 0x65, 0x43, 0x6f, 0x64, 0x65, 0x43, 0x65,
	0x6c, 0x6c, 0x12, 0x3c, 0x0a, 0x13, 0x65, 0x6e, 0x64, 0x73, 0x5f, 0x77, 0x69, 0x74, 0x68, 0x5f,
	0x63, 0x6f, 0x64, 0x65, 0x5f, 0x63, 0x65, 0x6c, 0x6c, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0e, 0x32,
	0x0d, 0x2e, 0x41, 0x73, 0x73, 0x65, 0x72, 0x74, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x52, 0x10,
	0x65, 0x6e, 0x64, 0x73, 0x57, 0x69, 0x74, 0x68, 0x43, 0x6f, 0x64, 0x65, 0x43, 0x65, 0x6c, 0x6c,
	0x22, 0x33, 0x0a, 0x15, 0x41, 0x73, 0x73, 0x65, 0x72, 0x74, 0x69, 0x6f, 0x6e, 0x54, 0x61, 0x62,
	0x6c, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1a, 0x0a, 0x08, 0x64, 0x61, 0x74,
	0x61, 0x62, 0x61, 0x73, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x64, 0x61, 0x74,
	0x61, 0x62, 0x61, 0x73, 0x65, 0x22, 0x3b, 0x0a, 0x16, 0x41, 0x73, 0x73, 0x65, 0x72, 0x74, 0x69,
	0x6f, 0x6e, 0x54, 0x61, 0x62, 0x6c, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12,
	0x21, 0x0a, 0x04, 0x72, 0x6f, 0x77, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0d, 0x2e,
	0x41, 0x73, 0x73, 0x65, 0x72, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x6f, 0x77, 0x52, 0x04, 0x72, 0x6f,
	0x77, 0x73, 0x2a, 0x47, 0x0a, 0x10, 0x45, 0x76, 0x61, 0x6c, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74,
	0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x1e, 0x0a, 0x1a, 0x55, 0x4e, 0x4b, 0x4e, 0x4f, 0x57,
	0x4e, 0x5f, 0x45, 0x56, 0x41, 0x4c, 0x5f, 0x52, 0x45, 0x53, 0x55, 0x4c, 0x54, 0x5f, 0x53, 0x54,
	0x41, 0x54, 0x55, 0x53, 0x10, 0x00, 0x12, 0x08, 0x0a, 0x04, 0x44, 0x4f, 0x4e, 0x45, 0x10, 0x01,
	0x12, 0x09, 0x0a, 0x05, 0x45, 0x52, 0x52, 0x4f, 0x52, 0x10, 0x02, 0x2a, 0x4d, 0x0a, 0x0c, 0x41,
	0x73, 0x73, 0x65, 0x72, 0x74, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x12, 0x18, 0x0a, 0x14, 0x55,
	0x4e, 0x4b, 0x4e, 0x4f, 0x57, 0x4e, 0x5f, 0x41, 0x73, 0x73, 0x65, 0x72, 0x74, 0x52, 0x65, 0x73,
	0x75, 0x6c, 0x74, 0x10, 0x00, 0x12, 0x0a, 0x0a, 0x06, 0x50, 0x41, 0x53, 0x53, 0x45, 0x44, 0x10,
	0x01, 0x12, 0x0a, 0x0a, 0x06, 0x46, 0x41, 0x49, 0x4c, 0x45, 0x44, 0x10, 0x02, 0x12, 0x0b, 0x0a,
	0x07, 0x53, 0x4b, 0x49, 0x50, 0x50, 0x45, 0x44, 0x10, 0x03, 0x32, 0x8d, 0x01, 0x0a, 0x0b, 0x45,
	0x76, 0x61, 0x6c, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x39, 0x0a, 0x04, 0x4c, 0x69,
	0x73, 0x74, 0x12, 0x16, 0x2e, 0x45, 0x76, 0x61, 0x6c, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x4c,
	0x69, 0x73, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x17, 0x2e, 0x45, 0x76, 0x61,
	0x6c, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x4c, 0x69, 0x73, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x43, 0x0a, 0x0e, 0x41, 0x73, 0x73, 0x65, 0x72, 0x74, 0x69,
	0x6f, 0x6e, 0x54, 0x61, 0x62, 0x6c, 0x65, 0x12, 0x16, 0x2e, 0x41, 0x73, 0x73, 0x65, 0x72, 0x74,
	0x69, 0x6f, 0x6e, 0x54, 0x61, 0x62, 0x6c, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a,
	0x17, 0x2e, 0x41, 0x73, 0x73, 0x65, 0x72, 0x74, 0x69, 0x6f, 0x6e, 0x54, 0x61, 0x62, 0x6c, 0x65,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x42, 0x3e, 0x42, 0x09, 0x45, 0x76,
	0x61, 0x6c, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x2f, 0x67, 0x69, 0x74, 0x68, 0x75,
	0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6a, 0x6c, 0x65, 0x77, 0x69, 0x2f, 0x66, 0x6f, 0x79, 0x6c,
	0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2f, 0x67, 0x6f, 0x2f, 0x66, 0x6f, 0x79, 0x6c,
	0x65, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
}

var (
	file_foyle_v1alpha1_eval_proto_rawDescOnce sync.Once
	file_foyle_v1alpha1_eval_proto_rawDescData = file_foyle_v1alpha1_eval_proto_rawDesc
)

func file_foyle_v1alpha1_eval_proto_rawDescGZIP() []byte {
	file_foyle_v1alpha1_eval_proto_rawDescOnce.Do(func() {
		file_foyle_v1alpha1_eval_proto_rawDescData = protoimpl.X.CompressGZIP(file_foyle_v1alpha1_eval_proto_rawDescData)
	})
	return file_foyle_v1alpha1_eval_proto_rawDescData
}

var file_foyle_v1alpha1_eval_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_foyle_v1alpha1_eval_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_foyle_v1alpha1_eval_proto_goTypes = []interface{}{
	(EvalResultStatus)(0),          // 0: EvalResultStatus
	(AssertResult)(0),              // 1: AssertResult
	(*EvalResult)(nil),             // 2: EvalResult
	(*Assertion)(nil),              // 3: Assertion
	(*EvalResultListRequest)(nil),  // 4: EvalResultListRequest
	(*EvalResultListResponse)(nil), // 5: EvalResultListResponse
	(*AssertionRow)(nil),           // 6: AssertionRow
	(*AssertionTableRequest)(nil),  // 7: AssertionTableRequest
	(*AssertionTableResponse)(nil), // 8: AssertionTableResponse
	(*Example)(nil),                // 9: Example
	(*Block)(nil),                  // 10: Block
	(*RAGResult)(nil),              // 11: RAGResult
}
var file_foyle_v1alpha1_eval_proto_depIdxs = []int32{
	9,  // 0: EvalResult.example:type_name -> Example
	10, // 1: EvalResult.actual:type_name -> Block
	0,  // 2: EvalResult.status:type_name -> EvalResultStatus
	11, // 3: EvalResult.best_rag_result:type_name -> RAGResult
	3,  // 4: EvalResult.assertions:type_name -> Assertion
	1,  // 5: Assertion.result:type_name -> AssertResult
	2,  // 6: EvalResultListResponse.items:type_name -> EvalResult
	1,  // 7: AssertionRow.code_after_markdown:type_name -> AssertResult
	1,  // 8: AssertionRow.one_code_cell:type_name -> AssertResult
	1,  // 9: AssertionRow.ends_with_code_cell:type_name -> AssertResult
	6,  // 10: AssertionTableResponse.rows:type_name -> AssertionRow
	4,  // 11: EvalService.List:input_type -> EvalResultListRequest
	7,  // 12: EvalService.AssertionTable:input_type -> AssertionTableRequest
	5,  // 13: EvalService.List:output_type -> EvalResultListResponse
	8,  // 14: EvalService.AssertionTable:output_type -> AssertionTableResponse
	13, // [13:15] is the sub-list for method output_type
	11, // [11:13] is the sub-list for method input_type
	11, // [11:11] is the sub-list for extension type_name
	11, // [11:11] is the sub-list for extension extendee
	0,  // [0:11] is the sub-list for field type_name
}

func init() { file_foyle_v1alpha1_eval_proto_init() }
func file_foyle_v1alpha1_eval_proto_init() {
	if File_foyle_v1alpha1_eval_proto != nil {
		return
	}
	file_foyle_v1alpha1_doc_proto_init()
	file_foyle_v1alpha1_trainer_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_foyle_v1alpha1_eval_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*EvalResult); i {
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
		file_foyle_v1alpha1_eval_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Assertion); i {
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
		file_foyle_v1alpha1_eval_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*EvalResultListRequest); i {
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
		file_foyle_v1alpha1_eval_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*EvalResultListResponse); i {
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
		file_foyle_v1alpha1_eval_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AssertionRow); i {
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
		file_foyle_v1alpha1_eval_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AssertionTableRequest); i {
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
		file_foyle_v1alpha1_eval_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AssertionTableResponse); i {
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
			RawDescriptor: file_foyle_v1alpha1_eval_proto_rawDesc,
			NumEnums:      2,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_foyle_v1alpha1_eval_proto_goTypes,
		DependencyIndexes: file_foyle_v1alpha1_eval_proto_depIdxs,
		EnumInfos:         file_foyle_v1alpha1_eval_proto_enumTypes,
		MessageInfos:      file_foyle_v1alpha1_eval_proto_msgTypes,
	}.Build()
	File_foyle_v1alpha1_eval_proto = out.File
	file_foyle_v1alpha1_eval_proto_rawDesc = nil
	file_foyle_v1alpha1_eval_proto_goTypes = nil
	file_foyle_v1alpha1_eval_proto_depIdxs = nil
}
