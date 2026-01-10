package cache

import (
	"log"

	"github.com/valkey-io/valkey-go"
)

func NewValkeyClient(addr, password string) *valkey.Client {
	client, err := valkey.NewClient(valkey.ClientOption{InitAddress: []string{addr}, Password: password})
	if err != nil {
		log.Printf("Warning: Could not connect to Valkey at %s: %v. Caching will be disabled.", addr, err)
		return nil
	}

	return &client
}
