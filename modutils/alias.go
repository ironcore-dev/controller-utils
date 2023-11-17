// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package modutils

var (
	// DefaultExecutor is the default executor.
	DefaultExecutor = &Executor{}

	// ListE is an alias to DefaultExecutor.ListE.
	ListE = DefaultExecutor.ListE
	// List is an alias to DefaultExecutor.List.
	List = DefaultExecutor.List

	// GetE is an alias to DefaultExecutor.GetE.
	GetE = DefaultExecutor.GetE
	// Get is an alias to DefaultExecutor.Get.
	Get = DefaultExecutor.Get

	// DirE is an alias to DefaultExecutor.DirE.
	DirE = DefaultExecutor.DirE
	// Dir is an alias to DefaultExecutor.Dir.
	Dir = DefaultExecutor.Dir

	// BuildE is an alias to DefaultExecutor.BuildE.
	BuildE = DefaultExecutor.BuildE
	// Build is an alias to DefaultExecutor.Build.
	Build = DefaultExecutor.Build
)
