// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: foyle/logs/traces.proto

package logspb

import (
	fmt "fmt"
	math "math"
	proto "github.com/golang/protobuf/proto"
	_ "google.golang.org/protobuf/types/known/structpb"
	_ "google.golang.org/protobuf/types/known/timestamppb"
	_ "github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	_ "github.com/stateful/runme/v3/pkg/api/gen/proto/go/runme/runner/v1"
	go_uber_org_zap_zapcore "go.uber.org/zap/zapcore"
	github_com_golang_protobuf_ptypes "github.com/golang/protobuf/ptypes"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

func (m *Trace) MarshalLogObject(enc go_uber_org_zap_zapcore.ObjectEncoder) error {
	var keyName string
	_ = keyName

	if m == nil {
		return nil
	}

	keyName = "id" // field id = 1
	enc.AddString(keyName, m.Id)

	keyName = "start_time" // field start_time = 2
	if t, err := github_com_golang_protobuf_ptypes.Timestamp(m.StartTime); err == nil {
		enc.AddTime(keyName, t)
	}

	keyName = "end_time" // field end_time = 3
	if t, err := github_com_golang_protobuf_ptypes.Timestamp(m.EndTime); err == nil {
		enc.AddTime(keyName, t)
	}

	keyName = "generate" // field generate = 4
	if ov, ok := m.GetData().(*Trace_Generate); ok {
		_ = ov
		if ov.Generate != nil {
			var vv interface{} = ov.Generate
			if marshaler, ok := vv.(go_uber_org_zap_zapcore.ObjectMarshaler); ok {
				enc.AddObject(keyName, marshaler)
			}
		}
	}

	keyName = "execute" // field execute = 5
	if ov, ok := m.GetData().(*Trace_Execute); ok {
		_ = ov
		if ov.Execute != nil {
			var vv interface{} = ov.Execute
			if marshaler, ok := vv.(go_uber_org_zap_zapcore.ObjectMarshaler); ok {
				enc.AddObject(keyName, marshaler)
			}
		}
	}

	keyName = "eval_mode" // field eval_mode = 6
	enc.AddBool(keyName, m.EvalMode)

	return nil
}

func (m *GenerateTrace) MarshalLogObject(enc go_uber_org_zap_zapcore.ObjectEncoder) error {
	var keyName string
	_ = keyName

	if m == nil {
		return nil
	}

	keyName = "request" // field request = 1
	if m.Request != nil {
		var vv interface{} = m.Request
		if marshaler, ok := vv.(go_uber_org_zap_zapcore.ObjectMarshaler); ok {
			enc.AddObject(keyName, marshaler)
		}
	}

	keyName = "response" // field response = 2
	if m.Response != nil {
		var vv interface{} = m.Response
		if marshaler, ok := vv.(go_uber_org_zap_zapcore.ObjectMarshaler); ok {
			enc.AddObject(keyName, marshaler)
		}
	}

	return nil
}

func (m *ExecuteTrace) MarshalLogObject(enc go_uber_org_zap_zapcore.ObjectEncoder) error {
	var keyName string
	_ = keyName

	if m == nil {
		return nil
	}

	keyName = "request" // field request = 1
	if m.Request != nil {
		var vv interface{} = m.Request
		if marshaler, ok := vv.(go_uber_org_zap_zapcore.ObjectMarshaler); ok {
			enc.AddObject(keyName, marshaler)
		}
	}

	keyName = "response" // field response = 2
	if m.Response != nil {
		var vv interface{} = m.Response
		if marshaler, ok := vv.(go_uber_org_zap_zapcore.ObjectMarshaler); ok {
			enc.AddObject(keyName, marshaler)
		}
	}

	return nil
}

func (m *RunMeTrace) MarshalLogObject(enc go_uber_org_zap_zapcore.ObjectEncoder) error {
	var keyName string
	_ = keyName

	if m == nil {
		return nil
	}

	keyName = "request" // field request = 1
	if m.Request != nil {
		var vv interface{} = m.Request
		if marshaler, ok := vv.(go_uber_org_zap_zapcore.ObjectMarshaler); ok {
			enc.AddObject(keyName, marshaler)
		}
	}

	return nil
}
