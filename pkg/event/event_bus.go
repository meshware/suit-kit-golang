package event

type EventBus[T Event] interface {
	GetPublisherByConfig(group, name string, config *PublisherConfig) Publisher[T]
	GetPublisher(group, name string) Publisher[T]
}

type GoEventBus[T Event] struct {
	EventBus[T]
	Publishers map[string]*PublisherGroup[T]
}

func NewGoEventBus[T Event]() *GoEventBus[T] {
	return &GoEventBus[T]{
		Publishers: make(map[string]*PublisherGroup[T]),
	}
}

func (geb *GoEventBus[T]) GetPublisher(group, name string) Publisher[T] {
	return geb.GetPublisherByConfig(group, name, nil)
}

func (geb *GoEventBus[T]) GetPublisherByConfig(group, name string, config *PublisherConfig) Publisher[T] {
	if len(group) == 0 || len(name) == 0 {
		return nil
	}
	if _, ok := geb.Publishers[group]; !ok {
		geb.Publishers[group] = NewPublisherGroup[T](name, config)
	}
	return geb.Publishers[group].GetPublisher(name)
}
