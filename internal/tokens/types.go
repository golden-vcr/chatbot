package tokens

import (
	"strings"

	"github.com/nicklaw5/helix/v2"
)

type Store interface {
	Save(credentials *helix.AccessCredentials) error
	Load() (*helix.AccessCredentials, error)
	Clear() error
}

func NewStore(root, botUsername string) Store {
	return &diskStore{
		root:        root,
		botUsername: strings.ToLower(botUsername),
	}
}
