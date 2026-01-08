package app

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"claude-watcher/internal/database"
	"claude-watcher/internal/server"
)

func Run(cfg *Config) error {
	db, err := database.NewTurso(cfg.TursoDatabaseURL, cfg.TursoAuthToken)
	if err != nil {
		return err
	}
	defer db.Close()

	httpSrv := server.NewHTTPServer(cfg.Addr, db)
	go func() {
		log.Printf("http server listening on %s", cfg.Addr)
		if err := httpSrv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	return httpSrv.Shutdown(ctx)
}
