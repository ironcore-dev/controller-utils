// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

//go:generate $MOCKGEN -copyright_file ../../../hack/boilerplate.go.txt -package metautils -destination=funcs.go github.com/ironcore-dev/controller-utils/mock/controller-utils/metautils EachListItemFunc
package metautils

import "sigs.k8s.io/controller-runtime/pkg/client"

type EachListItemFunc interface {
	Call(obj client.Object) error
}
