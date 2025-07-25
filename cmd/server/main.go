package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ripple/db"
	"ripple/handlers"

	"github.com/gorilla/mux"
)

func main() {
	// Parse command line flags
	mongoURI := flag.String("mongo-uri", "mongodb://localhost:27017", "MongoDB connection URI")
	dbName := flag.String("db-name", "agent_metrics", "MongoDB database name")
	port := flag.String("port", "9999", "HTTP server port")
	flag.Parse()

	// Connect to MongoDB
	mongodb, err := db.NewMongoDB(*mongoURI, *dbName)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer mongodb.Close()

	// Create repositories
	agentRepo := db.NewAgentRepository(mongodb)
	uiRepo := db.NewUIRepository(mongodb)

	// Create handlers
	agentHandler := handlers.NewAgentHandler(agentRepo)
	uiHandler := handlers.NewUIHandler(uiRepo)

	// Create router
	router := mux.NewRouter()

	// Register routes
	agentHandler.RegisterRoutes(router)
	uiHandler.RegisterRoutes(router)

	// Create server
	srv := &http.Server{
		Addr:         ":" + *port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server listening on port %s", *port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Create a deadline to wait for
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly")
}
