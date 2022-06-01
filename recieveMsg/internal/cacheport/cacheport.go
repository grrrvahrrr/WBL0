package cacheport

import (
	"context"
	"wbl0/recieveMsg/internal/entities"
)

type CacheStore interface {
	CacheSet(ctx context.Context, key string, order entities.Order) error
	CacheGet(ctx context.Context, key string) (*entities.Order, error)
	CacheRestore(ctx context.Context, key string, orderId string) error
}

type CachePort struct {
	cs CacheStore
}

func NewCacheStorage(cs CacheStore) *CachePort {
	return &CachePort{
		cs: cs,
	}
}

func (cp *CachePort) CacheSet(ctx context.Context, key string, order entities.Order) error {
	err := cp.cs.CacheSet(ctx, key, order)
	if err != nil {
		return err
	}
	return nil
}

func (cp *CachePort) CacheGet(ctx context.Context, key string) (*entities.Order, error) {
	order, err := cp.cs.CacheGet(ctx, key)
	if err != nil {
		return nil, err
	}
	return order, nil
}

func (cp *CachePort) CacheRestore(ctx context.Context, key string, orderId string) error {
	err := cp.cs.CacheRestore(ctx, key, orderId)
	if err != nil {
		return err
	}
	return nil
}
