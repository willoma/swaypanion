package swaypanion

import (
	"sync"
	"time"
)

const (
	subscriptionBuffer     = 2
	subscriptionDedupDelay = 250 * time.Millisecond
)

type subscription[T comparable] struct {
	mu        sync.Mutex
	receivers map[uint32]chan T
	nextID    uint32
	lastValue T
	lastSend  time.Time
}

func newSubscription[T comparable]() *subscription[T] {
	return &subscription[T]{
		receivers: make(map[uint32]chan T),
	}
}

func (s *subscription[T]) Subscribe() (channel <-chan T, id uint32) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ch := make(chan T, subscriptionBuffer)

	id = s.nextID
	s.nextID++
	s.receivers[id] = ch
	s.nextID++

	return ch, id
}

func (s *subscription[T]) Unsubscribe(id uint32) {
	s.mu.Lock()
	defer s.mu.Unlock()

	close(s.receivers[id])
	delete(s.receivers, id)
}

func (s *subscription[T]) Publish(value T) (published bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()

	if value == s.lastValue && s.lastSend.Add(subscriptionDedupDelay).After(now) {
		return false
	}

	s.lastValue = value
	s.lastSend = now

	for _, ch := range s.receivers {
		ch <- value
	}

	return true
}
