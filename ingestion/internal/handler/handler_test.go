package handler_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cloud-pricer/ingestion/internal/handler"
	"github.com/cloud-pricer/ingestion/internal/mocks"
	"github.com/cloud-pricer/ingestion/internal/validator"
	"github.com/cloud-pricer/shared/types"
)

func newHandler(pricing *mocks.PricingClient, inv *mocks.InvalidStore) *handler.Handler {
	return handler.New(nil, validator.New(50), pricing, inv)
}

// ─── Usage тесты ─────────────────────────────────────────────────

func TestHandler_Usage_Success(t *testing.T) {
	pricing := &mocks.PricingClient{
		UsageResp: &types.UsageResponse{
			Status:     "ok",
			UserID:     "550e8400-e29b-41d4-a716-446655440000",
			TotalPrice: 37.80,
			Breakdown:  []types.UsageLineItem{{ProductID: "vcpu", Quantity: 4, UnitPrice: 2.50, TotalPrice: 10}},
		},
	}
	inv := &mocks.InvalidStore{}
	h   := newHandler(pricing, inv)

	body := `{"user_id":"550e8400-e29b-41d4-a716-446655440000","items":[{"product_id":"vcpu","quantity":4}]}`
	req  := httptest.NewRequest(http.MethodPost, "/v1/usage", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Usage(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	if inv.SaveCount != 0 {
		t.Errorf("invalid Save called %d times, want 0", inv.SaveCount)
	}
	if pricing.LastUsageReq == nil {
		t.Error("pricing client was not called")
	}

	var resp types.UsageResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.TotalPrice != 37.80 {
		t.Errorf("total = %.2f, want 37.80", resp.TotalPrice)
	}
}

func TestHandler_Usage_ValidationError_EmptyUserID(t *testing.T) {
	pricing := &mocks.PricingClient{}
	inv     := &mocks.InvalidStore{}
	h       := newHandler(pricing, inv)

	body := `{"user_id":"","items":[{"product_id":"vcpu","quantity":4}]}`
	req  := httptest.NewRequest(http.MethodPost, "/v1/usage", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Usage(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
	if inv.SaveCount != 1 {
		t.Errorf("invalid Save called %d times, want 1", inv.SaveCount)
	}
	if pricing.LastUsageReq != nil {
		t.Error("pricing client should not be called on validation failure")
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	errObj := resp["error"].(map[string]interface{})
	if errObj["code"] != "VALIDATION_ERROR" {
		t.Errorf("error code = %v, want VALIDATION_ERROR", errObj["code"])
	}
}

func TestHandler_Usage_ValidationError_EmptyItems(t *testing.T) {
	inv := &mocks.InvalidStore{}
	h   := newHandler(&mocks.PricingClient{}, inv)

	body := `{"user_id":"550e8400-e29b-41d4-a716-446655440000","items":[]}`
	req  := httptest.NewRequest(http.MethodPost, "/v1/usage", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Usage(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
	if inv.SaveCount != 1 {
		t.Errorf("invalid Save called %d times, want 1", inv.SaveCount)
	}
}

func TestHandler_Usage_PricingEngineDown(t *testing.T) {
	pricing := &mocks.PricingClient{UsageErr: errors.New("connection refused")}
	inv     := &mocks.InvalidStore{}
	h       := newHandler(pricing, inv)

	body := `{"user_id":"550e8400-e29b-41d4-a716-446655440000","items":[{"product_id":"vcpu","quantity":4}]}`
	req  := httptest.NewRequest(http.MethodPost, "/v1/usage", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Usage(w, req)

	if w.Code != http.StatusBadGateway {
		t.Errorf("status = %d, want 502", w.Code)
	}
	// invalid_metrics НЕ пишем — валидация прошла, просто upstream упал
	if inv.SaveCount != 0 {
		t.Errorf("invalid Save should not be called on upstream error, got %d", inv.SaveCount)
	}
}

func TestHandler_Usage_InvalidJSON(t *testing.T) {
	h   := newHandler(&mocks.PricingClient{}, &mocks.InvalidStore{})
	req := httptest.NewRequest(http.MethodPost, "/v1/usage", strings.NewReader(`{broken`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Usage(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestHandler_Usage_SaveInvalidFails_StillReturns400(t *testing.T) {
	// saveInvalid упал — клиент всё равно получает 400, не 500
	h := handler.New(nil, validator.New(50), &mocks.PricingClient{}, &mocks.FailingInvalidStore{})

	body := `{"user_id":"","items":[]}`
	req  := httptest.NewRequest(http.MethodPost, "/v1/usage", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Usage(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestHandler_Usage_ResponseContainsBreakdown(t *testing.T) {
	pricing := &mocks.PricingClient{
		UsageResp: &types.UsageResponse{
			Status:     "ok",
			UserID:     "550e8400-e29b-41d4-a716-446655440000",
			TotalPrice: 22.80,
			Breakdown: []types.UsageLineItem{
				{ProductID: "vcpu",   Quantity: 4,  UnitPrice: 2.50, TotalPrice: 10.00},
				{ProductID: "ram_gb", Quantity: 16, UnitPrice: 0.80, TotalPrice: 12.80},
			},
		},
	}
	h := newHandler(pricing, &mocks.InvalidStore{})

	body := `{"user_id":"550e8400-e29b-41d4-a716-446655440000","items":[{"product_id":"vcpu","quantity":4},{"product_id":"ram_gb","quantity":16}]}`
	req  := httptest.NewRequest(http.MethodPost, "/v1/usage", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Usage(w, req)

	var resp types.UsageResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if len(resp.Breakdown) != 2 {
		t.Errorf("breakdown len = %d, want 2", len(resp.Breakdown))
	}
}

// ─── Estimate тесты ───────────────────────────────────────────────

func TestHandler_Estimate_Success(t *testing.T) {
	pricing := &mocks.PricingClient{
		EstimateResp: &types.EstimateResponse{
			TotalPrice: 37.80,
			Breakdown:  []types.UsageLineItem{{ProductID: "vcpu", Quantity: 4, UnitPrice: 2.50, TotalPrice: 10}},
		},
	}
	h := newHandler(pricing, &mocks.InvalidStore{})

	body := `{"items":[{"product_id":"vcpu","quantity":4}]}`
	req  := httptest.NewRequest(http.MethodPost, "/v1/estimate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Estimate(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestHandler_Estimate_ValidationError(t *testing.T) {
	pricing := &mocks.PricingClient{}
	inv     := &mocks.InvalidStore{}
	h       := newHandler(pricing, inv)

	body := `{"items":[]}`
	req  := httptest.NewRequest(http.MethodPost, "/v1/estimate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Estimate(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
	if inv.SaveCount != 1 {
		t.Errorf("invalid Save called %d times, want 1", inv.SaveCount)
	}
	if pricing.LastEstimateReq != nil {
		t.Error("pricing should not be called on validation failure")
	}
}

func TestHandler_Estimate_PricingEngineDown(t *testing.T) {
	pricing := &mocks.PricingClient{EstimateErr: errors.New("timeout")}
	h       := newHandler(pricing, &mocks.InvalidStore{})

	body := `{"items":[{"product_id":"vcpu","quantity":4}]}`
	req  := httptest.NewRequest(http.MethodPost, "/v1/estimate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Estimate(w, req)

	if w.Code != http.StatusBadGateway {
		t.Errorf("status = %d, want 502", w.Code)
	}
}

func TestHandler_Estimate_InvalidJSON(t *testing.T) {
	h   := newHandler(&mocks.PricingClient{}, &mocks.InvalidStore{})
	req := httptest.NewRequest(http.MethodPost, "/v1/estimate", strings.NewReader(`{bad`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Estimate(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

// ─── Health тест ─────────────────────────────────────────────────

func TestHandler_Health(t *testing.T) {
	h   := newHandler(&mocks.PricingClient{}, &mocks.InvalidStore{})
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w   := httptest.NewRecorder()

	h.Health(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["status"] != "ok" {
		t.Errorf("status = %q, want ok", resp["status"])
	}
}
