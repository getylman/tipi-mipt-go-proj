package usage

import (
	"context"
	"fmt"

	"github.com/cloud-pricer/pricing/internal/repository"
	"github.com/cloud-pricer/shared/types"
)

type Service struct {
	products    repository.ProductStore
	users       repository.UserStore
	consumption repository.ConsumptionStore
}

func NewService(
	products repository.ProductStore,
	users repository.UserStore,
	consumption repository.ConsumptionStore,
) *Service {
	return &Service{products: products, users: users, consumption: consumption}
}

func (s *Service) Process(ctx context.Context, req *types.UsageRequest) (*types.UsageResponse, error) {
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
	total = round2(total)

	if err := s.users.Upsert(ctx, req.UserID); err != nil {
		return nil, fmt.Errorf("upsert user: %w", err)
	}

	if err := s.consumption.SaveBatch(ctx, req.UserID, breakdown); err != nil {
		return nil, fmt.Errorf("save consumption: %w", err)
	}

	return &types.UsageResponse{
		Status:     "ok",
		UserID:     req.UserID,
		TotalPrice: total,
		Breakdown:  breakdown,
	}, nil
}

func round2(v float64) float64 { return float64(int(v*100+0.5)) / 100 }
