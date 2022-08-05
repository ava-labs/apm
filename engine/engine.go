// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package engine

import (
	"github.com/ava-labs/apm/storage"
	"github.com/ava-labs/apm/workflow"
)

var _ workflow.Executor = &WorkflowEngine{}

func NewWorkflowEngine(stateFile storage.StateFile) *WorkflowEngine {
	return &WorkflowEngine{
		stateFile: stateFile,
	}
}

type WorkflowEngine struct {
	stateFile storage.StateFile
}

func (w *WorkflowEngine) Execute(workflow workflow.Workflow) error {
	defer w.stateFile.Commit()
	return workflow.Execute()
}
