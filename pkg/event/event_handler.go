package event

type EventHandler[T Event] interface {
	Handler(event T)
}

type Predicate[T Event] func(event T) bool

type PredicateEventHandler[T Event] struct {
	EventHandler[T]
	PredicateFunc Predicate[T]
}

func (pe *PredicateEventHandler[T]) Handler(event T) {
	if pe.PredicateFunc(event) {
		pe.EventHandler.Handler(event)
	}
}

func NewPredicateEventHandler[T Event](predicateFunc Predicate[T], handler EventHandler[T]) EventHandler[T] {
	return &PredicateEventHandler[T]{
		EventHandler:  handler,
		PredicateFunc: predicateFunc,
	}
}
