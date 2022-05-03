package engine

type Workflow interface {
	Execute() error
}
