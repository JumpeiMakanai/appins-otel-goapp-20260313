package service

import (
	"context"
	"errors"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

func ExecuteBusiness(ctx context.Context, shouldFail bool) error {
	tracer := otel.Tracer("go-app/internal/service")

	_, span := tracer.Start(ctx, "business_logic")
	defer span.End()

	span.SetAttributes(
		attribute.Bool("business.should_fail", shouldFail),
	)

	time.Sleep(150 * time.Millisecond)

	if shouldFail {
		err := errors.New("business logic failed")
		span.RecordError(err)
		span.SetStatus(codes.Error, "business logic failed")
		return err
	}

	span.SetStatus(codes.Ok, "success")
	return nil
}