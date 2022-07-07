// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package util

import (
	"strings"

	"github.com/ava-labs/apm/constant"
)

func ParseQualifiedName(name string) (source string, plugin string) {
	parsed := strings.Split(name, constant.QualifiedNameDelimiter)

	return parsed[0], parsed[1]
}

func ParseAlias(alias string) (organization string, repository string) {
	parsed := strings.Split(alias, constant.AliasDelimiter)

	return parsed[0], parsed[1]
}

func ValidAlias(alias string) bool {
	if organization, repository := ParseAlias(alias); organization == "" || repository == "" {
		return false
	}

	return true
}
