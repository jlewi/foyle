// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: foyle/logs/blocks.proto

package logspb

import (
	fmt "fmt"
	math "math"
	proto "github.com/golang/protobuf/proto"
	_ "github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	_ "google.golang.org/protobuf/types/known/structpb"
	_ "google.golang.org/protobuf/types/known/timestamppb"
	go_uber_org_zap_zapcore "go.uber.org/zap/zapcore"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

func (m *BlockLog) MarshalLogObject(enc go_uber_org_zap_zapcore.ObjectEncoder) error {
	var keyName string
	_ = keyName

	if m == nil {
		return nil
	}

	keyName = "id" // field id = 1
	enc.AddString(keyName, m.Id)

	keyName = "gen_trace_id" // field gen_trace_id = 2
	enc.AddString(keyName, m.GenTraceId)

	keyName = "exec_trace_ids" // field exec_trace_ids = 3
	enc.AddArray(keyName, go_uber_org_zap_zapcore.ArrayMarshalerFunc(func(aenc go_uber_org_zap_zapcore.ArrayEncoder) error {
		for _, rv := range m.ExecTraceIds {
			_ = rv
			aenc.AppendString(rv)
		}
		return nil
	}))

	keyName = "doc" // field doc = 4
	if m.Doc != nil {
		var vv interface{} = m.Doc
		if marshaler, ok := vv.(go_uber_org_zap_zapcore.ObjectMarshaler); ok {
			enc.AddObject(keyName, marshaler)
		}
	}

	keyName = "generated_block" // field generated_block = 5
	if m.GeneratedBlock != nil {
		var vv interface{} = m.GeneratedBlock
		if marshaler, ok := vv.(go_uber_org_zap_zapcore.ObjectMarshaler); ok {
			enc.AddObject(keyName, marshaler)
		}
	}

	keyName = "executed_block" // field executed_block = 6
	if m.ExecutedBlock != nil {
		var vv interface{} = m.ExecutedBlock
		if marshaler, ok := vv.(go_uber_org_zap_zapcore.ObjectMarshaler); ok {
			enc.AddObject(keyName, marshaler)
		}
	}

	keyName = "exit_code" // field exit_code = 7
	enc.AddInt32(keyName, m.ExitCode)

	keyName = "eval_mode" // field eval_mode = 8
	enc.AddBool(keyName, m.EvalMode)

	return nil
}
