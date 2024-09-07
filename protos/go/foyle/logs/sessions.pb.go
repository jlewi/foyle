// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        (unknown)
// source: foyle/logs/sessions.proto

package logspb

import (
	v1alpha1 "github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	_ "github.com/stateful/runme/v3/pkg/api/gen/proto/go/runme/runner/v1"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// Session is a series of events in the logs
type Session struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// context_id is the unique id for the context
	ContextId string                 `protobuf:"bytes,1,opt,name=context_id,json=contextId,proto3" json:"context_id,omitempty"`
	StartTime *timestamppb.Timestamp `protobuf:"bytes,2,opt,name=start_time,json=startTime,proto3" json:"start_time,omitempty"`
	EndTime   *timestamppb.Timestamp `protobuf:"bytes,3,opt,name=end_time,json=endTime,proto3" json:"end_time,omitempty"`
	// Log events associated with the session.
	LogEvents []*v1alpha1.LogEvent `protobuf:"bytes,4,rep,name=log_events,json=logEvents,proto3" json:"log_events,omitempty"`
	// FullContext is the full context of the session
	FullContext *v1alpha1.FullContext `protobuf:"bytes,5,opt,name=full_context,json=fullContext,proto3" json:"full_context,omitempty"`
}

func (x *Session) Reset() {
	*x = Session{}
	if protoimpl.UnsafeEnabled {
		mi := &file_foyle_logs_sessions_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Session) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Session) ProtoMessage() {}

