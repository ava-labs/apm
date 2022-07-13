// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package workflow

import (
	"os"
	"os/exec"

	"github.com/spf13/afero"

	"github.com/ava-labs/apm/url"
)

type Installer interface {
	Download(url string, path string) error
	Decompress(source string, dest string) error
	// Install installs the VM. installScriptPath is a path relative to
	// workingDir.
	Install(workingDir string, args ...string) error
}

var _ Installer = &VMInstaller{}

type VMInstallerConfig struct {
	Fs        afero.Fs
	URLClient url.Client
}

func NewVMInstaller(config VMInstallerConfig) *VMInstaller {
	return &VMInstaller{
		fs:     config.Fs,
		Client: config.URLClient,
	}
}

type VMInstaller struct {
	fs afero.Fs
	url.Client
}

func (t VMInstaller) Decompress(source string, dest string) error {
	cmd := exec.Command("tar", "xf", source, "-C", dest, "--strip-components", "1")
	return cmd.Run()
}

func (t VMInstaller) Install(workingDir string, args ...string) error {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = workingDir

	return cmd.Run()
}
