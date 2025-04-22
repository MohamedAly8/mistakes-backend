package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/rs/cors"

	"app/mistakes/internal/database"
	"app/mistakes/internal/handlers"
)

func getRoot(w http.ResponseWriter, r *http.Request) {
    fmt.Printf("got / request\n")
    io.WriteString(w, "This is my website!\n")
}

func getHello(w http.ResponseWriter, r *http.Request) {
    fmt.Printf("got /hello request\n")
    io.WriteString(w, "Hello, HTTP!\n")
}

func main() {
    // --- Load environment variables from .env file ---
    if _, err := os.Stat(".env"); err == nil {
        err := godotenv.Load()
        if err != nil {
            log.Printf("Warning: Error loading .env file: %v\n", err)
        }
    } else if !os.IsNotExist(err) {
        log.Printf("Warning: Error checking for .env file: %v\n", err)
    }
    // --- End .env loading ---

    // Initialize Database Connection
    log.Println("Initializing database connection...")
    if err := database.InitDB(); err != nil {
        log.Fatalf("Could not initialize database: %v", err)
    }
    log.Println("Database connection established.")

    // --- Graceful Shutdown Setup ---
    stop := make(chan os.Signal, 1)
    signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
    defer database.CloseDB()

    r := mux.NewRouter()

    // Define your routes
    r.HandleFunc("/", getRoot).Methods("GET")
    r.HandleFunc("/hello", getHello).Methods("GET")
    r.HandleFunc("/mistakes", handlers.GetMistakes).Methods("GET")
    r.HandleFunc("/mistakes", handlers.CreateMistake).Methods("POST")

    // --- CORS Middleware Setup ---
    corsOptions := cors.New(cors.Options{
        AllowedOrigins: []string{"*"},
        AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowedHeaders: []string{"Content-Type"},
        AllowCredentials: true,
    })
    handler := corsOptions.Handler(r)
    // --- End CORS Middleware Setup ---

    log.Println("Starting server...")

    // Create the HTTP Server instance
    srv := &http.Server{
        Addr:           ":8080",
        Handler:        handler, // Use the CORS-wrapped handler
        ReadTimeout:    10 * time.Second,
        WriteTimeout:   10 * time.Second,
        IdleTimeout:    120 * time.Second,
    }

    // Run the server in a goroutine so it doesn't block
    go func() {
        fmt.Println("Server running")
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Could not start server: %v\n", err)
        }
    }()

    // Wait for interrupt signal to gracefully shut down the server
    <-stop
    log.Println("Shutting down server...")

    // Create a context with a timeout to allow active requests to finish
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // Shutdown the server gracefully
    if err := srv.Shutdown(ctx); err != nil {
        log.Printf("Server forced to shutdown: %v\n", err)
    } else {
        log.Println("Server gracefully stopped.")
    }

    // database.CloseDB() is deferred and will be called now
}