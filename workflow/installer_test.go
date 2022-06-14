// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package workflow

import (
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"

	"github.com/ava-labs/apm/url"
)

func TestTmpInstaller_Download(t *testing.T) {
	dummyErr := fmt.Errorf("something went wrong")

	type mocks struct {
		fs     afero.Fs
		client *url.MockClient
	}
	type args struct {
		url  string
		path string
	}
	var tests = []struct {
		name    string
		setup   func(mocks)
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "failure",
			setup: func(mocks mocks) {
				mocks.client.EXPECT().Download("tmp/file.tar.gz", "www.url.com/binary.tar.gz").Return(dummyErr)
			},
			args: args{
				url:  "www.url.com/binary.tar.gz",
				path: "tmp/file.tar.gz",
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Equal(t, dummyErr, err)
			},
		},
		{
			name: "success",
			setup: func(mocks mocks) {
				mocks.client.EXPECT().Download("tmp/file.tar.gz", "www.url.com/binary.tar.gz").Return(nil)
			},
			args: args{
				url:  "www.url.com/binary.tar.gz",
				path: "tmp/file.tar.gz",
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Nil(t, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t1 *testing.T) {
			ctrl := gomock.NewController(t)
			fs := afero.NewMemMapFs()
			client := url.NewMockClient(ctrl)

			tt.setup(mocks{
				fs:     fs,
				client: client,
			})

			installer := NewVMInstaller(VMInstallerConfig{
				Fs:        fs,
				UrlClient: client,
			})

			tt.wantErr(t1, installer.Download(tt.args.url, tt.args.path), fmt.Sprintf("Download(%v, %v)", tt.args.url, tt.args.path))
		})
	}
}
