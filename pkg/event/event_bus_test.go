package event

import (
	"context"
	"fmt"
	"testing"
	"time"
)

type TestEvent struct {
	AbstractEvent
}

type TestEventHandler struct {
	EventHandler[TestEvent]
}

func (t *TestEventHandler) Handler(event TestEvent) {
	fmt.Println(event.GetSource())
}

func TestPublisher(t *testing.T) {
	publisher := NewGoEventBus[TestEvent]().GetPublisher("event.common", "default")
	err := publisher.Start(context.Background())
	if err != nil {
		return
	}
	predicateEventHandler := NewPredicateEventHandler[TestEvent](
		func(event TestEvent) bool {
			return true
		},
		&TestEventHandler{},
	)
	publisher.AddHandler(predicateEventHandler)

	publisher.Offer(TestEvent{
		AbstractEvent{
			Source: "hello1",
		},
	})
	publisher.Offer(TestEvent{
		AbstractEvent{
			Source: "hello2",
			Target: &TestEventHandler{},
		},
	})
	publisher.Offer(TestEvent{
		AbstractEvent{
			Source: "hello3",
		},
	})
	publisher.Offer(TestEvent{
		AbstractEvent{
			Source: "hello4",
		},
	})
	publisher.Offer(TestEvent{
		AbstractEvent{
			Source: "hello5",
			Target: predicateEventHandler,
		},
	})
	time.Sleep(15 * time.Second)
}
