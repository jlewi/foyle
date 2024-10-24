// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: foyle/logs/conversion.proto

package logspb

import (
	fmt "fmt"
	math "math"
	proto "github.com/golang/protobuf/proto"
	_ "google.golang.org/protobuf/types/known/structpb"
	_ "google.golang.org/protobuf/types/known/timestamppb"
	_ "github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	go_uber_org_zap_zapcore "go.uber.org/zap/zapcore"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

func (m *ConvertDocRequest) MarshalLogObject(enc go_uber_org_zap_zapcore.ObjectEncoder) error {
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

	keyName = "format" // field format = 2
	enc.AddString(keyName, m.Format.String())

	return nil
}

func (m *ConvertDocResponse) MarshalLogObject(enc go_uber_org_zap_zapcore.ObjectEncoder) error {
	var keyName string
	_ = keyName

	if m == nil {
		return nil
	}

	keyName = "text" // field text = 1
	enc.AddString(keyName, m.Text)

	return nil
}
