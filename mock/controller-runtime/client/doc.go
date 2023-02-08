// Copyright 2021 OnMetal authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package client contains mocks for controller-runtime's client package.
//
//go:generate $MOCKGEN -copyright_file ../../../hack/boilerplate.go.txt -package client -destination mocks.go sigs.k8s.io/controller-runtime/pkg/client Client,FieldIndexer
//go:generate $MOCKGEN -copyright_file ../../../hack/boilerplate.go.txt -package client -destination funcs.go github.com/onmetal/controller-utils/mock/controller-runtime/client IndexerFunc
package client

import "sigs.k8s.io/controller-runtime/pkg/client"

type IndexerFunc interface {
	Call(object client.Object) []string
}
