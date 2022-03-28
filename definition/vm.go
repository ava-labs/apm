package definition

import "github.com/ava-labs/avalanchego/version"

type VM interface {
	Plugin

	Version() version.Version
	Repository() string
	SHA256() string
	Commit() string
	Branch() string
}
