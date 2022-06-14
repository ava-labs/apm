// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package types

type Definition interface {
	GetID() string
	GetAlias() string
	GetHomepage() string
	GetDescription() string
	GetMaintainers() []string
}
