package estimate

import (
	"context"

	"github.com/cloud-pricer/shared/types"
)

type ProductStore interface {
	GetByIDs(ctx context.Context, ids []string) (map[string]types.Product, error)
}