func (x *Session) ProtoReflect() protoreflect.Message {
	mi := &file_foyle_logs_sessions_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Session.ProtoReflect.Descriptor instead.
func (*Session) Descriptor() ([]byte, []int) {
	return file_foyle_logs_sessions_proto_rawDescGZIP(), []int{0}
}

func (x *Session) GetContextId() string {
	if x != nil {
		return x.ContextId
	}
	return ""
}

func (x *Session) GetStartTime() *timestamppb.Timestamp {
	if x != nil {
		return x.StartTime
	}
	return nil
}

func (x *Session) GetEndTime() *timestamppb.Timestamp {
	if x != nil {
		return x.EndTime
	}
	return nil
}

func (x *Session) GetLogEvents() []*v1alpha1.LogEvent {
	if x != nil {
		return x.LogEvents
	}
	return nil
}

func (x *Session) GetFullContext() *v1alpha1.FullContext {
	if x != nil {
		return x.FullContext
	}
	return nil
}

var File_foyle_logs_sessions_proto protoreflect.FileDescriptor

var file_foyle_logs_sessions_proto_rawDesc = []byte{
	0x0a, 0x19, 0x66, 0x6f, 0x79, 0x6c, 0x65, 0x2f, 0x6c, 0x6f, 0x67, 0x73, 0x2f, 0x73, 0x65, 0x73,
	0x73, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0a, 0x66, 0x6f, 0x79,
	0x6c, 0x65, 0x2e, 0x6c, 0x6f, 0x67, 0x73, 0x1a, 0x1a, 0x66, 0x6f, 0x79, 0x6c, 0x65, 0x2f, 0x76,
	0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2f, 0x61, 0x67, 0x65, 0x6e, 0x74, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x1a, 0x1c, 0x66, 0x6f, 0x79, 0x6c, 0x65, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70,
	0x68, 0x61, 0x31, 0x2f, 0x74, 0x72, 0x61, 0x69, 0x6e, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x1a, 0x17, 0x66, 0x6f, 0x79, 0x6c, 0x65, 0x2f, 0x6c, 0x6f, 0x67, 0x73, 0x2f, 0x62, 0x6c,
	0x6f, 0x63, 0x6b, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1c, 0x72, 0x75, 0x6e, 0x6d,
	0x65, 0x2f, 0x72, 0x75, 0x6e, 0x6e, 0x65, 0x72, 0x2f, 0x76, 0x31, 0x2f, 0x72, 0x75, 0x6e, 0x6e,
	0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74,
	0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xf5, 0x01, 0x0a, 0x07, 0x53, 0x65,
	0x73, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x1d, 0x0a, 0x0a, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74,
	0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x63, 0x6f, 0x6e, 0x74, 0x65,
	0x78, 0x74, 0x49, 0x64, 0x12, 0x39, 0x0a, 0x0a, 0x73, 0x74, 0x61, 0x72, 0x74, 0x5f, 0x74, 0x69,
	0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73,
	0x74, 0x61, 0x6d, 0x70, 0x52, 0x09, 0x73, 0x74, 0x61, 0x72, 0x74, 0x54, 0x69, 0x6d, 0x65, 0x12,
	0x35, 0x0a, 0x08, 0x65, 0x6e, 0x64, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x07, 0x65,
	0x6e, 0x64, 0x54, 0x69, 0x6d, 0x65, 0x12, 0x28, 0x0a, 0x0a, 0x6c, 0x6f, 0x67, 0x5f, 0x65, 0x76,
	0x65, 0x6e, 0x74, 0x73, 0x18, 0x04, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x09, 0x2e, 0x4c, 0x6f, 0x67,
	0x45, 0x76, 0x65, 0x6e, 0x74, 0x52, 0x09, 0x6c, 0x6f, 0x67, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x73,
	0x12, 0x2f, 0x0a, 0x0c, 0x66, 0x75, 0x6c, 0x6c, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74,
	0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0c, 0x2e, 0x46, 0x75, 0x6c, 0x6c, 0x43, 0x6f, 0x6e,
	0x74, 0x65, 0x78, 0x74, 0x52, 0x0b, 0x66, 0x75, 0x6c, 0x6c, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x78,
	0x74, 0x42, 0x9c, 0x01, 0x0a, 0x0e, 0x63, 0x6f, 0x6d, 0x2e, 0x66, 0x6f, 0x79, 0x6c, 0x65, 0x2e,
	0x6c, 0x6f, 0x67, 0x73, 0x42, 0x0d, 0x53, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x50, 0x72,
	0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x32, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f,
	0x6d, 0x2f, 0x6a, 0x6c, 0x65, 0x77, 0x69, 0x2f, 0x66, 0x6f, 0x79, 0x6c, 0x65, 0x2f, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x73, 0x2f, 0x67, 0x6f, 0x2f, 0x66, 0x6f, 0x79, 0x6c, 0x65, 0x2f, 0x6c, 0x6f,
	0x67, 0x73, 0x3b, 0x6c, 0x6f, 0x67, 0x73, 0x70, 0x62, 0xa2, 0x02, 0x03, 0x46, 0x4c, 0x58, 0xaa,
	0x02, 0x0a, 0x46, 0x6f, 0x79, 0x6c, 0x65, 0x2e, 0x4c, 0x6f, 0x67, 0x73, 0xca, 0x02, 0x0a, 0x46,
	0x6f, 0x79, 0x6c, 0x65, 0x5c, 0x4c, 0x6f, 0x67, 0x73, 0xe2, 0x02, 0x16, 0x46, 0x6f, 0x79, 0x6c,
	0x65, 0x5c, 0x4c, 0x6f, 0x67, 0x73, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61,
	0x74, 0x61, 0xea, 0x02, 0x0b, 0x46, 0x6f, 0x79, 0x6c, 0x65, 0x3a, 0x3a, 0x4c, 0x6f, 0x67, 0x73,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_foyle_logs_sessions_proto_rawDescOnce sync.Once
	file_foyle_logs_sessions_proto_rawDescData = file_foyle_logs_sessions_proto_rawDesc
)

func file_foyle_logs_sessions_proto_rawDescGZIP() []byte {
	file_foyle_logs_sessions_proto_rawDescOnce.Do(func() {
		file_foyle_logs_sessions_proto_rawDescData = protoimpl.X.CompressGZIP(file_foyle_logs_sessions_proto_rawDescData)
	})
	return file_foyle_logs_sessions_proto_rawDescData
}

var file_foyle_logs_sessions_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_foyle_logs_sessions_proto_goTypes = []interface{}{
	(*Session)(nil),               // 0: foyle.logs.Session
	(*timestamppb.Timestamp)(nil), // 1: google.protobuf.Timestamp
	(*v1alpha1.LogEvent)(nil),     // 2: LogEvent
	(*v1alpha1.FullContext)(nil),  // 3: FullContext
}
var file_foyle_logs_sessions_proto_depIdxs = []int32{
	1, // 0: foyle.logs.Session.start_time:type_name -> google.protobuf.Timestamp
	1, // 1: foyle.logs.Session.end_time:type_name -> google.protobuf.Timestamp
	2, // 2: foyle.logs.Session.log_events:type_name -> LogEvent
	3, // 3: foyle.logs.Session.full_context:type_name -> FullContext
	4, // [4:4] is the sub-list for method output_type
	4, // [4:4] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_foyle_logs_sessions_proto_init() }
func file_foyle_logs_sessions_proto_init() {
	if File_foyle_logs_sessions_proto != nil {
		return
	}
	file_foyle_logs_blocks_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_foyle_logs_sessions_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Session); i {
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
			RawDescriptor: file_foyle_logs_sessions_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_foyle_logs_sessions_proto_goTypes,
		DependencyIndexes: file_foyle_logs_sessions_proto_depIdxs,
		MessageInfos:      file_foyle_logs_sessions_proto_msgTypes,
	}.Build()
	File_foyle_logs_sessions_proto = out.File
	file_foyle_logs_sessions_proto_rawDesc = nil
	file_foyle_logs_sessions_proto_goTypes = nil
	file_foyle_logs_sessions_proto_depIdxs = nil
}
