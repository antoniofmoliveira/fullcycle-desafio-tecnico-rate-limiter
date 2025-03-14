package model

import (
	"encoding/json"
	"log/slog"
)

type RateLimiter struct {
	Id           string `json:"id"`
	Level        int    `json:"level"`
	AccessRemain int    `json:"accessRemain"`
	LimitTime    int    `json:"initTime"`
	BlockedUntil int    `json:"blockedUntil"`
}

// NewRateLimiter creates a new RateLimiter with the given id, level, initial number of remaining accesses,
// initial time of the rate limit, and time until the key is blocked.
func NewRateLimiter(id string, level int, accessRemain int, initTime int, blockedUntil int) *RateLimiter {
	return &RateLimiter{
		Id:           id,
		Level:        level,
		AccessRemain: accessRemain,
		LimitTime:    initTime,
		BlockedUntil: blockedUntil,
	}
}

// IsBlocked returns true if the rate limiter is blocked until a certain time.
// It will return false if it is not blocked.
func (r *RateLimiter) IsBlocked() bool {
	if r.BlockedUntil == 0 {
		return false
	}
	return true
}

// ToJson marshals the RateLimiter to a json string.
// If there is an error during marshaling, it logs the error and returns an empty string and false.
// Otherwise, it returns the json string and true.
func (r *RateLimiter) ToJson() (string, bool) {
	j, err := json.Marshal(r)
	if err != nil {
		slog.Error("RateLimiter.ToJson", "message", "Error trying to marshal RateLimiter to json")
		return "", false
	}
	return string(j), true
}
