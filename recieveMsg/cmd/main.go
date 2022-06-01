package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"wbl0/recieveMsg/internal/api"
	"wbl0/recieveMsg/internal/cacheport"
	"wbl0/recieveMsg/internal/cachestore"
	"wbl0/recieveMsg/internal/database"
	"wbl0/recieveMsg/internal/dbport"
	"wbl0/recieveMsg/internal/entities"
	"wbl0/recieveMsg/internal/natsstreaming"
	"wbl0/recieveMsg/internal/server"
)

func main() {
	//Forgot precommit install
	//Creating Context
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)

	//Creating DB Storage
	const dsn = "postgres://wbuser:wb@localhost:5433/wbl0db?sslmode=disable"
	udf, err := database.NewPgStorage(dsn)
	if err != nil {
		log.Println("Error creating database files: ", err)
	}

	db := dbport.NewDataStorage(udf)

	//Creating Cache Storage
	redisStore := cachestore.NewRedis(db)
	cacheStore := cacheport.NewCacheStorage(redisStore)

	//Nats streaming recieving msg
	ns, err := natsstreaming.NewNatsStreaming(db, cacheStore)
	if err != nil {
		log.Println("Error connecting to nats: ", err)
	}

	orderch := make(chan entities.Order)

	err = ns.ListenToNats(ctx, orderch)
	if err != nil {
		log.Println("Error listening to nats: ", err)
	}

	go ns.WritingMsgRoutine(ctx, orderch)

	//Front
	h := api.NewHandlers(cacheStore, db)
	router := api.NewApiChiRouter(h)
	srv := server.NewServer(":3333", router)

	//Start Server
	srv.Start()

	//Hello
	fmt.Println("Recieving messages!")

	//Shutdown
	<-ctx.Done()
	ns.SubClose()
	srv.Stop()
	cancel()
	fmt.Println("Server shutdown.")

}
