package natsstreaming

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"wbl0/recieveMsg/internal/cacheport"
	"wbl0/recieveMsg/internal/dbport"
	"wbl0/recieveMsg/internal/entities"

	"github.com/nats-io/stan.go"
)

const (
	clusterID = "test-cluster"
	clientID  = "event-store"
)

type NatsStreaming struct {
	sc         stan.Conn
	sub        stan.Subscription
	db         *dbport.DbPort
	cachestore *cacheport.CachePort
}

func NewNatsStreaming(db *dbport.DbPort, cachestore *cacheport.CachePort) (*NatsStreaming, error) {

	sc, err := stan.Connect(
		clusterID,
		clientID,
		stan.NatsURL(stan.DefaultNatsURL),
	)

	if err != nil {
		return nil, err
	}
	return &NatsStreaming{
		sc:         sc,
		db:         db,
		cachestore: cachestore,
	}, nil
}

func (ns *NatsStreaming) ListenToNats(ctx context.Context, orderch chan entities.Order) error {
	var order entities.Order

	sub, err := ns.sc.Subscribe("wbmodel",
		func(m *stan.Msg) {
			err := json.Unmarshal(m.Data, &order)
			if err != nil {
				log.Printf("Json marshal error: %s", err)
			}
			orderch <- order
		},
		stan.StartWithLastReceived())
	if err != nil {
		return err
	}

	ns.sub = sub

	return nil
}

func (ns *NatsStreaming) WriteMsgToDbAndCache(ctx context.Context, order entities.Order) error {
	//Writing order to DB
	err := ns.db.WriteOrderData(ctx, order)
	if err != nil {
		return err
	}

	//Writing to cache
	key := fmt.Sprintf("order:%s", order.OrderId)
	err = ns.cachestore.CacheSet(ctx, key, order)
	if err != nil {
		return err
	}
	return nil
}

func (ns *NatsStreaming) WritingMsgRoutine(ctx context.Context, orderch chan entities.Order) {
	var order entities.Order
	for i := range orderch {
		order = i
		if order.OrderId != "" {
			err := ns.WriteMsgToDbAndCache(ctx, order)
			if err != nil {
				log.Fatal("Error writing to db or redis: ", err)
			}
		}
	}
}

func (ns *NatsStreaming) SubClose() {
	ns.sub.Close()
}
