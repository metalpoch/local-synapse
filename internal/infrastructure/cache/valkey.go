package cache

import "github.com/valkey-io/valkey-go"

func NewValkeyClient(addr string) *valkey.Client {
	client, err := valkey.NewClient(valkey.ClientOption{InitAddress: []string{addr}})
	if err != nil {
		panic(err)
	}

	return &client
}
