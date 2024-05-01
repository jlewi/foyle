package executor

import (
	"context"
	"fmt"
	"github.com/jlewi/foyle/app/pkg/config"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/go-logr/logr"
	"go.opentelemetry.io/otel/trace"

	"github.com/go-cmd/cmd"
	"github.com/jlewi/foyle/app/pkg/logs"
	"github.com/jlewi/foyle/protos/go/foyle/v1alpha1"
	"github.com/jlewi/monogo/helpers"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Executor is responsible for executing the commands
type Executor struct {
	v1alpha1.UnimplementedExecuteServiceServer
	p      *BashishParser
	config config.Config
}

func NewExecutor(cfg config.Config) (*Executor, error) {
	p, err := NewBashishParser()
	if err != nil {
		return nil, err
	}
	return &Executor{
		p:      p,
		config: cfg,
	}, nil
}

func (e *Executor) Execute(ctx context.Context, req *v1alpha1.ExecuteRequest) (*v1alpha1.ExecuteResponse, error) {
	span := trace.SpanFromContext(ctx)
	log := logs.FromContext(ctx)
	log = log.WithValues("traceId", span.SpanContext().TraceID(), "evalMode", e.config.EvalMode())
	ctx = logr.NewContext(ctx, log)

	log.Info("Executor.Execute", "blockId", req.GetBlock().GetId(), zap.Object("request", req))

	if req.GetBlock() == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Block is required")
	}
	instructions, err := e.p.Parse(req.GetBlock().GetContents())
	if err != nil {
		log.Error(err, "Failed to parse instructions")
		return nil, status.Errorf(codes.InvalidArgument, "Failed to parse instructions: %v", err)
	}

	if len(instructions) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "No instructions to execute")
	}

	result := e.executeInstructions(ctx, instructions)
	resp := resultToProto(result)

	log.Info("Executed instructions", "instructions", instructions, zap.Object("response", resp))
	return resp, nil
}

type result struct {
	exitCode int
	stdOut   string
	stdErr   string
}

func (e *Executor) executeInstructions(ctx context.Context, instructions []Instruction) result {
	log := logs.FromContext(ctx)
	// Set a deadline for the execution
	deadline, ok := ctx.Deadline()
	if !ok {
		// N.B. We set a short deadline because the UI doesn't handle long running commands very well right now
		deadline = time.Now().Add(15 * time.Minute)
	}

	// We use in to pipe the output of one instruction to the next instruction if necessary
	var in *strings.Reader

	// initialize to a random negative value to indicate it hasn't been set
	exitCode := -73
	stdOut := ""
	stdErr := ""

	for _, i := range instructions {
		var statusChan <-chan cmd.Status
		// Start the command in non blocking mode
		if in == nil {
			statusChan = i.Command.Start()
		} else {
			statusChan = i.Command.StartWithStdin(in)
		}

		select {
		case finalStatus := <-statusChan:
			if stdOut != "" {
				stdOut += "\n"
			}
			if stdErr != "" {
				stdErr += "\n"
			}
			stdErr += strings.Join(finalStatus.Stderr, "\n")

			// What we do with stdout depends on whether the output is piped to the next instruction or not
			out := strings.Join(finalStatus.Stdout, "\n")
			if i.Piped {
				in = strings.NewReader(out)
			} else {
				stdOut += out
			}

			if finalStatus.Exit != 0 {
				exitCode = finalStatus.Exit
				return result{
					exitCode: exitCode,
					stdOut:   stdOut,
					stdErr:   stdErr,
				}
			}

		case <-time.After(time.Until(deadline)):
			log.Info("instruction timed out", "instruction", helpers.CmdToString(*i.Command))
			if stdErr != "" {
				stdErr += "\n"
			}
			stdErr += fmt.Sprintf("instruction timed out; instruction was %s", helpers.CmdToString(*i.Command))
			return result{
				//  Set a random negative value to indicate the command timed out
				exitCode: -83,
				stdOut:   stdOut,
				stdErr:   stdErr,
			}
		}
	}

	return result{
		// Since everything was successful we set the exit code to 0
		exitCode: 0,
		stdOut:   stdOut,
		stdErr:   stdErr,
	}
}

func resultToProto(r result) *v1alpha1.ExecuteResponse {
	resp := &v1alpha1.ExecuteResponse{
		Outputs: make([]*v1alpha1.BlockOutput, 0, 3),
	}

	// N.B Is using separate outputs really the best way to go?

	resp.Outputs = append(resp.Outputs, &v1alpha1.BlockOutput{
		Items: []*v1alpha1.BlockOutputItem{
			{
				Mime:     MimePlainText,
				TextData: fmt.Sprintf("exitCode: %d", r.exitCode),
			},
		},
	})

	if r.stdOut != "" {
		resp.Outputs = append(resp.Outputs, &v1alpha1.BlockOutput{
			Items: []*v1alpha1.BlockOutputItem{
				{
					Mime:     MimePlainText,
					TextData: "stdout:\n" + r.stdOut,
				},
			},
		})
	}

	if r.stdErr != "" {
		resp.Outputs = append(resp.Outputs, &v1alpha1.BlockOutput{
			Items: []*v1alpha1.BlockOutputItem{
				{
					Mime:     MimePlainText,
					TextData: "stderr:\n" + r.stdErr,
				},
			},
		})
	}

	return resp
}
