package cache

import (
	"context"
	"crypto/md5"
	"encoding/hex"
)

type Cacher interface {
	Set(ctx context.Context, key string, val *[]byte) (string, error)
	Get(ctx context.Context, key string) (string, bool)
	Delete(ctx context.Context, key string) error
}

func generateCacheValue(val *[]byte) (string, error) {
	h := md5.New()
	h.Write(*val)
	hashBytes := h.Sum(nil)

	return hex.EncodeToString(hashBytes), nil
}

/*
The Conditioner interface should be used for specific cache conditions.

It is loosely based on the concepts of "If-Match" and "If-None-Match" http headers.
*/
type Conditioner interface {
	Met(key, val string) bool
}

type Match struct {
	cache Cacher
}

func NewMatch(cache Cacher) *Match {
	return &Match{
		cache: cache,
	}
}

func (m Match) Met(ctx context.Context, key, val string) bool {
	cachedVal, ok := m.cache.Get(ctx, key)

	if !ok {
		return false
	}

	if val == cachedVal {
		return true
	}

	return false
}

type NoneMatch struct {
	cache Cacher
}

func NewNoneMatch(cache Cacher) *NoneMatch {
	return &NoneMatch{
		cache: cache,
	}
}

func (m NoneMatch) Met(ctx context.Context, key, val string) bool {
	cacheVal, ok := m.cache.Get(ctx, key)

	if !ok {
		return false
	}

	if cacheVal != val {
		return false
	}

	return true
}
