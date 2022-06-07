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
	Install(workingDir string, installScriptPath string) error
}

var _ Installer = &VMInstaller{}

type VMInstallerConfig struct {
	Fs        afero.Fs
	UrlClient url.Client
}

func NewVMInstaller(config VMInstallerConfig) *VMInstaller {
	return &VMInstaller{
		fs:     config.Fs,
		Client: config.UrlClient,
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

func (t VMInstaller) Install(workingDir string, installScriptRelativePath string) error {
	cmd := exec.Command(installScriptRelativePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = workingDir

	return cmd.Run()
}
