package handler

import (
	"context"

	"github.com/cloud-pricer/shared/types"
)

type ProductLister interface {
	ListAll(ctx context.Context) ([]types.Product, error)
}
