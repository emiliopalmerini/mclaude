package prometheus

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/emiliopalmerini/mclaude/internal/ports"
)

const (
	// Metric names for Claude Code native OTEL metrics
	tokenMetric = "claude_code_token_usage_total"
	costMetric  = "claude_code_cost_usage_USD_total"
)

// Client queries Prometheus for real-time usage metrics.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new Prometheus client.
func NewClient(cfg Config) (*Client, error) {
	if !cfg.Enabled || cfg.URL == "" {
		return nil, fmt.Errorf("Prometheus client is disabled or URL not configured")
	}

	return &Client{
		baseURL: cfg.URL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}

// prometheusResponse represents the JSON response from Prometheus query API.
type prometheusResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric map[string]string `json:"metric"`
			Value  []interface{}     `json:"value"`
		} `json:"result"`
	} `json:"data"`
}

// GetRollingWindowUsage retrieves aggregated usage for the specified rolling window.
func (c *Client) GetRollingWindowUsage(ctx context.Context, hours int) (*ports.UsageWindow, error) {
	// Query for tokens over the rolling window
	tokenQuery := fmt.Sprintf("increase(%s[%dh])", tokenMetric, hours)
	tokens, err := c.queryScalar(ctx, tokenQuery)
	if err != nil {
		return &ports.UsageWindow{
			WindowHours: hours,
			Available:   false,
		}, fmt.Errorf("querying tokens: %w", err)
	}

	// Query for cost over the rolling window
	costQuery := fmt.Sprintf("increase(%s[%dh])", costMetric, hours)
	cost, err := c.queryScalar(ctx, costQuery)
	if err != nil {
		return &ports.UsageWindow{
			TotalTokens: tokens,
			WindowHours: hours,
			Available:   false,
		}, fmt.Errorf("querying cost: %w", err)
	}

	return &ports.UsageWindow{
		TotalTokens: tokens,
		TotalCost:   cost,
		WindowHours: hours,
		Available:   true,
	}, nil
}

// IsAvailable checks if Prometheus is reachable.
func (c *Client) IsAvailable(ctx context.Context) bool {
	// Try a simple query to check availability
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/-/ready", nil)
	if err != nil {
		return false
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// queryScalar executes a Prometheus query and returns the scalar result.
func (c *Client) queryScalar(ctx context.Context, query string) (float64, error) {
	u, err := url.Parse(c.baseURL + "/api/v1/query")
	if err != nil {
		return 0, fmt.Errorf("parsing URL: %w", err)
	}

	q := u.Query()
	q.Set("query", query)
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return 0, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var promResp prometheusResponse
	if err := json.NewDecoder(resp.Body).Decode(&promResp); err != nil {
		return 0, fmt.Errorf("decoding response: %w", err)
	}

	if promResp.Status != "success" {
		return 0, fmt.Errorf("prometheus query failed: %s", promResp.Status)
	}

	if len(promResp.Data.Result) == 0 {
		return 0, nil
	}

	// Extract the scalar value from the result
	if len(promResp.Data.Result[0].Value) < 2 {
		return 0, fmt.Errorf("unexpected result format")
	}

	valueStr, ok := promResp.Data.Result[0].Value[1].(string)
	if !ok {
		return 0, fmt.Errorf("unexpected value type")
	}

	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return 0, fmt.Errorf("parsing value: %w", err)
	}

	return value, nil
}
