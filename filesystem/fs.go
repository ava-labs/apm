package filesystem

import "os"

var _ FileSystem = &fs{}

type FileSystem interface {
	Stat(string) (os.FileInfo, error)
	Mkdir(string, os.FileMode) error
	Remove(string) error
	RemoveAll(string) error
	Rename(string, string) error
}

type fs struct {
}

func (f fs) Stat(path string) (os.FileInfo, error) {
	return os.Stat(path)
}

func (f fs) Mkdir(path string, perm os.FileMode) error {
	return os.Mkdir(path, perm)
}

func (f fs) Remove(path string) error {
	return os.Remove(path)
}

func (f fs) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

func (f fs) Rename(old string, new string) error {
	return os.Rename(old, new)
}
