// Copyright@daidai53 2024
package events

type InconsistentEvent struct {
	ID        int64
	Direction string
	Type      string
}

const (
	InconsistentEventTypeTargetMissing = "target_missing"
	InconsistentEventTypeNEQ           = "neq"
	InconsistentEventTypeBaseMissing   = "base_missing"
)
