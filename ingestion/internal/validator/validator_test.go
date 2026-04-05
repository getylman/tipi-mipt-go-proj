package validator_test

import (
	"testing"

	"github.com/cloud-pricer/ingestion/internal/validator"
	"github.com/cloud-pricer/shared/types"
)

func newValidator() *validator.Validator {
	return validator.New(50)
}


func TestValidateUsage_Valid(t *testing.T) {
	err := newValidator().ValidateUsage(&types.UsageRequest{
		UserID: "550e8400-e29b-41d4-a716-446655440000",
		Items:  []types.UsageItem{{ProductID: "vcpu", Quantity: 4}},
	})
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestValidateUsage_EmptyUserID(t *testing.T) {
	err := newValidator().ValidateUsage(&types.UsageRequest{
		UserID: "",
		Items:  []types.UsageItem{{ProductID: "vcpu", Quantity: 4}},
	})
	if err == nil {
		t.Error("expected error for empty user_id")
	}
}

func TestValidateUsage_EmptyItems(t *testing.T) {
	err := newValidator().ValidateUsage(&types.UsageRequest{
		UserID: "550e8400-e29b-41d4-a716-446655440000",
		Items:  []types.UsageItem{},
	})
	if err == nil {
		t.Error("expected error for empty items")
	}
}

func TestValidateUsage_EmptyProductID(t *testing.T) {
	err := newValidator().ValidateUsage(&types.UsageRequest{
		UserID: "550e8400-e29b-41d4-a716-446655440000",
		Items:  []types.UsageItem{{ProductID: "", Quantity: 4}},
	})
	if err == nil {
		t.Error("expected error for empty product_id")
	}
}

func TestValidateUsage_ZeroQuantity(t *testing.T) {
	err := newValidator().ValidateUsage(&types.UsageRequest{
		UserID: "550e8400-e29b-41d4-a716-446655440000",
		Items:  []types.UsageItem{{ProductID: "vcpu", Quantity: 0}},
	})
	if err == nil {
		t.Error("expected error for zero quantity")
	}
}

func TestValidateUsage_NegativeQuantity(t *testing.T) {
	err := newValidator().ValidateUsage(&types.UsageRequest{
		UserID: "550e8400-e29b-41d4-a716-446655440000",
		Items:  []types.UsageItem{{ProductID: "vcpu", Quantity: -1}},
	})
	if err == nil {
		t.Error("expected error for negative quantity")
	}
}

func TestValidateUsage_TooManyItems(t *testing.T) {
	v     := validator.New(2) // лимит 2 элемента
	items := []types.UsageItem{
		{ProductID: "vcpu",    Quantity: 4},
		{ProductID: "ram_gb",  Quantity: 16},
		{ProductID: "disk_gb", Quantity: 100},
	}
	err := v.ValidateUsage(&types.UsageRequest{
		UserID: "550e8400-e29b-41d4-a716-446655440000",
		Items:  items,
	})
	if err == nil {
		t.Error("expected error for too many items")
	}
}


func TestValidateEstimate_Valid(t *testing.T) {
	err := newValidator().ValidateEstimate(&types.EstimateRequest{
		Items: []types.UsageItem{{ProductID: "vcpu", Quantity: 4}},
	})
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestValidateEstimate_NoUserIDRequired(t *testing.T) {
	// estimate не требует user_id
	err := newValidator().ValidateEstimate(&types.EstimateRequest{
		Items: []types.UsageItem{{ProductID: "vcpu", Quantity: 4}},
	})
	if err != nil {
		t.Errorf("estimate should not require user_id, got: %v", err)
	}
}

func TestValidateEstimate_EmptyItems(t *testing.T) {
	err := newValidator().ValidateEstimate(&types.EstimateRequest{
		Items: []types.UsageItem{},
	})
	if err == nil {
		t.Error("expected error for empty items")
	}
}
