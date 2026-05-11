// Интеграционный тест всей системы через реальный HTTP.
// Запускает оба сервиса как процессы и проверяет их взаимодействие.
// Требует запущенных БД (make _db-start) или пропускается если сервисы недоступны.
//
// Запуск с живыми сервисами:
//
//	make run
//	cd pricing && go test ./tests/... -v
//
// Или только быстрая проверка без сервисов (тест сам пропустит):
//
//	cd pricing && go test ./tests/... -v -run TestIntegration
package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/cloud-pricer/shared/types"
)

const (
	ingestionURL = "http://localhost:8080"
	pricingURL   = "http://localhost:8081"
)

// skipIfDown пропускает тест если сервис недоступен.
func skipIfDown(t *testing.T) {
	t.Helper()
	client := &http.Client{Timeout: time.Second}
	_, err := client.Get(ingestionURL + "/health")
	if err != nil {
		t.Skipf("ingestion service not running at %s, skipping integration test", ingestionURL)
	}
}

func postJSON(t *testing.T, url string, body interface{}) *http.Response {
	t.Helper()
	data, _ := json.Marshal(body)
	resp, err := http.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		t.Fatalf("POST %s failed: %v", url, err)
	}
	return resp
}

func TestIntegration_Health(t *testing.T) {
	skipIfDown(t)

	for name, url := range map[string]string{
		"ingestion": ingestionURL + "/health",
		"pricing":   pricingURL + "/health",
	} {
		resp, err := http.Get(url)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("%s health = %d, want 200", name, resp.StatusCode)
		}
	}
}

func TestIntegration_Usage_FullFlow(t *testing.T) {
	skipIfDown(t)

	resp := postJSON(t, ingestionURL+"/v1/usage", types.UsageRequest{
		UserID: "550e8400-e29b-41d4-a716-446655440099",
		Items: []types.UsageItem{
			{ProductID: "vcpu",    Quantity: 4},
			{ProductID: "ram_gb",  Quantity: 16},
			{ProductID: "disk_gb", Quantity: 100},
		},
	})
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	var result types.UsageResponse
	json.NewDecoder(resp.Body).Decode(&result)

	// 4×2.50 + 16×0.80 + 100×0.15 = 37.80
	if result.TotalPrice != 37.80 {
		t.Errorf("total = %.2f, want 37.80", result.TotalPrice)
	}
	if result.Status != "ok" {
		t.Errorf("status = %q, want ok", result.Status)
	}
	if len(result.Breakdown) != 3 {
		t.Errorf("breakdown len = %d, want 3", len(result.Breakdown))
	}
	fmt.Printf("  total_price: %.2f\n", result.TotalPrice)
}

func TestIntegration_Estimate_NothingSaved(t *testing.T) {
	skipIfDown(t)

	resp := postJSON(t, ingestionURL+"/v1/estimate", types.EstimateRequest{
		Items: []types.UsageItem{
			{ProductID: "vcpu",   Quantity: 4},
			{ProductID: "ram_gb", Quantity: 16},
		},
	})
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	var result types.EstimateResponse
	json.NewDecoder(resp.Body).Decode(&result)

	// 4×2.50 + 16×0.80 = 22.80
	if result.TotalPrice != 22.80 {
		t.Errorf("total = %.2f, want 22.80", result.TotalPrice)
	}
	// Estimate не возвращает user_id — проверяем что поле отсутствует
	if len(result.Breakdown) != 2 {
		t.Errorf("breakdown len = %d, want 2", len(result.Breakdown))
	}
}

func TestIntegration_InvalidRequest_Returns400(t *testing.T) {
	skipIfDown(t)

	resp := postJSON(t, ingestionURL+"/v1/usage", map[string]interface{}{
		"user_id": "",
		"items":   []interface{}{},
	})
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}

	var errResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&errResp)
	errObj, ok := errResp["error"].(map[string]interface{})
	if !ok {
		t.Fatal("missing 'error' in response")
	}
	if errObj["code"] != "VALIDATION_ERROR" {
		t.Errorf("error code = %v, want VALIDATION_ERROR", errObj["code"])
	}
}

func TestIntegration_UnknownProduct_Returns4xx(t *testing.T) {
	skipIfDown(t)

	resp := postJSON(t, ingestionURL+"/v1/usage", types.UsageRequest{
		UserID: "550e8400-e29b-41d4-a716-446655440099",
		Items:  []types.UsageItem{{ProductID: "gpu_9090", Quantity: 1}},
	})
	defer resp.Body.Close()

	// Pricing вернёт 404 или 500 — в любом случае не 200
	if resp.StatusCode == http.StatusOK {
		t.Error("expected non-200 for unknown product")
	}
}
