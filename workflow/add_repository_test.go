package workflow

import (
	"fmt"
	"testing"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/ava-labs/apm/storage"
)

func TestAddRepositoryExecute(t *testing.T) {
	errWrong := fmt.Errorf("something went wrong")

	type mocks struct {
		sourcesList *storage.MockStorage[storage.SourceInfo]
	}
	tests := []struct {
		name    string
		setup   func(mocks)
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "can't read from sources list",
			setup: func(mocks mocks) {
				mocks.sourcesList.EXPECT().Has([]byte("alias")).Return(false, errWrong)
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Equal(t, errWrong, err)
			},
		},
		{
			name: "duplicate alias",
			setup: func(mocks mocks) {
				mocks.sourcesList.EXPECT().Has([]byte("alias")).Return(true, nil)
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Error(t, err)
			},
		},
		{
			name: "adding to sources list fails",
			setup: func(mocks mocks) {
				mocks.sourcesList.EXPECT().Has([]byte("alias")).Return(false, nil)
				mocks.sourcesList.EXPECT().
					Put(
						[]byte("alias"),
						storage.SourceInfo{
							Alias:  "alias",
							URL:    "url",
							Commit: plumbing.ZeroHash,
						},
					).
					Return(errWrong)
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Equal(t, errWrong, err)
			},
		},
		{
			name: "success",
			setup: func(mocks mocks) {
				mocks.sourcesList.EXPECT().Has([]byte("alias")).Return(false, nil)
				mocks.sourcesList.EXPECT().
					Put(
						[]byte("alias"),
						storage.SourceInfo{
							Alias:  "alias",
							URL:    "url",
							Commit: plumbing.ZeroHash,
						},
					).
					Return(nil)
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Nil(t, err)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			var sourcesList *storage.MockStorage[storage.SourceInfo]

			sourcesList = storage.NewMockStorage[storage.SourceInfo](ctrl)

			test.setup(mocks{
				sourcesList: sourceList,
			})

			wf := NewAddRepository(
				AddRepositoryConfig{
					SourcesList: sourcesList,
					Alias:      "alias",
					Url:        "url",
				},
			)

			test.wantErr(t, wf.Execute())
		})
	}
}
