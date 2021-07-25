package main

import (
	"context"
	"os"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/merisho/binaryx-test/activerecord"
	"github.com/merisho/binaryx-test/api"
	"github.com/merisho/binaryx-test/service"
	log "github.com/sirupsen/logrus"
)

func main() {
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		dbURL = "postgres://admin:12345@localhost:5432/test?sslmode=disable"
	}

	db, err := pgxpool.Connect(context.Background(), dbURL)
	if err != nil {
		log.WithError(err).Fatal("could not connect to DB")
	}

	activeRecordFactory := activerecord.New(db)
	serviceWallets, err := service.NewWallets(activeRecordFactory)
	if err != nil {
		log.WithError(err).Fatal("could not connect to DB")
	}

	conf := api.Config{
		JWTSecret: "test",
		APIMode:   api.TestMode,
		Port: 8080,
		TokenTTLSeconds: 3600,
	}
	srv, err := api.NewServer(conf, activeRecordFactory, serviceWallets)
	if err != nil {
		log.WithError(err).Fatal()
	}

	go log.Fatal(srv.Listen())
	log.Printf("Listening on %d", conf.Port)
	select {}
}
