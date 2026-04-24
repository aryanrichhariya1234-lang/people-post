package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"peoplepost/internal/cache"
	"peoplepost/internal/config"
	"peoplepost/internal/routes"
)

func main() {
	initApp()

	router := routes.SetupRouter()

	port := getPort()

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go startServer(server, port)

	gracefulShutdown(server)
}

func initApp() {
	config.LoadEnv()
	config.ConnectMongo()
	cache.InitRedis()
	config.InitCloudinary()
}

func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		return "8080"
	}
	return port
}

func startServer(server *http.Server, port string) {
	log.Printf("Server running on port %s\n", port)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v\n", err)
	}
}

func gracefulShutdown(server *http.Server) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Forced shutdown: %v\n", err)
	}

	log.Println("Server exited cleanly")
}