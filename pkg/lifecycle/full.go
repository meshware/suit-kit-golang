package lifecycle

type Full interface {
	Init
	Start
	Stop
}
