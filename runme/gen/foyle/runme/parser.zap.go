// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: foyle/runme/parser.proto

package runme

import (
	fmt "fmt"
	math "math"
	proto "github.com/golang/protobuf/proto"
	_ "google.golang.org/protobuf/types/known/wrapperspb"
	go_uber_org_zap_zapcore "go.uber.org/zap/zapcore"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

func (m *Notebook) MarshalLogObject(enc go_uber_org_zap_zapcore.ObjectEncoder) error {
	var keyName string
	_ = keyName

	if m == nil {
		return nil
	}

	keyName = "cells" // field cells = 1
	enc.AddArray(keyName, go_uber_org_zap_zapcore.ArrayMarshalerFunc(func(aenc go_uber_org_zap_zapcore.ArrayEncoder) error {
		for _, rv := range m.Cells {
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

	keyName = "metadata" // field metadata = 2
	enc.AddObject(keyName, go_uber_org_zap_zapcore.ObjectMarshalerFunc(func(enc go_uber_org_zap_zapcore.ObjectEncoder) error {
		for mk, mv := range m.Metadata {
			key := mk
			_ = key
			enc.AddString(key, mv)
		}
		return nil
	}))

	keyName = "frontmatter" // field frontmatter = 3
	if m.Frontmatter != nil {
		var vv interface{} = m.Frontmatter
		if marshaler, ok := vv.(go_uber_org_zap_zapcore.ObjectMarshaler); ok {
			enc.AddObject(keyName, marshaler)
		}
	}

	return nil
}

func (m *ExecutionSummaryTiming) MarshalLogObject(enc go_uber_org_zap_zapcore.ObjectEncoder) error {
	var keyName string
	_ = keyName

	if m == nil {
		return nil
	}

	keyName = "start_time" // field start_time = 1
	if m.StartTime != nil {
		var vv interface{} = m.StartTime
		if marshaler, ok := vv.(go_uber_org_zap_zapcore.ObjectMarshaler); ok {
			enc.AddObject(keyName, marshaler)
		}
	}

	keyName = "end_time" // field end_time = 2
	if m.EndTime != nil {
		var vv interface{} = m.EndTime
		if marshaler, ok := vv.(go_uber_org_zap_zapcore.ObjectMarshaler); ok {
			enc.AddObject(keyName, marshaler)
		}
	}

	return nil
}

func (m *CellOutputItem) MarshalLogObject(enc go_uber_org_zap_zapcore.ObjectEncoder) error {
	var keyName string
	_ = keyName

	if m == nil {
		return nil
	}

	keyName = "data" // field data = 1
	enc.AddByteString(keyName, m.Data)

	keyName = "type" // field type = 2
	enc.AddString(keyName, m.Type)

	keyName = "mime" // field mime = 3
	enc.AddString(keyName, m.Mime)

	return nil
}

func (m *ProcessInfoExitReason) MarshalLogObject(enc go_uber_org_zap_zapcore.ObjectEncoder) error {
	var keyName string
	_ = keyName

	if m == nil {
		return nil
	}

	keyName = "type" // field type = 1
	enc.AddString(keyName, m.Type)

	keyName = "code" // field code = 2
	if m.Code != nil {
		var vv interface{} = m.Code
		if marshaler, ok := vv.(go_uber_org_zap_zapcore.ObjectMarshaler); ok {
			enc.AddObject(keyName, marshaler)
		}
	}

	return nil
}

func (m *CellOutputProcessInfo) MarshalLogObject(enc go_uber_org_zap_zapcore.ObjectEncoder) error {
	var keyName string
	_ = keyName

	if m == nil {
		return nil
	}

	keyName = "exit_reason" // field exit_reason = 1
	if m.ExitReason != nil {
		var vv interface{} = m.ExitReason
		if marshaler, ok := vv.(go_uber_org_zap_zapcore.ObjectMarshaler); ok {
			enc.AddObject(keyName, marshaler)
		}
	}

	keyName = "pid" // field pid = 2
	if m.Pid != nil {
		var vv interface{} = m.Pid
		if marshaler, ok := vv.(go_uber_org_zap_zapcore.ObjectMarshaler); ok {
			enc.AddObject(keyName, marshaler)
		}
	}

	return nil
}

func (m *CellOutput) MarshalLogObject(enc go_uber_org_zap_zapcore.ObjectEncoder) error {
	var keyName string
	_ = keyName

	if m == nil {
		return nil
	}

	keyName = "items" // field items = 1
	enc.AddArray(keyName, go_uber_org_zap_zapcore.ArrayMarshalerFunc(func(aenc go_uber_org_zap_zapcore.ArrayEncoder) error {
		for _, rv := range m.Items {
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

	keyName = "metadata" // field metadata = 2
	enc.AddObject(keyName, go_uber_org_zap_zapcore.ObjectMarshalerFunc(func(enc go_uber_org_zap_zapcore.ObjectEncoder) error {
		for mk, mv := range m.Metadata {
			key := mk
			_ = key
			enc.AddString(key, mv)
		}
		return nil
	}))

	keyName = "process_info" // field process_info = 3
	if m.ProcessInfo != nil {
		var vv interface{} = m.ProcessInfo
		if marshaler, ok := vv.(go_uber_org_zap_zapcore.ObjectMarshaler); ok {
			enc.AddObject(keyName, marshaler)
		}
	}

	return nil
}

func (m *CellExecutionSummary) MarshalLogObject(enc go_uber_org_zap_zapcore.ObjectEncoder) error {
	var keyName string
	_ = keyName

	if m == nil {
		return nil
	}

	keyName = "execution_order" // field execution_order = 1
	if m.ExecutionOrder != nil {
		var vv interface{} = m.ExecutionOrder
		if marshaler, ok := vv.(go_uber_org_zap_zapcore.ObjectMarshaler); ok {
			enc.AddObject(keyName, marshaler)
		}
	}

	keyName = "success" // field success = 2
	if m.Success != nil {
		var vv interface{} = m.Success
		if marshaler, ok := vv.(go_uber_org_zap_zapcore.ObjectMarshaler); ok {
			enc.AddObject(keyName, marshaler)
		}
	}

	keyName = "timing" // field timing = 3
	if m.Timing != nil {
		var vv interface{} = m.Timing
		if marshaler, ok := vv.(go_uber_org_zap_zapcore.ObjectMarshaler); ok {
			enc.AddObject(keyName, marshaler)
		}
	}

	return nil
}

func (m *TextRange) MarshalLogObject(enc go_uber_org_zap_zapcore.ObjectEncoder) error {
	var keyName string
	_ = keyName

	if m == nil {
		return nil
	}

	keyName = "start" // field start = 1
	enc.AddUint32(keyName, m.Start)

	keyName = "end" // field end = 2
	enc.AddUint32(keyName, m.End)

	return nil
}

func (m *Cell) MarshalLogObject(enc go_uber_org_zap_zapcore.ObjectEncoder) error {
	var keyName string
	_ = keyName

	if m == nil {
		return nil
	}

	keyName = "kind" // field kind = 1
	enc.AddString(keyName, m.Kind.String())

	keyName = "value" // field value = 2
	enc.AddString(keyName, m.Value)

	keyName = "language_id" // field language_id = 3
	enc.AddString(keyName, m.LanguageId)

	keyName = "metadata" // field metadata = 4
	enc.AddObject(keyName, go_uber_org_zap_zapcore.ObjectMarshalerFunc(func(enc go_uber_org_zap_zapcore.ObjectEncoder) error {
		for mk, mv := range m.Metadata {
			key := mk
			_ = key
			enc.AddString(key, mv)
		}
		return nil
	}))

	keyName = "text_range" // field text_range = 5
	if m.TextRange != nil {
		var vv interface{} = m.TextRange
		if marshaler, ok := vv.(go_uber_org_zap_zapcore.ObjectMarshaler); ok {
			enc.AddObject(keyName, marshaler)
		}
	}

	keyName = "outputs" // field outputs = 6
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

	keyName = "execution_summary" // field execution_summary = 7
	if m.ExecutionSummary != nil {
		var vv interface{} = m.ExecutionSummary
		if marshaler, ok := vv.(go_uber_org_zap_zapcore.ObjectMarshaler); ok {
			enc.AddObject(keyName, marshaler)
		}
	}

	return nil
}

func (m *RunmeSessionDocument) MarshalLogObject(enc go_uber_org_zap_zapcore.ObjectEncoder) error {
	var keyName string
	_ = keyName

	if m == nil {
		return nil
	}

	keyName = "relative_path" // field relative_path = 1
	enc.AddString(keyName, m.RelativePath)

	return nil
}

func (m *RunmeSession) MarshalLogObject(enc go_uber_org_zap_zapcore.ObjectEncoder) error {
	var keyName string
	_ = keyName

	if m == nil {
		return nil
	}

	keyName = "id" // field id = 1
	enc.AddString(keyName, m.Id)

	keyName = "document" // field document = 2
	if m.Document != nil {
		var vv interface{} = m.Document
		if marshaler, ok := vv.(go_uber_org_zap_zapcore.ObjectMarshaler); ok {
			enc.AddObject(keyName, marshaler)
		}
	}

	return nil
}

func (m *FrontmatterRunme) MarshalLogObject(enc go_uber_org_zap_zapcore.ObjectEncoder) error {
	var keyName string
	_ = keyName

	if m == nil {
		return nil
	}

	keyName = "id" // field id = 1
	enc.AddString(keyName, m.Id)

	keyName = "version" // field version = 2
	enc.AddString(keyName, m.Version)

	keyName = "session" // field session = 3
	if m.Session != nil {
		var vv interface{} = m.Session
		if marshaler, ok := vv.(go_uber_org_zap_zapcore.ObjectMarshaler); ok {
			enc.AddObject(keyName, marshaler)
		}
	}

	return nil
}

func (m *Frontmatter) MarshalLogObject(enc go_uber_org_zap_zapcore.ObjectEncoder) error {
	var keyName string
	_ = keyName

	if m == nil {
		return nil
	}

	keyName = "shell" // field shell = 1
	enc.AddString(keyName, m.Shell)

	keyName = "cwd" // field cwd = 2
	enc.AddString(keyName, m.Cwd)

	keyName = "skip_prompts" // field skip_prompts = 3
	enc.AddBool(keyName, m.SkipPrompts)

	keyName = "runme" // field runme = 4
	if m.Runme != nil {
		var vv interface{} = m.Runme
		if marshaler, ok := vv.(go_uber_org_zap_zapcore.ObjectMarshaler); ok {
			enc.AddObject(keyName, marshaler)
		}
	}

	keyName = "category" // field category = 5
	enc.AddString(keyName, m.Category)

	return nil
}

func (m *DeserializeRequestOptions) MarshalLogObject(enc go_uber_org_zap_zapcore.ObjectEncoder) error {
	var keyName string
	_ = keyName

	if m == nil {
		return nil
	}

	keyName = "identity" // field identity = 1
	enc.AddString(keyName, m.Identity.String())

	return nil
}

func (m *DeserializeRequest) MarshalLogObject(enc go_uber_org_zap_zapcore.ObjectEncoder) error {
	var keyName string
	_ = keyName

	if m == nil {
		return nil
	}

	keyName = "source" // field source = 1
	enc.AddByteString(keyName, m.Source)

	keyName = "options" // field options = 2
	if m.Options != nil {
		var vv interface{} = m.Options
		if marshaler, ok := vv.(go_uber_org_zap_zapcore.ObjectMarshaler); ok {
			enc.AddObject(keyName, marshaler)
		}
	}

	return nil
}

func (m *DeserializeResponse) MarshalLogObject(enc go_uber_org_zap_zapcore.ObjectEncoder) error {
	var keyName string
	_ = keyName

	if m == nil {
		return nil
	}

	keyName = "notebook" // field notebook = 1
	if m.Notebook != nil {
		var vv interface{} = m.Notebook
		if marshaler, ok := vv.(go_uber_org_zap_zapcore.ObjectMarshaler); ok {
			enc.AddObject(keyName, marshaler)
		}
	}

	return nil
}

func (m *SerializeRequestOutputOptions) MarshalLogObject(enc go_uber_org_zap_zapcore.ObjectEncoder) error {
	var keyName string
	_ = keyName

	if m == nil {
		return nil
	}

	keyName = "enabled" // field enabled = 1
	enc.AddBool(keyName, m.Enabled)

	keyName = "summary" // field summary = 2
	enc.AddBool(keyName, m.Summary)

	return nil
}

func (m *SerializeRequestOptions) MarshalLogObject(enc go_uber_org_zap_zapcore.ObjectEncoder) error {
	var keyName string
	_ = keyName

	if m == nil {
		return nil
	}

	keyName = "outputs" // field outputs = 1
	if m.Outputs != nil {
		var vv interface{} = m.Outputs
		if marshaler, ok := vv.(go_uber_org_zap_zapcore.ObjectMarshaler); ok {
			enc.AddObject(keyName, marshaler)
		}
	}

	keyName = "session" // field session = 2
	if m.Session != nil {
		var vv interface{} = m.Session
		if marshaler, ok := vv.(go_uber_org_zap_zapcore.ObjectMarshaler); ok {
			enc.AddObject(keyName, marshaler)
		}
	}

	return nil
}

func (m *SerializeRequest) MarshalLogObject(enc go_uber_org_zap_zapcore.ObjectEncoder) error {
	var keyName string
	_ = keyName

	if m == nil {
		return nil
	}

	keyName = "notebook" // field notebook = 1
	if m.Notebook != nil {
		var vv interface{} = m.Notebook
		if marshaler, ok := vv.(go_uber_org_zap_zapcore.ObjectMarshaler); ok {
			enc.AddObject(keyName, marshaler)
		}
	}

	keyName = "options" // field options = 2
	if m.Options != nil {
		var vv interface{} = m.Options
		if marshaler, ok := vv.(go_uber_org_zap_zapcore.ObjectMarshaler); ok {
			enc.AddObject(keyName, marshaler)
		}
	}

	return nil
}

func (m *SerializeResponse) MarshalLogObject(enc go_uber_org_zap_zapcore.ObjectEncoder) error {
	var keyName string
	_ = keyName

	if m == nil {
		return nil
	}

	keyName = "result" // field result = 1
	enc.AddByteString(keyName, m.Result)

	return nil
}