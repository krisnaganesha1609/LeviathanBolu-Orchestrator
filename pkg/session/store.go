// Package session manages per-device conversation history in Redis.
// Each device gets its own key so multiple Flutter clients (phone, tablet,
// desktop) can each maintain an independent conversation with the assistant.
package session

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/krisnaganesha1609/LeviathanBolu-BE/internal/llm"
	"github.com/redis/go-redis/v9"
)

const keyPrefix = "leviathan:session:"

// Store is a Redis-backed store for per-device conversation history.
type Store struct {
	rdb *redis.Client
	ttl time.Duration // how long a session lives after the last message
}

func NewStore(rdb *redis.Client, ttl time.Duration) *Store {
	return &Store{rdb: rdb, ttl: ttl}
}

func (s *Store) historyKey(deviceID string) string {
	return keyPrefix + deviceID + ":history"
}

// GetHistory returns the stored conversation history for a device.
// Returns nil (not an error) if the device has no history yet.
func (s *Store) GetHistory(ctx context.Context, deviceID string) ([]llm.Message, error) {
	val, err := s.rdb.Get(ctx, s.historyKey(deviceID)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil // fresh session, not an error
		}
		return nil, fmt.Errorf("session: get history for %q: %w", deviceID, err)
	}

	var history []llm.Message
	if err := json.Unmarshal(val, &history); err != nil {
		// Corrupted history — treat as empty and let the session restart.
		return nil, nil
	}
	return history, nil
}

// SaveHistory persists the full conversation history for a device and
// resets the TTL. Call this after every completed turn.
func (s *Store) SaveHistory(ctx context.Context, deviceID string, history []llm.Message) error {
	b, err := json.Marshal(history)
	if err != nil {
		return fmt.Errorf("session: marshal history for %q: %w", deviceID, err)
	}
	if err := s.rdb.Set(ctx, s.historyKey(deviceID), b, s.ttl).Err(); err != nil {
		return fmt.Errorf("session: save history for %q: %w", deviceID, err)
	}
	return nil
}

// ClearHistory deletes the conversation history for a device.
// Used when the user says "forget everything" or starts a new conversation.
func (s *Store) ClearHistory(ctx context.Context, deviceID string) error {
	if err := s.rdb.Del(ctx, s.historyKey(deviceID)).Err(); err != nil && !errors.Is(err, redis.Nil) {
		return fmt.Errorf("session: clear history for %q: %w", deviceID, err)
	}
	return nil
}
