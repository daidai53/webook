// Copyright@daidai53 2024
package events

import "context"

type Producer interface {
	ProduceIncosistentEvent(ctx context.Context, evt InconsistentEvent) error
}
