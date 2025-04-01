package cache

import (
	"context"
	"log"
	"sync"

	"github.com/go-gorm/caches/v4"
)

type MemoryCacher struct {
	store *sync.Map
}

func (c *MemoryCacher) init() {
	if c.store == nil {
		c.store = &sync.Map{}
	}
}

func (c *MemoryCacher) Get(ctx context.Context, key string, q *caches.Query[any]) (*caches.Query[any], error) {
	c.init()
	val, ok := c.store.Load(key)
	if !ok {
		return nil, nil
	}
	log.Printf("cache hit: %s", key)
	if err := q.Unmarshal(val.([]byte)); err != nil {
		return nil, err
	}

	return q, nil
}

func (c *MemoryCacher) Store(ctx context.Context, key string, val *caches.Query[any]) error {
	c.init()
	res, err := val.Marshal()
	if err != nil {
		return err
	}
	c.store.Store(key, res)
	log.Printf("cache: %v", c.store)
	return nil
}

func (c *MemoryCacher) Invalidate(ctx context.Context) error {
	c.store = &sync.Map{}
	return nil
}
