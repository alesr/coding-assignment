package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	"go.uber.org/zap"

	"github.com/qredo-external/go-alessandro-resta/app"
	"github.com/qredo-external/go-alessandro-resta/internal/service"

	"github.com/go-chi/chi"
	envars "github.com/netflix/go-env"
)

const gracefullyShutdownTimeout = 5 * time.Second

type config struct {
	Port   string `env:"PORT,default=8080"`
	JWTKey string `env:"JWT,default=secret"`
}

func newConfig() *config {
	var cfg config
	if _, err := envars.UnmarshalFromEnviron(&cfg); err != nil {
		log.Fatal(err)
	}
	return &cfg
}

func main() {
	// Decide on dev or prod log based on env var.
	// But keeping it simple here.
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalln("failed to create logger", err)
	}
	defer logger.Sync()

	cfg := newConfig()

	svc := service.NewDefaultService(logger, []byte(cfg.JWTKey))
	rest := app.NewRESTApp(logger, cfg.Port, chi.NewRouter(), svc)

	go func() {
		if err := rest.Start(); err != nil {
			logger.Fatal("failed to start REST app", zap.Error(err))
		}
	}()

	// Wait for a signal to stop the server.

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	defer signal.Stop(c)

	<-c

	ctx, cancel := context.WithTimeout(context.Background(), gracefullyShutdownTimeout)
	defer cancel()

	if err := rest.Stop(ctx); err != nil {
		logger.Fatal("failed to stop REST app", zap.Error(err))
	}
}
