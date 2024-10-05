package storage

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

const chatStoragePath = "chat-storage.json"

type Storage struct {
	AccessTokens  map[string]string `json:"access_tokens"`
	RefreshTokens map[string]string `json:"refresh_tokens"`
}

func Load() *Storage {
	var storage *Storage
	data, err := os.ReadFile(chatStoragePath)
	if err != nil && os.IsNotExist(err) {
		storage = &Storage{
			AccessTokens:  make(map[string]string),
			RefreshTokens: make(map[string]string),
		}

		storage.Flush()

		return storage
	} else if err != nil {
		log.Fatalf("failed to init storage: %s", err)
	}

	err = json.Unmarshal(data, storage)
	if err != nil {
		log.Fatalf("failed to load storage: %s", err)
	}

	return storage
}

func (s *Storage) SetRefreshToken(username, token string) {
	s.RefreshTokens[username] = token
}

func (s *Storage) GetRefreshToken(username string) string {
	return s.RefreshTokens[username]
}

func (s *Storage) SetAccessToken(username, token string) {
	s.AccessTokens[username] = token
}

func (s *Storage) GetAccessToken(username string) string {
	return s.AccessTokens[username]
}

func (s *Storage) Flush() {
	data, err := json.Marshal(s)
	if err != nil {
		log.Fatalf("failed to encode storage data: %s", err)
	}

	err = os.WriteFile(chatStoragePath, data, 0644)
	if err != nil {
		log.Fatalf("failed to write to storage file: %s", err)
	}
}

func (s *Storage) GetAuthHeader(username string) string {
	return fmt.Sprintf("Bearer %s", s.GetAccessToken(username))
}
