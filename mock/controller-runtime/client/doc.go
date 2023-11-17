// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

// Package client contains mocks for controller-runtime's client package.
//
//go:generate $MOCKGEN -copyright_file ../../../hack/boilerplate.go.txt -package client -destination mocks.go sigs.k8s.io/controller-runtime/pkg/client Client,FieldIndexer
//go:generate $MOCKGEN -copyright_file ../../../hack/boilerplate.go.txt -package client -destination funcs.go github.com/ironcore-dev/controller-utils/mock/controller-runtime/client IndexerFunc
package client

import "sigs.k8s.io/controller-runtime/pkg/client"

type IndexerFunc interface {
	Call(object client.Object) []string
}
