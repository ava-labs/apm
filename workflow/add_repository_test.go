// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package workflow

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ava-labs/apm/state"
)

func TestAddRepositoryExecute(t *testing.T) {
	type mocks struct {
		sourcesList map[string]*state.SourceInfo
	}
	tests := []struct {
		name    string
		setup   func(mocks)
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "already exists",
			setup: func(mocks mocks) {
				mocks.sourcesList["alias"] = nil
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Error(t, err)
			},
		},
		{
			name: "success",
			setup: func(mocks mocks) {
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Nil(t, err)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sourcesList := make(map[string]*state.SourceInfo)

			test.setup(mocks{
				sourcesList: sourcesList,
			})

			wf := NewAddRepository(
				AddRepositoryConfig{
					SourcesList: sourcesList,
					Alias:       "alias",
					URL:         "url",
					Branch:      "master",
				},
			)

			test.wantErr(t, wf.Execute())
		})
	}
}
