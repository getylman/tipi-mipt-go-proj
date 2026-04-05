package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/cloud-pricer/shared/types"
)

type PricingClient struct {
	baseURL    string
	httpClient *http.Client
}

func New(baseURL string, timeout time.Duration) *PricingClient {
	return &PricingClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: timeout},
	}
}

func (c *PricingClient) Usage(ctx context.Context, req *types.UsageRequest) (*types.UsageResponse, error) {
	return doPost[types.UsageResponse](ctx, c, "/v1/usage", req)
}

func (c *PricingClient) Estimate(ctx context.Context, req *types.EstimateRequest) (*types.EstimateResponse, error) {
	return doPost[types.EstimateResponse](ctx, c, "/v1/estimate", req)
}

func doPost[T any](ctx context.Context, c *PricingClient, path string, body interface{}) (*T, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("pricing engine unavailable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("pricing engine returned status %d", resp.StatusCode)
	}

	var result T
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &result, nil
}
