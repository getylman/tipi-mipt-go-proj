package usage_test

import (
	"context"
	"errors"
	"testing"

	"github.com/cloud-pricer/pricing/internal/mocks"
	"github.com/cloud-pricer/pricing/internal/usage"
	"github.com/cloud-pricer/shared/types"
)

func newSvc(p *mocks.ProductStore, u *mocks.UserStore, c *mocks.ConsumptionStore) *usage.Service {
	return usage.NewService(p, u, c)
}

func TestService_Process_Success(t *testing.T) {
	productsMock := &mocks.ProductStore{
		Products: map[string]types.Product{
			"vcpu":   {ID: "vcpu",   PricePerUnit: 2.50},
			"ram_gb": {ID: "ram_gb", PricePerUnit: 0.80},
		},
	}
	usersMock       := &mocks.UserStore{}
	consumptionMock := &mocks.ConsumptionStore{}

	resp, err := newSvc(productsMock, usersMock, consumptionMock).Process(
		context.Background(),
		&types.UsageRequest{
			UserID: "550e8400-e29b-41d4-a716-446655440000",
			Items: []types.UsageItem{
				{ProductID: "vcpu",   Quantity: 4},  // 4 × 2.50 = 10.00
				{ProductID: "ram_gb", Quantity: 16}, // 16 × 0.80 = 12.80
			},
		},
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.TotalPrice != 22.80 {
		t.Errorf("total = %.2f, want 22.80", resp.TotalPrice)
	}
	if resp.Status != "ok" {
		t.Errorf("status = %q, want ok", resp.Status)
	}
	if len(resp.Breakdown) != 2 {
		t.Errorf("breakdown len = %d, want 2", len(resp.Breakdown))
	}
	if usersMock.UpsertedID != "550e8400-e29b-41d4-a716-446655440000" {
		t.Errorf("upsert user_id = %q", usersMock.UpsertedID)
	}
	if !consumptionMock.Called {
		t.Error("SaveBatch was not called")
	}
	if len(consumptionMock.SavedItems) != 2 {
		t.Errorf("saved items = %d, want 2", len(consumptionMock.SavedItems))
	}
}

func TestService_Process_ProductNotFound(t *testing.T) {
	consumptionMock := &mocks.ConsumptionStore{}

	_, err := newSvc(
		&mocks.ProductStore{Products: map[string]types.Product{}},
		&mocks.UserStore{},
		consumptionMock,
	).Process(context.Background(), &types.UsageRequest{
		UserID: "550e8400-e29b-41d4-a716-446655440000",
		Items:  []types.UsageItem{{ProductID: "unknown-x99", Quantity: 1}},
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if consumptionMock.Called {
		t.Error("SaveBatch should not be called when product not found")
	}
}

func TestService_Process_DBError_OnGetProducts(t *testing.T) {
	dbErr := errors.New("connection refused")

	_, err := newSvc(
		&mocks.ProductStore{Err: dbErr},
		&mocks.UserStore{},
		&mocks.ConsumptionStore{},
	).Process(context.Background(), &types.UsageRequest{
		UserID: "550e8400-e29b-41d4-a716-446655440000",
		Items:  []types.UsageItem{{ProductID: "vcpu", Quantity: 4}},
	})

	if !errors.Is(err, dbErr) {
		t.Errorf("expected dbErr in chain, got: %v", err)
	}
}

func TestService_Process_DBError_OnUpsertUser(t *testing.T) {
	dbErr        := errors.New("user insert failed")
	consumptionMock := &mocks.ConsumptionStore{}

	_, err := newSvc(
		&mocks.ProductStore{Products: map[string]types.Product{
			"vcpu": {ID: "vcpu", PricePerUnit: 2.50},
		}},
		&mocks.UserStore{Err: dbErr},
		consumptionMock,
	).Process(context.Background(), &types.UsageRequest{
		UserID: "550e8400-e29b-41d4-a716-446655440000",
		Items:  []types.UsageItem{{ProductID: "vcpu", Quantity: 4}},
	})

	if !errors.Is(err, dbErr) {
		t.Errorf("expected dbErr in chain, got: %v", err)
	}
	if consumptionMock.Called {
		t.Error("SaveBatch should not be called when user upsert fails")
	}
}

func TestService_Process_DBError_OnSaveBatch(t *testing.T) {
	dbErr := errors.New("insert failed")

	_, err := newSvc(
		&mocks.ProductStore{Products: map[string]types.Product{
			"vcpu": {ID: "vcpu", PricePerUnit: 2.50},
		}},
		&mocks.UserStore{},
		&mocks.ConsumptionStore{Err: dbErr},
	).Process(context.Background(), &types.UsageRequest{
		UserID: "550e8400-e29b-41d4-a716-446655440000",
		Items:  []types.UsageItem{{ProductID: "vcpu", Quantity: 4}},
	})

	if !errors.Is(err, dbErr) {
		t.Errorf("expected dbErr in chain, got: %v", err)
	}
}

func TestService_Process_CalculationRounding(t *testing.T) {
	resp, err := newSvc(
		&mocks.ProductStore{Products: map[string]types.Product{
			"disk_gb": {ID: "disk_gb", PricePerUnit: 0.15},
		}},
		&mocks.UserStore{},
		&mocks.ConsumptionStore{},
	).Process(context.Background(), &types.UsageRequest{
		UserID: "550e8400-e29b-41d4-a716-446655440000",
		Items:  []types.UsageItem{{ProductID: "disk_gb", Quantity: 100}},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.TotalPrice != 15.00 { // 100 × 0.15
		t.Errorf("total = %.4f, want 15.00", resp.TotalPrice)
	}
}
