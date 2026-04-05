package validator

import (
	"fmt"

	"github.com/cloud-pricer/shared/types"
)

type Validator struct {
	MaxItems int
}

func New(maxItems int) *Validator {
	return &Validator{MaxItems: maxItems}
}

// ValidateUsage проверяет запрос на ручку /v1/usage.
func (v *Validator) ValidateUsage(req *types.UsageRequest) error {
	if req.UserID == "" {
		return fmt.Errorf("user_id is required")
	}
	return v.validateItems(req.Items)
}

// ValidateEstimate проверяет запрос на ручку /v1/estimate.
func (v *Validator) ValidateEstimate(req *types.EstimateRequest) error {
	return v.validateItems(req.Items)
}

func (v *Validator) validateItems(items []types.UsageItem) error {
	if len(items) == 0 {
		return fmt.Errorf("items list is empty")
	}
	if len(items) > v.MaxItems {
		return fmt.Errorf("too many items: %d, max %d", len(items), v.MaxItems)
	}
	for i, item := range items {
		if item.ProductID == "" {
			return fmt.Errorf("items[%d]: product_id is required", i)
		}
		if item.Quantity <= 0 {
			return fmt.Errorf("items[%d]: quantity must be > 0 (got %v)", i, item.Quantity)
		}
	}
	return nil
}
