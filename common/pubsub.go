package common

import (
	"sync"
	"time"
)

const subscriptionBuffer = 3

type Config[T Data[T]] struct {
	PollInterval time.Duration
	PollFn       func() (value T, ok bool)
}

type Pubsub[T Data[T]] struct {
	config Config[T]

	mu           sync.Mutex
	subscribers  map[any]chan T
	currentValue T
	pollStop     chan struct{}
}

func NewPubsub[T Data[T]](config ...Config[T]) *Pubsub[T] {
	p := &Pubsub[T]{
		subscribers: map[any]chan T{},
	}

	if len(config) > 0 {
		p.Reconfigure(config[0])
	}

	return p
}

func (p *Pubsub[T]) Subscribe(id any, withInitialValue bool, callback func(T)) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, ok := p.subscribers[id]; ok {
		// Already subscribed
		return
	}

	ch := make(chan T, subscriptionBuffer)

	if id == nil {
		id = ch
	}

	p.subscribers[id] = ch

	p.unsafeStartPoll()

	if withInitialValue {
		ch <- p.currentValue
	}

	go func() {
		for {
			value, ok := <-ch
			if !ok {
				break
			}

			callback(value)
		}
	}()
}

func (p *Pubsub[T]) Unsubscribe(id any) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for thisId, ch := range p.subscribers {
		if id == thisId {
			delete(p.subscribers, thisId)
			close(ch)
		}
	}

	if len(p.subscribers) == 0 {
		p.unsafeStopPoll()
	}
}

func (p *Pubsub[T]) Reconfigure(config Config[T]) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.subscribers) == 0 {
		// No subscriber, we just need to store the configuration
		p.config = config
		return
	}

	p.unsafeStopPoll()

	p.config = config

	p.unsafeStartPoll()
}

func (p *Pubsub[T]) Publish(value T) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.unsafePublish(value)
}

func (s *Pubsub[T]) unsafePublish(value T) {
	if value.Equal(s.currentValue) {
		return
	}

	s.currentValue = value

	for _, r := range s.subscribers {
		r <- value
	}
}

func (p *Pubsub[T]) unsafeStartPoll() {
	if p.config.PollFn == nil || p.pollStop != nil {
		return
	}

	if currentValue, ok := p.config.PollFn(); ok {
		p.currentValue = currentValue
	}

	p.pollStop = make(chan struct{})

	go func() {
		ticker := time.NewTicker(p.config.PollInterval)

		for {
			select {
			case <-ticker.C:
				if value, ok := p.config.PollFn(); ok {
					p.Publish(value)
				}
			case <-p.pollStop:
				ticker.Stop()
				return
			}
		}
	}()
}

func (p *Pubsub[T]) unsafeStopPoll() {
	if p.pollStop != nil {
		close(p.pollStop)

		p.pollStop = nil
	}
}
