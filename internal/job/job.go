// Copyright@daidai53 2024
package job

type Job interface {
	Name() string
	Run() error
}
