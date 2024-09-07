package oai

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

func tracer() trace.Tracer {
	return otel.Tracer("github.com/jlewi/foyle/app/pkg/oai")
}
