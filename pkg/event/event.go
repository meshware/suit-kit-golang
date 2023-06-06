package event

type Event interface {
	GetSource() interface{}
	GetTarget() interface{}
}

var _ Event = &AbstractEvent{}

type AbstractEvent struct {
	Source interface{}
	Target interface{}
}

func (ae AbstractEvent) GetSource() interface{} {
	return ae.Source
}

func (ae AbstractEvent) GetTarget() interface{} {
	return ae.Target
}
