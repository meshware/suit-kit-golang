package lifecycle

import "context"

type Stop interface {
	Stop(ctx context.Context) error
}
