package limiter

import (
	"encoding/json"
	"log/slog"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/antoniofmoliveira/fullcycle-desafio-tecnico-rate-limiter/internal/infra/database"
	"github.com/antoniofmoliveira/fullcycle-desafio-tecnico-rate-limiter/internal/model"
)

type RateLimiter struct {
	DB database.DBInterface
	mu sync.Mutex
}

// NewRateLimiter creates a new RateLimiter with the given database interface.
// It returns a pointer to the created RateLimiter.
func NewRateLimiter(db *database.DBInterface) *RateLimiter {
	return &RateLimiter{DB: *db, mu: sync.Mutex{}}
}

// CanRequest checks if a request can be made to the server. It takes a token and a level as parameters.
// If the token is empty, it returns false. If the token is not found in the database, it creates a new
// RateLimiter with the given level and sets it in the database. If the token is found, it checks if the
// key is blocked. If it is, it checks if the block has expired. If it has, it unblocks the key. If it hasn't,
// it denies access. If the key is not blocked, it checks if there are remaining access. If there are, it
// decreases the access remain and updates the key in the database. If there aren't, it checks if it's time
// to renew the access remain. If it is, it renews it and updates the key in the database. If it's not, it
// blocks the key and updates the key in the database.
func (r *RateLimiter) CanRequest(token string, level int) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	// key not sent. Cannot process. Deny access
	if token == "" {
		slog.Info("RateLimiter.CanRequest", "message", "Token not sent")
		return false
	}
	slog.Info("RateLimiter.CanRequest", "message", "Token", "token", token, "level", level)
	// get key
	s, err := r.DB.Get(token)
	slog.Info("RateLimiter.CanRequest", "message", "Key", "key", s, "err", err)

	if err != nil { // key not found
		qt_reqs := 0
		if level == 0 { // level 0 equals IP - get qt_reqs from env
			qt_reqs, err = strconv.Atoi(os.Getenv("IP_QT_REQS_SECOND"))
			if err != nil {
				slog.Error("RateLimiter.CanRequest", "message", "Error trying to convert .env var IP_QT_REQS_SECOND to int. Defaulting to 5")
				qt_reqs = 5
			}
		} else { // level > 0 equals token - get qt_reqs from env
			qt_reqs, err = strconv.Atoi(os.Getenv("TOKEN_" + strconv.Itoa(int(level)) + "_QT_REQS_SECOND"))
			if err != nil {
				slog.Error("RateLimiter.CanRequest", "message", "Error trying to convert .env var TOKEN_"+strconv.Itoa(int(level))+"_QT_REQS_SECOND to int. Defaulting to 10")
				qt_reqs = 10
			}
		}
		nowMore1Sec := time.Now().Add(time.Second)

		rl := model.NewRateLimiter(token, level, qt_reqs-1, int(nowMore1Sec.Unix()), 0)
		sl, ok := rl.ToJson()
		if !ok {
			slog.Error("RateLimiter.CanRequest", "message", "Error trying to marshal RateLimiter to json")
			return false
		}

		r.DB.Set(token, sl)
		return true

	} else { // key found
		rl := model.RateLimiter{}
		err = json.Unmarshal([]byte(s), &rl)
		if err != nil {
			slog.Error("RateLimiter.CanRequest", "message", "Error trying to unmarshal RateLimiter from json")
			return false
		}
		slog.Info("RateLimiter.CanRequest", "message", "Unmarshaled RateLimiter", "rl", rl, "err", err)

		if rl.IsBlocked() { // key is blocked
			slog.Info("RateLimiter.CanRequest", "rl.BlockedUntil", rl.BlockedUntil, "now", int(time.Now().Unix()))
			if rl.BlockedUntil < int(time.Now().Unix()) { // unblock if block expired
				slog.Info("RateLimiter.CanRequest", "message", "Unblocking "+rl.Id)
				rl.BlockedUntil = 0
			} else { // deny access
				slog.Info("RateLimiter.CanRequest", "message", rl.Id+" key is blocked")
				return false
			}
		}

		if rl.AccessRemain > 0 { // key is not blocked and has remaining access
			slog.Info("RateLimiter.CanRequest", "message", "Access remain for "+rl.Id)
			rl.AccessRemain = rl.AccessRemain - 1
			sl, ok := rl.ToJson()
			if !ok {
				slog.Error("RateLimiter.CanRequest", "message", "Error trying to marshal RateLimiter to json")
				return false
			}
			r.DB.Set(token, sl) // update key
			return true
		} else { // not enough access remain
			if rl.LimitTime < int(time.Now().Unix()) && rl.BlockedUntil < int(time.Now().Unix()) { // renew access remain
				slog.Info("RateLimiter.CanRequest", "message", "Renewing access remain for "+rl.Id)
				rl.LimitTime = int(time.Now().Add(time.Second).Unix())
				qt_reqs := 0
				if rl.Level == 0 { // level 0 equals IP - get qt_reqs from env
					qt_reqs, err = strconv.Atoi(os.Getenv("IP_QT_REQS_SECOND"))
					if err != nil {
						slog.Error("RateLimiter.CanRequest", "message", "Error trying to convert .env var IP_QT_REQS_SECOND to int. Defaulting to 5")
					}
				} else {
					qt_reqs, err = strconv.Atoi(os.Getenv("TOKEN_" + strconv.Itoa(int(rl.Level)) +
						"_QT_REQS_SECOND"))
					if err != nil {
						slog.Error("RateLimiter.CanRequest", "message", "Error trying to convert .env var TOKEN_"+strconv.Itoa(int(rl.Level))+"_QT_REQS_SECOND to int. Defaulting to 10")
						qt_reqs = 10
					}
				}

				rl.AccessRemain = qt_reqs
				rl.BlockedUntil = 0

				sl, ok := rl.ToJson()
				if !ok {
					slog.Error("RateLimiter.CanRequest", "message", "Error trying to marshal RateLimiter to json")
					return false
				}

				r.DB.Set(token, sl)
				return true
			} else {
				// key will be blocked
				blockDuration, err := time.ParseDuration(os.Getenv("TOKEN_" + strconv.Itoa(int(rl.Level)) + "_BLOCK_DURATION"))
				if err != nil {
					slog.Error("RateLimiter.CanRequest", "message", "Error trying to convert .env var TOKEN_"+strconv.Itoa(int(rl.Level))+"_BLOCK_DURATION to time.Duration. Defaulting to 1s")
					blockDuration = time.Second
				}
				slog.Info("RateLimiter.CanRequest", "message", "Blocking "+rl.Id+" for "+blockDuration.String())
				rl.BlockedUntil = int(time.Now().Add(blockDuration).Unix())
				sl, ok := rl.ToJson()
				if !ok {
					slog.Error("RateLimiter.CanRequest", "message", "Error trying to marshal RateLimiter to json")
					return false
				}
				r.DB.Set(token, sl)
				return false
			}
		}
	}
}
