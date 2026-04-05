// Package mocks содержит in-memory реализации интерфейсов репозиториев.
// Используются во всех юнит-тестах Pricing Engine.
package mocks

import (
	"context"

	"github.com/cloud-pricer/shared/types"
)

// ProductStore — мок хранилища продуктов.
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
	var list []types.Product
	for _, p := range m.Products {
		list = append(list, p)
	}
	return list, nil
}

// UserStore — мок хранилища пользователей.
type UserStore struct {
	UpsertedID string
	Err        error
}

func (m *UserStore) Upsert(_ context.Context, userID string) error {
	m.UpsertedID = userID
	return m.Err
}

// ConsumptionStore — мок хранилища потребления.
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

func (m *ConsumptionStore) GetByUser(_ context.Context, _ string) ([]types.UsageLineItem, error) {
	return m.SavedItems, m.Err
}

func (m *ConsumptionStore) SumByUser(_ context.Context, _ string) (float64, error) {
	var total float64
	for _, item := range m.SavedItems {
		total += item.TotalPrice
	}
	return total, m.Err
}
