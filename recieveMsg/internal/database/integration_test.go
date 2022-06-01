//go:build integration_tests

package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
	"wbl0/recieveMsg/internal/entities"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"

	_ "github.com/jackc/pgx/v4/stdlib" //psql driver
)

var dsn string

type setupResult struct {
	Pool              *dockertest.Pool
	PostgresContainer *dockertest.Resource
}

func TestMain(m *testing.M) {
	os.Exit(testMain(m))
}

func testMain(m *testing.M) int {
	setupResult, err := setup()
	if err != nil {
		log.Panicln("setup err: ", err)
		return -1
	}
	defer teardown(setupResult)
	return m.Run()
}

func setup() (r *setupResult, err error) {
	testFileDir, err := getTestFileDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get the script dir: %w", err)
	}

	pool, err := dockertest.NewPool("")
	if err != nil {
		return nil, fmt.Errorf("failed to create a new docketest pool: %w", err)
	}
	pool.MaxWait = time.Second * 5

	postgresContainer, err := runPostgresContainer(pool, testFileDir)
	if err != nil {
		return nil, fmt.Errorf("failed to run the Postgres container: %w", err)
	}
	defer func() {
		if err != nil {
			if err := pool.Purge(postgresContainer); err != nil {
				log.Println("failed to purge the postgres container: %w", err)
			}
		}
	}()

	return &setupResult{
		Pool:              pool,
		PostgresContainer: postgresContainer,
	}, nil
}

func getTestFileDir() (string, error) {
	_, fileName, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("failed to get the caller info")
	}
	fileDir := filepath.Dir(fileName)
	dir, err := filepath.Abs(fileDir)
	if err != nil {
		return "", fmt.Errorf("failed to get the absolute path to the directory %s: %w", dir, err)
	}
	log.Println(fileDir)
	return fileDir, nil
}

func runPostgresContainer(pool *dockertest.Pool, testFileDir string) (*dockertest.Resource, error) {
	postgresContainer, err := pool.RunWithOptions(
		&dockertest.RunOptions{
			Repository: "postgres",
			Tag:        "14.0",
			Env: []string{
				"POSTGRES_PASSWORD=wb",
			},
		},
		func(config *docker.HostConfig) {
			config.AutoRemove = false
			config.RestartPolicy = docker.RestartPolicy{Name: "no"}
			config.Mounts = []docker.HostMount{
				{
					Target: "/docker-entrypoint-initdb.d",
					Source: filepath.Join(testFileDir, "testinit"),
					Type:   "bind",
				},
			}
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start the postgres docker container: %w", err)
	}

	postgresContainer.Expire(120)
	port := postgresContainer.GetPort("5432/tcp")
	fmt.Println(port)
	dsn = fmt.Sprintf("postgres://wbuser:wb@localhost:%s/wbl0db?sslmode=disable", port)

	// Wait for the DB to start
	if err := pool.Retry(func() error {
		db, err := getDBConnector()
		if err != nil {
			return fmt.Errorf("failed to get a DB connector: %w", err)
		}
		return db.Ping()
	}); err != nil {
		pool.Purge(postgresContainer)
		return nil, fmt.Errorf("failed to ping the created DB: %w", err)
	}
	return postgresContainer, nil
}

func getDBConnector() (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func teardown(r *setupResult) {
	if err := r.Pool.Purge(r.PostgresContainer); err != nil {
		log.Printf("failed to purge the Postgres container: %v", err)
	}
}
func TestWriteOrderData(t *testing.T) {
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

	conn, err := getDBConnector()
	if err != nil {
		t.Fatalf("failed to get a connector to the DB: %v", err)
	}

	ctx := context.Background()

	pg := PgStorage{
		db: conn,
	}

	err = pg.WriteOrderData(ctx, testorder)
	if err != nil {
		t.Fatalf("failed to Write Order Data to the DB: %v", err)
	}
}

func TestGetOrderInfo(t *testing.T) {

	conn, err := getDBConnector()
	if err != nil {
		t.Fatalf("failed to get a connector to the DB: %v", err)
	}

	ctx := context.Background()

	pg := PgStorage{
		db: conn,
	}

	newTestOrder, err := pg.GetOrderInfo(ctx, "testid")
	if err != nil {
		t.Fatalf("failed to Read Order Info from the DB: %v", err)
	}

	fmt.Println("result: ", *newTestOrder)

	if newTestOrder.TrackNumber != "testnumber" {
		t.Errorf("Failed to Read Full Url, got %s", newTestOrder.TrackNumber)
	}

}
