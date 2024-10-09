package storage

import (
	"chat-cli/internal/logger"
	"encoding/json"
	"fmt"
	"log"
	"os"
)

const chatStoragePath = "chat-storage.json"

type Storage struct {
	Username     string
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func Load() *Storage {
	var storage *Storage
	data, err := os.ReadFile(chatStoragePath)
	if err != nil && os.IsNotExist(err) {
		storage = &Storage{
			Username:     "",
			AccessToken:  "",
			RefreshToken: "",
		}

		storage.Flush()

		return storage
	} else if err != nil {
		logger.ErrorWithExit("failed to init storage: %s", err)
	}

	err = json.Unmarshal(data, &storage)
	if err != nil {
		logger.ErrorWithExit("failed to load storage: %s", err)
	}

	return storage
}

func (s *Storage) SetRefreshToken(token string) {
	s.RefreshToken = token
}

func (s *Storage) GetRefreshToken() string {
	return s.RefreshToken
}

func (s *Storage) SetAccessToken(token string) {
	s.AccessToken = token
}

func (s *Storage) GetAccessToken() string {
	return s.AccessToken
}

func (s *Storage) GetAuthHeader() string {
	token := s.GetAccessToken()
	if token == "" {
		log.Fatalf("you need to log in first")
	}
	return fmt.Sprintf("Bearer %s", token)
}

func (s *Storage) SetUsername(username string) {
	s.Username = username
}

func (s *Storage) GetUsername() string {
	return s.Username
}

func (s *Storage) Flush() {
	data, err := json.Marshal(s)
	if err != nil {
		logger.ErrorWithExit("failed to encode storage data: %s", err)
	}

	err = os.WriteFile(chatStoragePath, data, 0644)
	if err != nil {
		logger.ErrorWithExit("failed to write to storage file: %s", err)
	}
}
