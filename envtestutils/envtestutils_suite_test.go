// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package envtestutils

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestEnvtestutils(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Envtestutils Suite")
}
