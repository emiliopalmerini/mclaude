package otel

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/emiliopalmerini/mclaude/internal/adapters/turso"
	"github.com/emiliopalmerini/mclaude/internal/domain"
	collectormetrics "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	commonv1 "go.opentelemetry.io/proto/otlp/common/v1"
	metricsv1 "go.opentelemetry.io/proto/otlp/metrics/v1"
	"google.golang.org/protobuf/proto"
)

type Receiver struct {
	metricsRepo *turso.UsageMetricsRepository
}

func NewReceiver(metricsRepo *turso.UsageMetricsRepository) *Receiver {
	return &Receiver{
		metricsRepo: metricsRepo,
	}
}

func (r *Receiver) HandleRequest(ctx context.Context, body io.Reader, contentType string) error {
	data, err := io.ReadAll(body)
	if err != nil {
		return fmt.Errorf("failed to read request body: %w", err)
	}

	var req collectormetrics.ExportMetricsServiceRequest

	if contentType == "application/json" {
		if err := json.Unmarshal(data, &req); err != nil {
			return fmt.Errorf("failed to parse JSON metrics: %w", err)
		}
	} else {
		if err := proto.Unmarshal(data, &req); err != nil {
			return fmt.Errorf("failed to parse protobuf metrics: %w", err)
		}
	}

	return r.processMetrics(ctx, &req)
}

func (r *Receiver) processMetrics(ctx context.Context, req *collectormetrics.ExportMetricsServiceRequest) error {
	for _, resourceMetrics := range req.GetResourceMetrics() {
		for _, scopeMetrics := range resourceMetrics.GetScopeMetrics() {
			for _, metric := range scopeMetrics.GetMetrics() {
				if !strings.HasPrefix(metric.GetName(), "claude_code.") {
					continue
				}

				if err := r.processMetric(ctx, metric); err != nil {
					log.Printf("Error processing metric %s: %v", metric.GetName(), err)
				}
			}
		}
	}
	return nil
}

func (r *Receiver) processMetric(ctx context.Context, metric *metricsv1.Metric) error {
	metricName := metric.GetName()

	// Handle Sum metrics (counters)
	if sum := metric.GetSum(); sum != nil {
		for _, dp := range sum.GetDataPoints() {
			value := r.getDataPointValue(dp)
			attrs := r.extractAttributes(dp.GetAttributes())
			if err := r.storeMetric(ctx, metricName, value, attrs); err != nil {
				return err
			}
		}
	}

	// Handle Gauge metrics
	if gauge := metric.GetGauge(); gauge != nil {
		for _, dp := range gauge.GetDataPoints() {
			value := r.getDataPointValue(dp)
			attrs := r.extractAttributes(dp.GetAttributes())
			if err := r.storeMetric(ctx, metricName, value, attrs); err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *Receiver) getDataPointValue(dp *metricsv1.NumberDataPoint) float64 {
	switch v := dp.GetValue().(type) {
	case *metricsv1.NumberDataPoint_AsDouble:
		return v.AsDouble
	case *metricsv1.NumberDataPoint_AsInt:
		return float64(v.AsInt)
	default:
		return 0
	}
}

func (r *Receiver) extractAttributes(attrs []*commonv1.KeyValue) *string {
	if len(attrs) == 0 {
		return nil
	}
	// Convert to a simple map for JSON serialization
	attrMap := make(map[string]interface{})
	for _, kv := range attrs {
		key := kv.GetKey()
		val := kv.GetValue()
		if val != nil {
			switch v := val.GetValue().(type) {
			case *commonv1.AnyValue_StringValue:
				attrMap[key] = v.StringValue
			case *commonv1.AnyValue_IntValue:
				attrMap[key] = v.IntValue
			case *commonv1.AnyValue_DoubleValue:
				attrMap[key] = v.DoubleValue
			case *commonv1.AnyValue_BoolValue:
				attrMap[key] = v.BoolValue
			}
		}
	}
	if len(attrMap) == 0 {
		return nil
	}
	data, err := json.Marshal(attrMap)
	if err != nil {
		return nil
	}
	s := string(data)
	return &s
}

func (r *Receiver) storeMetric(ctx context.Context, name string, value float64, attributes *string) error {
	metric := &domain.UsageMetric{
		MetricName: name,
		Value:      value,
		Attributes: attributes,
		RecordedAt: time.Now(),
	}
	log.Printf("Storing metric: %s = %f", name, value)
	return r.metricsRepo.Create(ctx, metric)
}
