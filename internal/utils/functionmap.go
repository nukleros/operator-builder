// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package utils

import (
	"strings"
	"text/template"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

const (
	quoteStringFuncName    = "quoteString"
	removeStringFuncName   = "removeString"
	containsStringFuncName = "containsString"
)

// QuoteStringHelper returns the function map for quoting strings in templates.
func QuoteStringHelper() template.FuncMap {
	funcMap := machinery.DefaultFuncMap()
	funcMap[quoteStringFuncName] = func(value string) string {
		if string(value[0]) != `"` {
			value = `"` + value
		}

		if string(value[len(value)-1]) != `"` {
			value += `"`
		}

		return value
	}

	return funcMap
}

// RemoveStringHelper returns the function map for quoting strings in templates.
func RemoveStringHelper() template.FuncMap {
	funcMap := machinery.DefaultFuncMap()
	funcMap[removeStringFuncName] = func(value, with string) string {
		return strings.ReplaceAll(with, value, "")
	}

	return funcMap
}

// ContainsStringHelper returns the function map for seeing if strings are
// contained within other strings.
func ContainsStringHelper() template.FuncMap {
	funcMap := machinery.DefaultFuncMap()
	funcMap[containsStringFuncName] = func(value, in string) bool {
		return strings.Contains(in, value)
	}

	return funcMap
}

// TemplateHelpers returns all of the template helpers in the utils package.
func TemplateHelpers() template.FuncMap {
	funcMap := machinery.DefaultFuncMap()
	funcMap[quoteStringFuncName] = QuoteStringHelper()[quoteStringFuncName]
	funcMap[removeStringFuncName] = RemoveStringHelper()[removeStringFuncName]
	funcMap[containsStringFuncName] = ContainsStringHelper()[containsStringFuncName]

	return funcMap
}
