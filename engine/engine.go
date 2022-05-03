package engine

var _ Engine = &WorkflowEngine{}

type Engine interface {
	Execute(workflow Workflow) error
}

func NewWorkflowEngine() *WorkflowEngine {
	return &WorkflowEngine{}
}

type WorkflowEngine struct {
}

func (w WorkflowEngine) Execute(workflow Workflow) error {
	return workflow.Execute()
}
