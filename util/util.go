// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package util

import (
	"strings"
)

const (
	qualifiedNameDelimiter = ":"
	aliasDelimiter         = "/"
)

func ParseQualifiedName(name string) (source string, plugin string) {
	parsed := strings.Split(name, qualifiedNameDelimiter)

	return parsed[0], parsed[1]
}

func ParseAlias(alias string) (organization string, repository string) {
	parsed := strings.Split(alias, aliasDelimiter)

	return parsed[0], parsed[1]
}
