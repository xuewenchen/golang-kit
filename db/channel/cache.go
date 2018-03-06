package channel

import (
	"errors"
	"sync"
)

var ErrFull = errors.New("cache chan full")

type Cache struct {
	ch chan func()
}

func New(size int, wait *sync.WaitGroup) *Cache {
	c := &Cache{
		ch: make(chan func(), size),
	}
	go c.proc(wait)
	return c
}

func (c *Cache) proc(wait *sync.WaitGroup) {
	for {
		select {
		case f, ok := <-c.ch:
			if !ok {
				wait.Done()
				return
			}
			f()
		}
	}
}

func (c *Cache) Save(f func()) (err error) {
	select {
	case c.ch <- f:
	default:
		err = ErrFull
	}
	return
}

func (c *Cache) Close() {
	close(c.ch)
}
