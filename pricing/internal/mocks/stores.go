package mocks

import (
	"context"

	"github.com/cloud-pricer/shared/types"
)

type ProductStore struct {
	Products   map[string]types.Product
	Err        error
	CalledWith []string
}

func (m *ProductStore) GetByIDs(_ context.Context, ids []string) (map[string]types.Product, error) {
	m.CalledWith = ids
	return m.Products, m.Err
}

func (m *ProductStore) ListAll(_ context.Context) ([]types.Product, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	list := make([]types.Product, 0, len(m.Products))
	for _, p := range m.Products {
		list = append(list, p)
	}
	return list, nil
}

type UserStore struct {
	UpsertedID string
	Err        error
}

func (m *UserStore) Upsert(_ context.Context, userID string) error {
	m.UpsertedID = userID
	return m.Err
}

type ConsumptionStore struct {
	SavedUserID string
	SavedItems  []types.UsageLineItem
	Err         error
	Called      bool
}

func (m *ConsumptionStore) SaveBatch(_ context.Context, userID string, items []types.UsageLineItem) error {
	m.Called = true
	m.SavedUserID = userID
	m.SavedItems = items
	return m.Err
}
