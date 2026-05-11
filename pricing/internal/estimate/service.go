package estimate

import (
	"context"
	"fmt"

	"github.com/cloud-pricer/shared/types"
)

type Service struct {
	products ProductStore
}

func NewService(products ProductStore) *Service {
	return &Service{products: products}
}

func (s *Service) Calculate(ctx context.Context, req *types.EstimateRequest) (*types.EstimateResponse, error) {
	ids := make([]string, len(req.Items))
	for i, item := range req.Items {
		ids[i] = item.ProductID
	}

	products, err := s.products.GetByIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("load products: %w", err)
	}

	for _, item := range req.Items {
		if _, ok := products[item.ProductID]; !ok {
			return nil, fmt.Errorf("product %q not found", item.ProductID)
		}
	}

	breakdown := make([]types.UsageLineItem, 0, len(req.Items))
	var total float64
	for _, item := range req.Items {
		p := products[item.ProductID]
		itemTotal := round2(item.Quantity * p.PricePerUnit)
		breakdown = append(breakdown, types.UsageLineItem{
			ProductID:  item.ProductID,
			Quantity:   item.Quantity,
			UnitPrice:  p.PricePerUnit,
			TotalPrice: itemTotal,
		})
		total += itemTotal
	}

	return &types.EstimateResponse{
		TotalPrice: round2(total),
		Breakdown:  breakdown,
	}, nil
}

func round2(v float64) float64 { return float64(int(v*100+0.5)) / 100 }
