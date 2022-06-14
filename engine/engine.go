// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package engine

import "github.com/ava-labs/apm/workflow"

var _ workflow.Executor = &WorkflowEngine{}

func NewWorkflowEngine() *WorkflowEngine {
	return &WorkflowEngine{}
}

type WorkflowEngine struct {
}

func (w WorkflowEngine) Execute(workflow workflow.Workflow) error {
	return workflow.Execute()
}
