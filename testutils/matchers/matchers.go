// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package matchers

import (
	"fmt"
	"reflect"
	"runtime"

	"github.com/onsi/gomega/format"
	"k8s.io/utils/semantic"
)

// EqualitiesEqualMatcher is a matcher that matches the Expected value using the given Equalities
// and semantic.Equalities.DeepEqual.
type EqualitiesEqualMatcher struct {
	Equalities semantic.Equalities
	Expected   interface{}
}

func (m *EqualitiesEqualMatcher) FailureMessage(actual interface{}) (message string) {
	actualString, actualOK := actual.(string)
	expectedString, expectedOK := m.Expected.(string)
	if actualOK && expectedOK {
		return format.MessageWithDiff(actualString, "to equal", expectedString)
	}

	return format.Message(actual, "to equal with equality", m.Expected)
}

func (m *EqualitiesEqualMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "not to equal with equality", m.Expected)
}

func (m *EqualitiesEqualMatcher) Match(actual interface{}) (bool, error) {
	if m.Equalities == nil {
		return false, fmt.Errorf("must set Equalities")
	}

	if actual == nil && m.Expected == nil {
		return false, fmt.Errorf("refusing to compare <nil> to <nil>, BeNil() should be used instead")
	}

	return m.Equalities.DeepEqual(actual, m.Expected), nil
}

// EqualitiesDerivativeMatcher is a matcher that matches the Expected value using the given Equalities
// and semantic.Equalities.DeepDerivative.
type EqualitiesDerivativeMatcher struct {
	Equalities semantic.Equalities
	Expected   interface{}
}

func (m *EqualitiesDerivativeMatcher) FailureMessage(actual interface{}) (message string) {
	actualString, actualOK := actual.(string)
	expectedString, expectedOK := m.Expected.(string)
	if actualOK && expectedOK {
		return format.MessageWithDiff(actualString, "to derive", expectedString)
	}

	return format.Message(actual, "to derive with equality", m.Expected)
}

func (m *EqualitiesDerivativeMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "not to derive with equality", m.Expected)
}

func (m *EqualitiesDerivativeMatcher) Match(actual interface{}) (bool, error) {
	if m.Equalities == nil {
		return false, fmt.Errorf("must set Equalities")
	}

	if actual == nil && m.Expected == nil {
		return false, fmt.Errorf("refusing to compare <nil> to <nil>, BeNil() should be used instead")
	}

	return m.Equalities.DeepDerivative(actual, m.Expected), nil
}

type ErrorFuncMatcher struct {
	Name string
	Func func(err error) bool
}

func (m *ErrorFuncMatcher) Match(actual interface{}) (success bool, err error) {
	if m.Func == nil {
		return false, fmt.Errorf("must set Func")
	}

	actualErr, ok := actual.(error)
	if !ok {
		return false, fmt.Errorf("expected an error-type but got %s", format.Object(actual, 0))
	}

	return m.Func(actualErr), nil
}

func (m *ErrorFuncMatcher) nameOrFuncName() string {
	if m.Name != "" {
		return m.Name
	}

	return runtime.FuncForPC(reflect.ValueOf(m.Func).Pointer()).Name()
}

func (m *ErrorFuncMatcher) FailureMessage(actual interface{}) (message string) {
	name := m.nameOrFuncName()
	return fmt.Sprintf("expected an error matching %s to have occurred but got %s", name, format.Object(actual, 0))
}

func (m *ErrorFuncMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	name := m.nameOrFuncName()
	return fmt.Sprintf("expected an error not matching %s to have occurred but got %s", name, format.Object(actual, 0))
}
