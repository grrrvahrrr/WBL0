package api

import (
	"context"
	"fmt"
	"wbl0/recieveMsg/internal/cacheport"
	"wbl0/recieveMsg/internal/dbport"
	"wbl0/recieveMsg/internal/entities"
)

//DTO here as well

type Handlers struct {
	cache   *cacheport.CachePort
	db      *dbport.DbPort
	initRun bool
}

func NewHandlers(cache *cacheport.CachePort, db *dbport.DbPort) *Handlers {
	return &Handlers{
		cache:   cache,
		db:      db,
		initRun: true,
	}
}

func (h *Handlers) GetOrderByIdHandler(ctx context.Context, orderId string) (*entities.Order, error) {
	//Restore cache from db
	key := fmt.Sprintf("order:%s", orderId)
	if h.initRun {
		err := h.cache.CacheRestore(ctx, key, orderId)
		if err != nil {
			return nil, err
		}
		h.initRun = false
	}

	//Read order from cache
	orderRedis, err := h.cache.CacheGet(ctx, key)
	if err != nil {
		return nil, err
	}

	return orderRedis, nil
}
