package cachestore

import (
	"context"
	"encoding/json"
	"time"
	"wbl0/recieveMsg/internal/dbport"
	"wbl0/recieveMsg/internal/entities"

	"github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"
)

type CacheStore struct {
	cache *cache.Cache
	db    *dbport.DbPort
}

func NewCache(db *dbport.DbPort) *CacheStore {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	rCache := cache.New(&cache.Options{
		Redis:      rdb,
		LocalCache: cache.NewTinyLFU(1000, time.Minute),
	})

	return &CacheStore{
		cache: rCache,
		db:    db,
	}
}

func (cs *CacheStore) CacheSet(ctx context.Context, key string, order entities.Order) error {
	p, err := json.Marshal(order)
	if err != nil {
		return err
	}
	err = cs.cache.Set(&cache.Item{
		Ctx:   ctx,
		Key:   key,
		Value: p,
		TTL:   time.Hour,
	})
	if err != nil {
		return err
	}
	return nil
}

func (cs *CacheStore) CacheGet(ctx context.Context, key string) (*entities.Order, error) {
	var value []byte
	var order entities.Order
	err := cs.cache.Get(ctx, key, &value)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(value, &order)
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (cs *CacheStore) CacheRestore(ctx context.Context, key string, orderId string) error {
	order, err := cs.db.GetOrderInfo(ctx, orderId)
	if err != nil {
		return err
	}

	err = cs.CacheSet(ctx, key, *order)
	if err != nil {
		return err
	}
	return nil
}
