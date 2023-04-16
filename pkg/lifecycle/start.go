package lifecycle

import "context"

type Start interface {
	Start(ctx context.Context) error
}
