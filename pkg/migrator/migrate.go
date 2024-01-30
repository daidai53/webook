// Copyright@daidai53 2024
package migrator

type Entity interface {
	ID() int64
	CompareTo(dst Entity) bool
}
