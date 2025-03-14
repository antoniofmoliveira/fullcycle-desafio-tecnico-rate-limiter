package middleware

import (
	"net/http"
	"os"
	"strconv"

	"log/slog"

	"github.com/antoniofmoliveira/fullcycle-desafio-tecnico-rate-limiter/internal/limiter"
	rip "github.com/vikram1565/request-ip"
)

// RateLimiterMiddleware is a middleware function that controls the rate of incoming requests
// using a rate limiter. It checks the request's IP address and/or API key to determine
// whether the request is allowed based on the configured rate limits.
// The middleware can be configured to use only IP-based or token-based rate limiting
// through environment variables USE_ONLY_IP_LIMITER and USE_ONLY_TOKEN_LIMITER.
// If both environment variables are set to true, an error is logged and the request is passed through.
// When a request exceeds the rate limit, a 429 HTTP status code is returned with a relevant message.

func RateLimiterMiddleware(rl *limiter.RateLimiter, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// USE_ONLY_IP_LIMITER is set to true
		useOnlyIP := os.Getenv("USE_ONLY_IP_LIMITER")
		useOnlyIPBool, err := strconv.ParseBool(useOnlyIP)
		if err != nil {
			slog.Info("RateLimiterMiddleware", "message", "Error trying to convert USE_ONLY_IP_LIMITER to bool")
			useOnlyIPBool = false
		}

		// USE_ONLY_TOKEN_LIMITER is set to true
		userOnlyToken := os.Getenv("USE_ONLY_TOKEN_LIMITER")
		userOnlyTokenBool, err := strconv.ParseBool(userOnlyToken)
		if err != nil {
			slog.Info("RateLimiterMiddleware", "message", "Error trying to convert USE_ONLY_TOKEN_LIMITER to bool")
			userOnlyTokenBool = false
		}

		// Both USE_ONLY_IP_LIMITER and USE_ONLY_TOKEN_LIMITER are set to true. This is a configuration error. Bypassing middleware.
		if useOnlyIPBool && userOnlyTokenBool {
			slog.Error("RateLimiterMiddleware", "message", ".env vars USE_ONLY_IP_LIMITER and USE_ONLY_TOKEN_LIMITER cannot be true at the same time")
			h.ServeHTTP(w, r)
			return
		}

		// Get IP
		ip := rip.GetClientIP(r)
		slog.Info("RateLimiterMiddleware", "message", "IP", "ip", ip)
		// default level
		level := 0
		// Get token
		token := r.Header.Get("API_KEY")
		slog.Info("RateLimiterMiddleware", "message", "Token", "token", token)
		// Simple validation of token. Level is first digit of token
		if token != "" {
			level, err = strconv.Atoi(string(token[0]))
			if err != nil {
				slog.Info("RateLimiterMiddleware", "message", "token malformed")
				http.Error(w, "token malformed", http.StatusUnauthorized)
				return
			}
		}
		if level > 4 {
			slog.Info("RateLimiterMiddleware", "message", "token malformed")
			http.Error(w, "token malformed", http.StatusUnauthorized)
			return
		}

		// If USE_ONLY_IP_LIMITER is true, use IP as token
		if useOnlyIPBool {
			token = ip
		} else if userOnlyTokenBool {
			// If USE_ONLY_TOKEN_LIMITER is true, use token
			if token == "" { // token not found
				slog.Info("RateLimiterMiddleware", "message", ".env var USE_ONLY_TOKEN_LIMITER is true and token not found")
				http.Error(w, "token not found", http.StatusUnauthorized)
				return
			}
		} else if token == "" { // Both USE_ONLY_IP_LIMITER and USE_ONLY_TOKEN_LIMITER are false and token not found
			slog.Info("RateLimiterMiddleware", "message", "token not found. Using IP as token")
			token = ip
		}

		// Can proceed?
		canRequest := rl.CanRequest(token, level)
		slog.Info("RateLimiterMiddleware", "message", "CanRequest", "canRequest", canRequest)

		if canRequest {
			// Handle is allowed to proceed
			h.ServeHTTP(w, r)
			return
		} else {
			// Request recused
			slog.Info("RateLimiterMiddleware", "message", token+" have reached the maximum number of requests or actions allowed within a certain time frame")
			http.Error(w, "you have reached the maximum number of requests or actions allowed within a certain time frame", http.StatusTooManyRequests)
			return
		}
	})

}
