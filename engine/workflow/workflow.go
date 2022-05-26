package workflow

type Workflow interface {
	Execute() error
}
