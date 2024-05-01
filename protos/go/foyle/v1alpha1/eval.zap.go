// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: foyle/v1alpha1/eval.proto

package v1alpha1

import (
	fmt "fmt"
	math "math"
	proto "github.com/golang/protobuf/proto"
	_ "google.golang.org/protobuf/types/known/structpb"
	go_uber_org_zap_zapcore "go.uber.org/zap/zapcore"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

func (m *EvalResult) MarshalLogObject(enc go_uber_org_zap_zapcore.ObjectEncoder) error {
	var keyName string
	_ = keyName

	if m == nil {
		return nil
	}

	keyName = "example" // field example = 1
	if m.Example != nil {
		var vv interface{} = m.Example
		if marshaler, ok := vv.(go_uber_org_zap_zapcore.ObjectMarshaler); ok {
			enc.AddObject(keyName, marshaler)
		}
	}

	keyName = "example_file" // field example_file = 2
	enc.AddString(keyName, m.ExampleFile)

	keyName = "actual" // field actual = 3
	enc.AddArray(keyName, go_uber_org_zap_zapcore.ArrayMarshalerFunc(func(aenc go_uber_org_zap_zapcore.ArrayEncoder) error {
		for _, rv := range m.Actual {
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

	keyName = "distance" // field distance = 4
	enc.AddInt32(keyName, m.Distance)

	keyName = "error" // field error = 5
	enc.AddString(keyName, m.Error)

	keyName = "status" // field status = 6
	enc.AddString(keyName, m.Status.String())

	return nil
}
