// Package mocks содержит in-memory реализации клиентских интерфейсов.
// Используются во всех юнит-тестах Ingestion Service.
package mocks

import (
	"context"
	"fmt"

	"github.com/cloud-pricer/shared/types"
)

// PricingClient — мок HTTP-клиента к Pricing Engine.
type PricingClient struct {
	UsageResp       *types.UsageResponse
	UsageErr        error
	EstimateResp    *types.EstimateResponse
	EstimateErr     error
	LastUsageReq    *types.UsageRequest
	LastEstimateReq *types.EstimateRequest
}

func (m *PricingClient) Usage(_ context.Context, req *types.UsageRequest) (*types.UsageResponse, error) {
	m.LastUsageReq = req
	return m.UsageResp, m.UsageErr
}

func (m *PricingClient) Estimate(_ context.Context, req *types.EstimateRequest) (*types.EstimateResponse, error) {
	m.LastEstimateReq = req
	return m.EstimateResp, m.EstimateErr
}

// InvalidStore — мок хранилища невалидных метрик.
type InvalidStore struct {
	SaveCount  int
	LastReason string
	Err        error
}

func (m *InvalidStore) Save(_ interface{}, reason string) error {
	m.SaveCount++
	m.LastReason = reason
	return m.Err
}

// FailingInvalidStore — мок который всегда падает при Save.
// Используется для проверки что ошибка сохранения не блокирует ответ клиенту.
type FailingInvalidStore struct{}

func (f *FailingInvalidStore) Save(_ interface{}, _ string) error {
	return fmt.Errorf("db write failed")
}
