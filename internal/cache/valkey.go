package cache

import (
	"context"

	"github.com/valkey-io/valkey-go"
)

const EMPTY_STRING string = ""

type ValkeyCache struct {
	client valkey.Client
}

func (v *ValkeyCache) Set(ctx context.Context, key string, val *[]byte) (string, error) {
	genVal, err := generateCacheValue(val)

	if err != nil {
		return EMPTY_STRING, err
	}

	err = v.client.Do(
		ctx,
		v.client.B().Set().Key(key).Value(genVal).Build()).Error()

	if err != nil {
		return EMPTY_STRING, err
	}

	return genVal, nil
}

func (v *ValkeyCache) Get(ctx context.Context, key string) (string, bool) {
	val, err := v.client.Do(ctx, v.client.B().Get().Key(key).Build()).ToString()

	if err != nil {
		return EMPTY_STRING, false
	}

	return val, true
}

func (v *ValkeyCache) Delete(ctx context.Context, key string) error {
	err := v.client.Do(ctx, v.client.B().Del().Key(key).Build()).Error()

	if err != nil {
		return err
	}

	return nil
}

func NewValkeyCache(client valkey.Client) *ValkeyCache {
	return &ValkeyCache{
		client: client,
	}
}
