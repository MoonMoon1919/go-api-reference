package cache

import "context"

type InMemoryCache struct {
	items map[string]string
}

func (m *InMemoryCache) Set(ctx context.Context, key string, val *[]byte) (string, error) {
	genVal, err := generateCacheValue(val)

	if err != nil {
		return "", err
	}

	m.items[key] = genVal

	return genVal, nil
}

func (m *InMemoryCache) Get(ctx context.Context, key string) (string, bool) {
	val, ok := m.items[key]

	return val, ok
}

func (m *InMemoryCache) Delete(ctx context.Context, key string) error {
	delete(m.items, key)

	return nil
}

func NewInMemoryCache() *InMemoryCache {
	return &InMemoryCache{
		items: make(map[string]string),
	}
}
