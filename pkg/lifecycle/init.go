package lifecycle

import "context"

type Init interface {
	Init(ctx context.Context) error
}
