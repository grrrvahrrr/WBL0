package dbport

import (
	"context"
	"wbl0/recieveMsg/internal/entities"
)

type DbStore interface {
	WriteOrderData(ctx context.Context, order entities.Order) error
	GetOrderInfo(ctx context.Context, orderId string) (*entities.Order, error)
}

type DbPort struct {
	dbstore DbStore
}

func NewDataStorage(dbstore DbStore) *DbPort {
	return &DbPort{
		dbstore: dbstore,
	}
}

func (dp *DbPort) WriteOrderData(ctx context.Context, order entities.Order) error {
	err := dp.dbstore.WriteOrderData(ctx, order)
	if err != nil {
		return err
	}
	return nil
}

func (dp *DbPort) GetOrderInfo(ctx context.Context, orderId string) (*entities.Order, error) {
	order, err := dp.dbstore.GetOrderInfo(ctx, orderId)
	if err != nil {
		return nil, err
	}
	return order, nil
}
