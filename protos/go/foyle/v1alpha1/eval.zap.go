// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: foyle/v1alpha1/eval.proto

package v1alpha1

import (
	fmt "fmt"
	math "math"
	proto "github.com/golang/protobuf/proto"
	_ "github.com/stateful/runme/v3/pkg/api/gen/proto/go/runme/parser/v1"
	_ "google.golang.org/protobuf/types/known/timestamppb"
	_ "google.golang.org/protobuf/types/known/structpb"
	go_uber_org_zap_zapcore "go.uber.org/zap/zapcore"
	github_com_golang_protobuf_ptypes "github.com/golang/protobuf/ptypes"
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

	keyName = "actual_cells" // field actual_cells = 11
	enc.AddArray(keyName, go_uber_org_zap_zapcore.ArrayMarshalerFunc(func(aenc go_uber_org_zap_zapcore.ArrayEncoder) error {
		for _, rv := range m.ActualCells {
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

	keyName = "error" // field error = 5
	enc.AddString(keyName, m.Error)

	keyName = "status" // field status = 6
	enc.AddString(keyName, m.Status.String())

	keyName = "gen_trace_id" // field gen_trace_id = 8
	enc.AddString(keyName, m.GenTraceId)

	keyName = "best_rag_result" // field best_rag_result = 9
	if m.BestRagResult != nil {
		var vv interface{} = m.BestRagResult
		if marshaler, ok := vv.(go_uber_org_zap_zapcore.ObjectMarshaler); ok {
			enc.AddObject(keyName, marshaler)
		}
	}

	keyName = "assertions" // field assertions = 10
	enc.AddArray(keyName, go_uber_org_zap_zapcore.ArrayMarshalerFunc(func(aenc go_uber_org_zap_zapcore.ArrayEncoder) error {
		for _, rv := range m.Assertions {
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

	keyName = "cells_match_result" // field cells_match_result = 12
	enc.AddString(keyName, m.CellsMatchResult.String())

	keyName = "judge_explanation" // field judge_explanation = 13
	enc.AddString(keyName, m.JudgeExplanation)

	return nil
}

func (m *Assertion) MarshalLogObject(enc go_uber_org_zap_zapcore.ObjectEncoder) error {
	var keyName string
	_ = keyName

	if m == nil {
		return nil
	}

	keyName = "name" // field name = 1
	enc.AddString(keyName, m.Name)

	keyName = "result" // field result = 2
	enc.AddString(keyName, m.Result.String())

	keyName = "detail" // field detail = 3
	enc.AddString(keyName, m.Detail)

	return nil
}

func (m *EvalResultListRequest) MarshalLogObject(enc go_uber_org_zap_zapcore.ObjectEncoder) error {
	var keyName string
	_ = keyName

	if m == nil {
		return nil
	}

	keyName = "database" // field database = 1
	enc.AddString(keyName, m.Database)

	return nil
}

func (m *EvalResultListResponse) MarshalLogObject(enc go_uber_org_zap_zapcore.ObjectEncoder) error {
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

	return nil
}

func (m *AssertionRow) MarshalLogObject(enc go_uber_org_zap_zapcore.ObjectEncoder) error {
	var keyName string
	_ = keyName

	if m == nil {
		return nil
	}

	keyName = "id" // field id = 1
	enc.AddString(keyName, m.Id)

	keyName = "exampleFile" // field exampleFile = 2
	enc.AddString(keyName, m.ExampleFile)

	keyName = "doc_md" // field doc_md = 3
	enc.AddString(keyName, m.DocMd)

	keyName = "answer_md" // field answer_md = 4
	enc.AddString(keyName, m.AnswerMd)

	keyName = "code_after_markdown" // field code_after_markdown = 5
	enc.AddString(keyName, m.CodeAfterMarkdown.String())

	keyName = "one_code_cell" // field one_code_cell = 6
	enc.AddString(keyName, m.OneCodeCell.String())

	keyName = "ends_with_code_cell" // field ends_with_code_cell = 7
	enc.AddString(keyName, m.EndsWithCodeCell.String())

	return nil
}

func (m *AssertionTableRequest) MarshalLogObject(enc go_uber_org_zap_zapcore.ObjectEncoder) error {
	var keyName string
	_ = keyName

	if m == nil {
		return nil
	}

	keyName = "database" // field database = 1
	enc.AddString(keyName, m.Database)

	return nil
}

func (m *EvalExample) MarshalLogObject(enc go_uber_org_zap_zapcore.ObjectEncoder) error {
	var keyName string
	_ = keyName

	if m == nil {
		return nil
	}

	keyName = "id" // field id = 1
	enc.AddString(keyName, m.Id)

	keyName = "time" // field time = 4
	if t, err := github_com_golang_protobuf_ptypes.Timestamp(m.Time); err == nil {
		enc.AddTime(keyName, t)
	}

	keyName = "full_context" // field full_context = 2
	if m.FullContext != nil {
		var vv interface{} = m.FullContext
		if marshaler, ok := vv.(go_uber_org_zap_zapcore.ObjectMarshaler); ok {
			enc.AddObject(keyName, marshaler)
		}
	}

	keyName = "expected_cells" // field expected_cells = 3
	enc.AddArray(keyName, go_uber_org_zap_zapcore.ArrayMarshalerFunc(func(aenc go_uber_org_zap_zapcore.ArrayEncoder) error {
		for _, rv := range m.ExpectedCells {
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

func (m *AssertionTableResponse) MarshalLogObject(enc go_uber_org_zap_zapcore.ObjectEncoder) error {
	var keyName string
	_ = keyName

	if m == nil {
		return nil
	}

	keyName = "rows" // field rows = 1
	enc.AddArray(keyName, go_uber_org_zap_zapcore.ArrayMarshalerFunc(func(aenc go_uber_org_zap_zapcore.ArrayEncoder) error {
		for _, rv := range m.Rows {
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

func (m *GetEvalResultRequest) MarshalLogObject(enc go_uber_org_zap_zapcore.ObjectEncoder) error {
	var keyName string
	_ = keyName

	if m == nil {
		return nil
	}

	keyName = "id" // field id = 1
	enc.AddString(keyName, m.Id)

	return nil
}

func (m *GetEvalResultResponse) MarshalLogObject(enc go_uber_org_zap_zapcore.ObjectEncoder) error {
	var keyName string
	_ = keyName

	if m == nil {
		return nil
	}

	keyName = "reportHTML" // field reportHTML = 1
	enc.AddString(keyName, m.ReportHTML)

	return nil
}
