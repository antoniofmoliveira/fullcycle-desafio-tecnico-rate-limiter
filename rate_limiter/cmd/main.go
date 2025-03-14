package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/antoniofmoliveira/fullcycle-desafio-tecnico-rate-limiter/internal/infra/database"
	"github.com/antoniofmoliveira/fullcycle-desafio-tecnico-rate-limiter/internal/limiter"
	"github.com/antoniofmoliveira/fullcycle-desafio-tecnico-rate-limiter/internal/web/middleware"
	"github.com/go-redis/redis"
	"github.com/joho/godotenv"
)

// main initializes the server with a rate limiter middleware.
// It loads environment variables, connects to a Redis database,
// and sets up a simple HTTP server listening on port 8080.
// The server uses a rate limiter to control request rate based on IP or token.

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	if err := godotenv.Load(".env"); err != nil {
		slog.Error("godotenv.Load", "message", "Error trying to load env variables")
		return
	}

	slog.Info("server", "message", "Starting server")

	client := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
	})
	pong, err := client.Ping().Result()
	if err != nil {
		slog.Error("redis.NewClient", "message", "Error trying to connect to redis")
		return
	}
	slog.Info("redis.NewClient", "message", "Connected to redis", "pong", pong)

	db := database.NewDBRedis(client)
	rl := limiter.NewRateLimiter(&db)

	http.Handle("/", middleware.RateLimiterMiddleware(rl, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10*time.Millisecond)
		w.Write([]byte("OK"))
	})))

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		slog.Error("http.ListenAndServe", "message", "Error trying to start server")
	}

}
