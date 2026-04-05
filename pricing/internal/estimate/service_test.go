package estimate_test

import (
	"context"
	"errors"
	"testing"

	"github.com/cloud-pricer/pricing/internal/estimate"
	"github.com/cloud-pricer/pricing/internal/mocks"
	"github.com/cloud-pricer/shared/types"
)

func TestEstimate_Success(t *testing.T) {
	resp, err := estimate.NewService(&mocks.ProductStore{
		Products: map[string]types.Product{
			"vcpu":    {ID: "vcpu",    PricePerUnit: 2.50},
			"ram_gb":  {ID: "ram_gb",  PricePerUnit: 0.80},
			"disk_gb": {ID: "disk_gb", PricePerUnit: 0.15},
		},
	}).Calculate(context.Background(), &types.EstimateRequest{
		Items: []types.UsageItem{
			{ProductID: "vcpu",    Quantity: 4},   // 10.00
			{ProductID: "ram_gb",  Quantity: 16},  // 12.80
			{ProductID: "disk_gb", Quantity: 100}, // 15.00
		},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.TotalPrice != 37.80 {
		t.Errorf("total = %.2f, want 37.80", resp.TotalPrice)
	}
	if len(resp.Breakdown) != 3 {
		t.Errorf("breakdown len = %d, want 3", len(resp.Breakdown))
	}
	for _, item := range resp.Breakdown {
		switch item.ProductID {
		case "vcpu":
			if item.TotalPrice != 10.00 {
				t.Errorf("vcpu total = %.2f, want 10.00", item.TotalPrice)
			}
		case "ram_gb":
			if item.TotalPrice != 12.80 {
				t.Errorf("ram_gb total = %.2f, want 12.80", item.TotalPrice)
			}
		case "disk_gb":
			if item.TotalPrice != 15.00 {
				t.Errorf("disk_gb total = %.2f, want 15.00", item.TotalPrice)
			}
		}
	}
}

func TestEstimate_ProductNotFound(t *testing.T) {
	_, err := estimate.NewService(&mocks.ProductStore{
		Products: map[string]types.Product{},
	}).Calculate(context.Background(), &types.EstimateRequest{
		Items: []types.UsageItem{{ProductID: "gpu_4090", Quantity: 1}},
	})

	if err == nil {
		t.Fatal("expected error for unknown product, got nil")
	}
}

func TestEstimate_DBError(t *testing.T) {
	dbErr := errors.New("timeout")

	_, err := estimate.NewService(&mocks.ProductStore{Err: dbErr}).
		Calculate(context.Background(), &types.EstimateRequest{
			Items: []types.UsageItem{{ProductID: "vcpu", Quantity: 1}},
		})

	if !errors.Is(err, dbErr) {
		t.Errorf("expected dbErr in chain, got: %v", err)
	}
}

func TestEstimate_NoUserIDField(t *testing.T) {
	// estimate не требует user_id — убеждаемся что работает без него
	resp, err := estimate.NewService(&mocks.ProductStore{
		Products: map[string]types.Product{
			"vcpu": {ID: "vcpu", PricePerUnit: 2.50},
		},
	}).Calculate(context.Background(), &types.EstimateRequest{
		Items: []types.UsageItem{{ProductID: "vcpu", Quantity: 2}},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.TotalPrice != 5.00 { // 2 × 2.50
		t.Errorf("total = %.2f, want 5.00", resp.TotalPrice)
	}
}
