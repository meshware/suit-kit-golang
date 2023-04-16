package lifecycle

type Stop interface {
	Stop() error
}
