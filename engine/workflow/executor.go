package workflow

type Executor interface {
	Execute(Workflow) error
}
