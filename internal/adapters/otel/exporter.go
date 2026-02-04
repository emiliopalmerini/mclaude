package otel

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/emiliopalmerini/mclaude/internal/ports"
)

const (
	serviceName    = "mclaude"
	serviceVersion = "1.0.0"
)

// Exporter exports session metrics to an OTEL Collector.
type Exporter struct {
	provider       *sdkmetric.MeterProvider
	meter          metric.Meter
	tokensTotal    metric.Int64Counter
	costTotal      metric.Float64Counter
	durationHist   metric.Float64Histogram
	turnsHist      metric.Int64Histogram
	sessionsTotal  metric.Int64Counter
}

// NewExporter creates a new OTEL metrics exporter.
func NewExporter(ctx context.Context, cfg Config) (*Exporter, error) {
	if !cfg.Enabled || cfg.Endpoint == "" {
		return nil, fmt.Errorf("OTEL exporter is disabled or endpoint not configured")
	}

	opts := []otlpmetricgrpc.Option{
		otlpmetricgrpc.WithEndpoint(cfg.Endpoint),
	}
	if cfg.Insecure {
		opts = append(opts, otlpmetricgrpc.WithDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())))
		opts = append(opts, otlpmetricgrpc.WithInsecure())
	}

	exp, err := otlpmetricgrpc.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("creating OTLP exporter: %w", err)
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(serviceVersion),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("creating resource: %w", err)
	}

	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exp)),
		sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(provider)

	meter := provider.Meter(serviceName)

	tokensTotal, err := meter.Int64Counter(
		"mclaude_session_tokens_total",
		metric.WithDescription("Total tokens used in sessions"),
		metric.WithUnit("{token}"),
	)
	if err != nil {
		return nil, fmt.Errorf("creating tokens counter: %w", err)
	}

	costTotal, err := meter.Float64Counter(
		"mclaude_session_cost_usd",
		metric.WithDescription("Total estimated cost in USD"),
		metric.WithUnit("USD"),
	)
	if err != nil {
		return nil, fmt.Errorf("creating cost counter: %w", err)
	}

	durationHist, err := meter.Float64Histogram(
		"mclaude_session_duration_seconds",
		metric.WithDescription("Session duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, fmt.Errorf("creating duration histogram: %w", err)
	}

	turnsHist, err := meter.Int64Histogram(
		"mclaude_session_turns",
		metric.WithDescription("Number of turns per session"),
		metric.WithUnit("{turn}"),
	)
	if err != nil {
		return nil, fmt.Errorf("creating turns histogram: %w", err)
	}

	sessionsTotal, err := meter.Int64Counter(
		"mclaude_sessions_total",
		metric.WithDescription("Total number of sessions"),
		metric.WithUnit("{session}"),
	)
	if err != nil {
		return nil, fmt.Errorf("creating sessions counter: %w", err)
	}

	return &Exporter{
		provider:      provider,
		meter:         meter,
		tokensTotal:   tokensTotal,
		costTotal:     costTotal,
		durationHist:  durationHist,
		turnsHist:     turnsHist,
		sessionsTotal: sessionsTotal,
	}, nil
}

// ExportSessionMetrics exports enriched metrics for a completed session.
func (e *Exporter) ExportSessionMetrics(ctx context.Context, m *ports.EnrichedMetrics) error {
	attrs := []attribute.KeyValue{
		attribute.String("project_id", m.ProjectID),
		attribute.String("project_name", m.ProjectName),
		attribute.String("exit_reason", m.ExitReason),
	}
	if m.ExperimentID != nil {
		attrs = append(attrs, attribute.String("experiment_id", *m.ExperimentID))
	}
	if m.ExperimentName != nil {
		attrs = append(attrs, attribute.String("experiment_name", *m.ExperimentName))
	}

	opt := metric.WithAttributes(attrs...)

	// Record token metrics
	totalTokens := m.TokenInput + m.TokenOutput
	e.tokensTotal.Add(ctx, totalTokens, opt)

	// Record cost
	e.costTotal.Add(ctx, m.CostEstimateUSD, opt)

	// Record duration
	e.durationHist.Record(ctx, float64(m.DurationSeconds), opt)

	// Record turns
	e.turnsHist.Record(ctx, m.TurnCount, opt)

	// Increment session count
	e.sessionsTotal.Add(ctx, 1, opt)

	return nil
}

// Close shuts down the exporter and flushes any pending metrics.
func (e *Exporter) Close(ctx context.Context) error {
	return e.provider.Shutdown(ctx)
}
