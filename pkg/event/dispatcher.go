package event

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"
)

type Dispatcher[T Event] struct {
	name    string
	queue   chan *Message[T]
	stopCh  chan bool
	started int32
	timeout time.Duration
}

func NewDispatcher[T Event](name string, config *PublisherConfig) *Dispatcher[T] {
	if config.Capacity <= 0 {
		config.Capacity = 1024
	}
	if config.Timeout <= 0 {
		config.Timeout = 5 * time.Second
	}
	return &Dispatcher[T]{
		name:    name,
		queue:   make(chan *Message[T], config.Capacity),
		timeout: config.Timeout,
	}
}

func (d *Dispatcher[T]) Offer(message *Message[T]) bool {
	if message == nil {
		return false
	}
	select {
	case d.queue <- message:
		return true
	default:
		return false
	}
}

func (d *Dispatcher[T]) OfferWithTimeout(message *Message[T], timeout time.Duration) bool {
	if message == nil {
		return false
	}
	select {
	case d.queue <- message:
		return true
	case <-time.After(timeout):
		return false
	}
}

func (d *Dispatcher[T]) Start(ctx context.Context) error {
	if atomic.CompareAndSwapInt32(&d.started, 0, 1) {
		d.stopCh = make(chan bool)
		go d.worker(d.stopCh)
	}
	return nil
}

func (d *Dispatcher[T]) worker(stopCh <-chan bool) {
	for {
		select {
		default:
			d.Publish()
		case <-stopCh:
			fmt.Println("Dispatcher stopped")
			return
		}
	}
}

func (d *Dispatcher[T]) Publish() {
	select {
	case message := <-d.queue:
		message.Publish()
	case <-time.After(d.timeout):
		//fmt.Println("Publish stop by timeout!")
	}
}

func (d *Dispatcher[T]) Stop(ctx context.Context) error {
	if atomic.CompareAndSwapInt32(&d.started, 1, 0) {
		if d.stopCh != nil {
			d.stopCh <- true
			d.stopCh = nil
		}
	}
	return nil
}
