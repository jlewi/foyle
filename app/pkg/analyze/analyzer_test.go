package analyze

import (
	"context"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/jlewi/foyle/app/pkg/testutil"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	"math/rand"
	"testing"
	"time"
)

func timeMustParse(layoutString, value string) time.Time {
	t, err := time.Parse(layoutString, value)
	if err != nil {
		panic(err)
	}
	return t
}

func shuffle(in []string) []string {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(in), func(i, j int) { in[i], in[j] = in[j], in[i] })
	return in
}

func Test_BuildBlockLog(t *testing.T) {
	type testCase struct {
		name     string
		block    *BlockLog
		traces   map[string]Trace
		expected *BlockLog
	}

	traces := make(map[string]Trace)

	const bid1 = "g123output1"
	genTrace := &GenerateTrace{
		TraceID:   "g123",
		StartTime: timeMustParse(time.RFC3339, "2021-01-01T00:00:00Z"),
		EndTime:   timeMustParse(time.RFC3339, "2021-01-01T00:01:00Z"),
		Request: &v1alpha1.GenerateRequest{
			Doc: &v1alpha1.Doc{
				Blocks: []*v1alpha1.Block{
					{
						Contents: "echo hello",
					},
				},
			},
		},
		Response: &v1alpha1.GenerateResponse{
			Blocks: []*v1alpha1.Block{
				{
					Id:       bid1,
					Contents: "outcell",
				},
			},
		},
	}

	execTrace1 := &ExecuteTrace{
		TraceID:   "e456",
		StartTime: timeMustParse(time.RFC3339, "2021-01-02T00:00:00Z"),
		EndTime:   timeMustParse(time.RFC3339, "2021-01-02T00:01:00Z"),
		Request: &v1alpha1.ExecuteRequest{
			Block: &v1alpha1.Block{
				Contents: "echo hello",
				Id:       bid1,
			},
		},
		Response: &v1alpha1.ExecuteResponse{
			Outputs: []*v1alpha1.BlockOutput{
				{
					Items: []*v1alpha1.BlockOutputItem{
						{
							TextData: "exitCode: 4",
						},
					},
				},
			},
		},
	}

	execTrace2 := &ExecuteTrace{
		TraceID:   "e789",
		StartTime: timeMustParse(time.RFC3339, "2021-01-03T00:00:00Z"),
		EndTime:   timeMustParse(time.RFC3339, "2021-01-03T00:01:00Z"),
		Request: &v1alpha1.ExecuteRequest{
			Block: &v1alpha1.Block{
				Contents: "echo hello",
				Id:       bid1,
			},
		},
		Response: &v1alpha1.ExecuteResponse{
			Outputs: []*v1alpha1.BlockOutput{
				{
					Items: []*v1alpha1.BlockOutputItem{
						{
							TextData: "exitCode: 7",
						},
					},
				},
			},
		},
	}

	traces[genTrace.TraceID] = genTrace
	traces[execTrace1.TraceID] = execTrace1
	traces[execTrace2.TraceID] = execTrace2

	// We shuffle ExecTraceIds to make sure we properly set block log based on the later trace
	execTraceIds := shuffle([]string{execTrace1.TraceID, execTrace2.TraceID})
	cases := []testCase{
		{
			name: "basic",
			block: &BlockLog{
				ID:         bid1,
				GenTraceID: genTrace.TraceID,

				ExecTraceIDs: execTraceIds,
			},
			expected: &BlockLog{
				ID:             bid1,
				GenTraceID:     genTrace.TraceID,
				ExecTraceIDs:   execTraceIds,
				Doc:            genTrace.Request.Doc,
				GeneratedBlock: genTrace.Response.Blocks[0],
				ExecutedBlock:  execTrace2.Request.Block,
				ExitCode:       7,
			},
			traces: traces,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if err := buildBlockLog(context.Background(), c.block, c.traces); err != nil {
				t.Fatalf("buildBlockLog failed: %v", err)
			}

			if d := cmp.Diff(c.expected, c.block, testutil.BlockComparer, cmpopts.IgnoreUnexported(v1alpha1.Doc{})); d != "" {
				t.Errorf("Unexpected diff:\n%s", d)
			}
		})
	}
}
