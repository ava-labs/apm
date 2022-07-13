// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package checksum

import (
	"crypto/sha256"
	"hash"
	"io"

	"github.com/spf13/afero"
)

type Checksummer interface {
	Checksum(path string) []byte
}

var _ Checksummer = &SHA256{}

func NewSHA256(fs afero.Fs) *SHA256 {
	return &SHA256{
		h:  sha256.New(),
		Fs: fs,
	}
}

type SHA256 struct {
	h  hash.Hash
	Fs afero.Fs
}

func (s SHA256) Checksum(path string) []byte {
	defer s.h.Reset()

	f, err := s.Fs.Open(path)
	if err != nil {
		return nil
	}

	defer f.Close()
	if _, err := io.Copy(s.h, f); err != nil {
		return nil
	}

	return s.h.Sum(nil)
}
