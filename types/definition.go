// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package types

type Definition interface {
	ID() string
	Alias() string
	Homepage() string
	Description() string
	Maintainers() []string
}
