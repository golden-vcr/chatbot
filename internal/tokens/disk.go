package tokens

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/nicklaw5/helix/v2"
)

type diskStore struct {
	root        string
	botUsername string
}

func (s *diskStore) Save(credentials *helix.AccessCredentials) error {
	// Get the path to our file and ensure the parent directory exists
	if err := os.MkdirAll(s.root, os.ModePerm); err != nil {
		return err
	}
	path := filepath.Join(s.root, s.botUsername+".json")

	// Write our credentials, JSON-serialized, to disk
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(credentials)
}

func (s *diskStore) Load() (*helix.AccessCredentials, error) {
	// Open the credentials file for read, if it exists
	path := filepath.Join(s.root, s.botUsername+".json")
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Parse the JSON-serialized credentials from that file
	var credentials helix.AccessCredentials
	if err := json.NewDecoder(f).Decode(&credentials); err != nil {
		return nil, err
	}
	return &credentials, nil
}

func (s *diskStore) Clear() error {
	path := filepath.Join(s.root, s.botUsername+".json")
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}
