package api

import (
	"context"
	"github.com/StepanShevelev/test-test/config"
	"github.com/StepanShevelev/test-test/web"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func InitBackendApi() {

	cfg := config.New()
	if err := cfg.Load("./configs", "config", "yml"); err != nil {
		log.Fatal(err)
	}

	srv := web.NewServer(cfg, web.Init())

	go func() {
		log.Printf("Running HTTP server on :%s", cfg.Port)
		srv.Run()
	}()

	// catch signals for quit from application
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// graceful shutdown
	ctx, shutdown := context.WithCancel(context.Background())
	defer shutdown()

	if err := srv.Stop(ctx); err != nil {
		log.Printf("Failed to stop server: %v", err)
	} else {
		log.Println("Graceful shutdown")
	}
}
