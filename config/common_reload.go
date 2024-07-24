package config

import "sync"

var trueValue = true

type config[T any] struct {
	mu              sync.Mutex
	reloadListeners []chan T
}

func (c *config[T]) announceReloaded(conf T) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, ch := range c.reloadListeners {
		ch <- conf
	}
}

func (c *config[T]) ListenReload(cb func(T)) func() {
	c.mu.Lock()
	defer c.mu.Unlock()

	ch := make(chan T)
	c.reloadListeners = append(c.reloadListeners, ch)

	stop := make(chan struct{})

	go func() {
		for {
			select {
			case conf := <-ch:
				cb(conf)
			case <-stop:
				c.UnlistenReload(ch)
				return
			}
		}
	}()

	return func() {
		close(stop)
	}
}

func (c *config[T]) UnlistenReload(ch chan T) {
	c.mu.Lock()
	defer c.mu.Unlock()

	newListeners := make([]chan T, 0, len(c.reloadListeners)-1)

	for _, l := range c.reloadListeners {
		if l != ch {
			newListeners = append(newListeners, l)
		}
	}

	c.reloadListeners = newListeners
}
