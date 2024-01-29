package csrf

import (
	"context"
	"sync"
	"time"
)

type Buffer interface {
	Peek() string
	Contains(token string) bool
}

func NewBuffer(ctx context.Context, generationInterval time.Duration, capacity int) Buffer {
	if capacity < 1 {
		panic("csrf.Buffer capacity must be >= 1")
	}
	b := &buffer{
		items: make([]string, 0, capacity),
	}
	b.items = append(b.items, generateToken())
	go func() {
		done := false
		for !done {
			select {
			case <-ctx.Done():
				done = true
			case <-time.After(generationInterval):
				b.add(generateToken())
			}
		}
	}()
	return b
}

type buffer struct {
	generationInterval time.Duration
	items              []string
	mu                 sync.RWMutex
}

func (b *buffer) add(token string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.items) == cap(b.items) {
		b.items = append(b.items[1:], token)
	} else {
		b.items = append(b.items, token)
	}
}

func (b *buffer) Peek() string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.items[len(b.items)-1]
}

func (b *buffer) Contains(token string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, s := range b.items {
		if token == s {
			return true
		}
	}
	return false
}
