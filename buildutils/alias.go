// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package buildutils

var (
	// DefaultBuilder is the default Builder.
	DefaultBuilder = Builder{}

	// Build is an alias for DefaultBuilder.Build.
	Build = DefaultBuilder.Build
)
