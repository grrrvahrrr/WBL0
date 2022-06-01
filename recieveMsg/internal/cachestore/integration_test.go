//go:build integration_tests

package cachestore

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"testing"
	"time"
	"wbl0/recieveMsg/internal/entities"

	"github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"
	"github.com/ory/dockertest/v3"
)

var client *redis.Client

func TestMain(m *testing.M) {
	os.Exit(testWrapper(m))
}

func testWrapper(m *testing.M) int {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	resource, err := pool.Run("redis", "latest", nil)
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}
	defer pool.Purge(resource)

	if err := pool.Retry(func() error {
		client = redis.NewClient(&redis.Options{
			Addr:     net.JoinHostPort("localhost", resource.GetPort("6379/tcp")),
			Password: "", // no password set
			DB:       0,  // use default DB
		})

		ping := client.Ping(context.Background())
		return ping.Err()
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	return m.Run()
}

func TestCache(t *testing.T) {
	testdelivery := entities.Delivery{
		Name:    "testname",
		Phone:   "testphone",
		Zip:     "testzip",
		City:    "testcity",
		Address: "testaddress",
		Region:  "testregion",
		Email:   "testemail",
	}

	testpayment := entities.Payment{
		Transaction:  "testid",
		RequestID:    "testreqid",
		Currency:     "cur",
		Provider:     "testprv",
		Amount:       1,
		PaymentDt:    2,
		Bank:         "testbank",
		DeliveryCost: 3,
		GoodsTotal:   4,
		CustomFee:    5,
	}

	testcartitem := entities.CartItem{
		ChrtId:      0,
		TrackNumber: "testtracknum",
		Price:       1,
		Rid:         "testrid",
		Name:        "testname",
		Sale:        2,
		Size:        "testsize",
		TotalPrice:  3,
		NmId:        4,
		Brand:       "testbrand",
		Status:      5,
	}

	var testitems []entities.CartItem
	testitems = append(testitems, testcartitem)

	testorder := entities.Order{
		OrderId:           "testid",
		TrackNumber:       "testnumber",
		Entry:             "testentry",
		Delivery:          testdelivery,
		Payment:           testpayment,
		Items:             testitems,
		Locale:            "testlocale",
		InternalSignature: "testinternalsig",
		CustomerId:        "testcustomerid",
		DeliveryService:   "testdelser",
		ShardKey:          "testshardkey",
		SmID:              0,
		DateCreated:       "testdate",
		OofShard:          "testoof",
	}

	rCache := cache.New(&cache.Options{
		Redis:      client,
		LocalCache: cache.NewTinyLFU(1000, time.Minute),
	})

	rs := RedisStore{
		cache: rCache,
	}

	key := fmt.Sprintf("order:%s", testorder.OrderId)
	err := rs.CacheSet(context.Background(), key, testorder)
	if err != nil {
		t.Fatalf("failed to write to Redis: %v", err)
	}

	newTestOrder, err := rs.CacheGet(context.Background(), key)
	if err != nil {
		t.Fatalf("failed to get data from Redis: %v", err)
	}

	if newTestOrder.TrackNumber != testorder.TrackNumber {
		t.Errorf("Failed to Read Full Url, got %s", newTestOrder.TrackNumber)
	}
}
