// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: foyle/v1alpha1/agent.proto

package v1alpha1

import (
	fmt "fmt"
	math "math"
	proto "github.com/golang/protobuf/proto"
	_ "google.golang.org/protobuf/types/known/structpb"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	go_uber_org_zap_zapcore "go.uber.org/zap/zapcore"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

func (m *GenerateRequest) MarshalLogObject(enc go_uber_org_zap_zapcore.ObjectEncoder) error {
	var keyName string
	_ = keyName

	if m == nil {
		return nil
	}

	keyName = "doc" // field doc = 1
	if m.Doc != nil {
		var vv interface{} = m.Doc
		if marshaler, ok := vv.(go_uber_org_zap_zapcore.ObjectMarshaler); ok {
			enc.AddObject(keyName, marshaler)
		}
	}

	return nil
}

func (m *GenerateResponse) MarshalLogObject(enc go_uber_org_zap_zapcore.ObjectEncoder) error {
	var keyName string
	_ = keyName

	if m == nil {
		return nil
	}

	keyName = "blocks" // field blocks = 1
	enc.AddArray(keyName, go_uber_org_zap_zapcore.ArrayMarshalerFunc(func(aenc go_uber_org_zap_zapcore.ArrayEncoder) error {
		for _, rv := range m.Blocks {
			_ = rv
			if rv != nil {
				var vv interface{} = rv
				if marshaler, ok := vv.(go_uber_org_zap_zapcore.ObjectMarshaler); ok {
					aenc.AppendObject(marshaler)
				}
			}
		}
		return nil
	}))

	keyName = "trace_id" // field trace_id = 2
	enc.AddString(keyName, m.TraceId)

	return nil
}

func (m *ExecuteRequest) MarshalLogObject(enc go_uber_org_zap_zapcore.ObjectEncoder) error {
	var keyName string
	_ = keyName

	if m == nil {
		return nil
	}

	keyName = "block" // field block = 1
	if m.Block != nil {
		var vv interface{} = m.Block
		if marshaler, ok := vv.(go_uber_org_zap_zapcore.ObjectMarshaler); ok {
			enc.AddObject(keyName, marshaler)
		}
	}

	return nil
}

func (m *ExecuteResponse) MarshalLogObject(enc go_uber_org_zap_zapcore.ObjectEncoder) error {
	var keyName string
	_ = keyName

	if m == nil {
		return nil
	}

	keyName = "outputs" // field outputs = 1
	enc.AddArray(keyName, go_uber_org_zap_zapcore.ArrayMarshalerFunc(func(aenc go_uber_org_zap_zapcore.ArrayEncoder) error {
		for _, rv := range m.Outputs {
			_ = rv
			if rv != nil {
				var vv interface{} = rv
				if marshaler, ok := vv.(go_uber_org_zap_zapcore.ObjectMarshaler); ok {
					aenc.AppendObject(marshaler)
				}
			}
		}
		return nil
	}))

	return nil
}

func (m *StreamGenerateRequest) MarshalLogObject(enc go_uber_org_zap_zapcore.ObjectEncoder) error {
	var keyName string
	_ = keyName

	if m == nil {
		return nil
	}

	keyName = "full_context" // field full_context = 1
	if ov, ok := m.GetRequest().(*StreamGenerateRequest_FullContext); ok {
		_ = ov
		if ov.FullContext != nil {
			var vv interface{} = ov.FullContext
			if marshaler, ok := vv.(go_uber_org_zap_zapcore.ObjectMarshaler); ok {
				enc.AddObject(keyName, marshaler)
			}
		}
	}

	keyName = "update" // field update = 2
	if ov, ok := m.GetRequest().(*StreamGenerateRequest_Update); ok {
		_ = ov
		if ov.Update != nil {
			var vv interface{} = ov.Update
			if marshaler, ok := vv.(go_uber_org_zap_zapcore.ObjectMarshaler); ok {
				enc.AddObject(keyName, marshaler)
			}
		}
	}

	return nil
}

func (m *FullContext) MarshalLogObject(enc go_uber_org_zap_zapcore.ObjectEncoder) error {
	var keyName string
	_ = keyName

	if m == nil {
		return nil
	}

	keyName = "doc" // field doc = 1
	if m.Doc != nil {
		var vv interface{} = m.Doc
		if marshaler, ok := vv.(go_uber_org_zap_zapcore.ObjectMarshaler); ok {
			enc.AddObject(keyName, marshaler)
		}
	}

	keyName = "selected" // field selected = 2
	enc.AddInt32(keyName, m.Selected)

	return nil
}

func (m *BlockUpdate) MarshalLogObject(enc go_uber_org_zap_zapcore.ObjectEncoder) error {
	var keyName string
	_ = keyName

	if m == nil {
		return nil
	}

	keyName = "block_id" // field block_id = 1
	enc.AddString(keyName, m.BlockId)

	keyName = "block_content" // field block_content = 2
	enc.AddString(keyName, m.BlockContent)

	return nil
}

func (m *Finish) MarshalLogObject(enc go_uber_org_zap_zapcore.ObjectEncoder) error {
	var keyName string
	_ = keyName

	if m == nil {
		return nil
	}

	keyName = "accepted" // field accepted = 1
	enc.AddBool(keyName, m.Accepted)

	return nil
}

func (m *StreamGenerateResponse) MarshalLogObject(enc go_uber_org_zap_zapcore.ObjectEncoder) error {
	var keyName string
	_ = keyName

	if m == nil {
		return nil
	}

	keyName = "blocks" // field blocks = 1
	enc.AddArray(keyName, go_uber_org_zap_zapcore.ArrayMarshalerFunc(func(aenc go_uber_org_zap_zapcore.ArrayEncoder) error {
		for _, rv := range m.Blocks {
			_ = rv
			if rv != nil {
				var vv interface{} = rv
				if marshaler, ok := vv.(go_uber_org_zap_zapcore.ObjectMarshaler); ok {
					aenc.AppendObject(marshaler)
				}
			}
		}
		return nil
	}))

	return nil
}
