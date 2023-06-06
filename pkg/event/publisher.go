package event

import (
	"context"
	"suit-kit-golang/pkg/lifecycle"
	"sync"
	"time"
)

type Publisher[T Event] interface {
	lifecycle.Start
	lifecycle.Stop
	AddHandler(handler EventHandler[T]) bool
	RemoveHandler(handler EventHandler[T]) bool
	Size() int
	Offer(event T) bool
	OfferWithTimeout(event T, duration time.Duration) bool
}

// PublisherConfig Configure publisher-related queue length and timeout
type PublisherConfig struct {
	Capacity uint64        `json:"capacity"`
	Timeout  time.Duration `json:"timeout"`
}

type GoPublisher[T Event] struct {
	Publisher[T]
	Name     string
	Group    *PublisherGroup[T]
	Polling  *Dispatcher[T]
	Handlers map[interface{}]EventHandler[T]
	Consumer func(event T)
}

func NewGoPublisher[T Event](name string, group *PublisherGroup[T]) *GoPublisher[T] {
	return &GoPublisher[T]{
		Name:     name,
		Group:    group,
		Polling:  group.dispatcher,
		Handlers: map[interface{}]EventHandler[T]{},
	}
}

func (gp *GoPublisher[T]) AddHandler(handler EventHandler[T]) bool {
	if handler != nil {
		gp.Handlers[handler] = handler
		return true
	} else {
		return false
	}
}

func (gp *GoPublisher[T]) RemoveHandler(handler EventHandler[T]) bool {
	if handler != nil {
		delete(gp.Handlers, handler)
		return true
	} else {
		return false
	}
}

func (gp *GoPublisher[T]) Size() int {
	return len(gp.Handlers)
}

func (gp *GoPublisher[T]) Offer(event T) bool {
	return gp.Polling != nil && gp.Polling.Offer(NewMessage(event, gp.publish))
}

func (gp *GoPublisher[T]) OfferWithTimeout(event T, duration time.Duration) bool {
	return gp.Polling != nil && gp.Polling.OfferWithTimeout(NewMessage(event, gp.publish), duration)
}

func (gp *GoPublisher[T]) publish(event T) {
	if event.GetTarget() != nil {
		handler := gp.Handlers[event.GetTarget()]
		if handler != nil {
			handler.Handler(event)
			return
		}
	} else {
		for _, handler := range gp.Handlers {
			handler.Handler(event)
		}
	}
}

func (gp *GoPublisher[T]) Start(ctx context.Context) error {
	if gp.Polling == nil {
		if !gp.Group.Contains(gp.Name) {
			gp.Group.publishers[gp.Name] = gp
		}
		gp.Polling = gp.Group.dispatcher
	}
	return gp.Polling.Start(ctx)
}

func (gp *GoPublisher[T]) Stop(ctx context.Context) error {
	//移除，防止保留大量的无用的发布器
	gp.Group.remove(gp.Name)
	return gp.Polling.Stop(ctx)
}

type PublisherGroup[T Event] struct {
	name       string
	config     *PublisherConfig
	dispatcher *Dispatcher[T]
	publishers map[string]*GoPublisher[T]
	mu         sync.Mutex
}

func NewPublisherGroup[T Event](name string, config *PublisherConfig) *PublisherGroup[T] {
	if config == nil {
		config = &PublisherConfig{}
	}
	capacity := config.Capacity
	if capacity <= 0 {
		capacity = 1024
	}
	dispatcher := NewDispatcher[T]("GoEventBus-"+name, config)
	return &PublisherGroup[T]{
		name:       name,
		config:     config,
		dispatcher: dispatcher,
		publishers: make(map[string]*GoPublisher[T]),
	}
}

func (pg *PublisherGroup[T]) Contains(name string) bool {
	pg.mu.Lock()
	defer pg.mu.Unlock()
	_, ok := pg.publishers[name]
	return ok
}

func (pg *PublisherGroup[T]) remove(name string) *GoPublisher[T] {
	pg.mu.Lock()
	defer pg.mu.Unlock()
	publisher := pg.publishers[name]
	delete(pg.publishers, name)
	return publisher
}

func (pg *PublisherGroup[T]) GetPublisher(name string) *GoPublisher[T] {
	pg.mu.Lock()
	defer pg.mu.Unlock()
	publisher, ok := pg.publishers[name]
	if !ok {
		publisher = NewGoPublisher[T](name, pg)
		pg.publishers[name] = publisher
	}
	return publisher
}
