package main

import (
	"assignment/api"
	"assignment/domain"
	"assignment/storage"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"

	"github.com/jackc/pgx/v5/pgxpool"
)

func serveHttp(exitCh <-chan os.Signal, log *slog.Logger, c api.Controller) {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/create_segment", c.CreateSegment)
	mux.HandleFunc("/api/delete_segment", c.DeleteSegment)
	mux.HandleFunc("/api/change_user_segments", c.ChangeUserSegments)
	mux.HandleFunc("/api/get_user_segments", c.GetUserSegments)

	srv := &http.Server{Addr: "0.0.0.0:80", Handler: mux}

	go func() {
		<-exitCh
		log.Info("shutting down")
		srv.Shutdown(context.TODO())
	}()

	if err := srv.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			fmt.Fprintf(os.Stderr, "failed to listen and serve: %v\n", err)
			os.Exit(1)
		}
	}
}

func main() {
	dbpool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
		os.Exit(1)
	}
	defer dbpool.Close()

	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	exitCh := make(chan os.Signal, 1)
	signal.Notify(exitCh, os.Interrupt)

	sqlStore := storage.NewSqlStorage(dbpool)
	if err := sqlStore.InitDb(context.Background()); err != nil {
		log.Error("failed to init database", slog.String("error", err.Error()))
		os.Exit(1)
	}

	c := api.Controller{
		SegmentService: domain.NewSegmentService(sqlStore),
		Log:            log,
	}

	serveHttp(exitCh, log, c)
}
