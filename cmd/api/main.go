package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/iosh/go-greenlight/internal/data"
	"github.com/iosh/go-greenlight/internal/jsonlog"
	"github.com/jackc/pgx/v5/pgxpool"
)

const version = "1.0.0"

func openDB(cfg config) (*pgxpool.Pool, error) {
	dbpool, err := pgxpool.New(context.Background(), cfg.db.dsn)

	if err != nil {
		return nil, err
	}

	return dbpool, nil
}

type config struct {
	port int
	env  string
	db   struct {
		dsn string
	}
}

type application struct {
	config config
	logger *jsonlog.Logger
	models data.Models
}

func main() {
	var cfg config

	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.StringVar(&cfg.db.dsn, "db-dsn", "postgres://root:root@localhost/greenlight?sslmode=disable", "PostgreSQL DSN")

	flag.Parse()

	db, err := openDB(cfg)

	defer db.Close()

	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)
	if err != nil {
		logger.PrintFatal(err, nil)
	}
	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
	}
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routers(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	logger.Printf("starting %s server on %s", cfg.env, srv.Addr)
	serverErr := srv.ListenAndServe()

	logger.Fatal(serverErr)
}
