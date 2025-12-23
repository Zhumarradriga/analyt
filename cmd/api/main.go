package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"analytservice/internal/handler"
	"analytservice/internal/repository"
	"analytservice/internal/service"
)

func main() {
	// хранилище, пока что кликхаус
	// креды, лучше из env конечн
	repo, err := repository.NewClickHouseRepo("5.35.125.252:9000", "default", "default", "my_secret_password")
	if err != nil {
		log.Fatalf("ClickHouse init failed: %v", err)
	}
	defer repo.Close()

	analyticsSvc := service.NewAnalyticsService(repo, 10000, 2*time.Second)
	analyticsSvc.Start()

	r := gin.Default()
	eventHandler := handler.NewEventHandler(analyticsSvc)
	eventHandler.RegisterRoutes(r)

	srv := &http.Server{
		Addr:    ":8100",
		Handler: r,
	}

	go func() {
		log.Println("Server started on :8100")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Flushing remaining events...")
	analyticsSvc.Stop()

	log.Println("Server exited properly")
}
