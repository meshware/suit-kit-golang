package event

type Message[T Event] struct {
	event    T
	consumer func(event T)
}

func NewMessage[T Event](event T, consumer func(e T)) *Message[T] {
	return &Message[T]{
		event:    event,
		consumer: consumer,
	}
}

func (m *Message[T]) Publish() {
	m.consumer(m.event)
}
