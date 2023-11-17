// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

// Package clientutils contains mocks for the actual clientutils package.
//
//go:generate $MOCKGEN -copyright_file ../../../hack/boilerplate.go.txt -package clientutils -destination=mocks.go github.com/ironcore-dev/controller-utils/clientutils PatchProvider
package clientutils
