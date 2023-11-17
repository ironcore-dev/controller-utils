// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package testutils

import (
	"github.com/ironcore-dev/controller-utils/testutils/matchers"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/utils/semantic"
)

// EqualWithEquality returns a matcher that determines whether the expected value is equal to an actual
// value using the supplied semantic.Equalities.
func EqualWithEquality(equalities semantic.Equalities, expected interface{}) *matchers.EqualitiesEqualMatcher {
	return &matchers.EqualitiesEqualMatcher{
		Equalities: equalities,
		Expected:   expected,
	}
}

// SemanticEqual returns a matcher that determines whether the expected value is equal to an actual value
// using equality.Semantic.Equalities.
func SemanticEqual(expected interface{}) *matchers.EqualitiesEqualMatcher {
	return EqualWithEquality(semantic.Equalities(equality.Semantic.Equalities), expected)
}

// DerivativeWithEquality returns a matcher that determines whether the actual value derives from the expected
// value using the supplied semantic.Equalities.
func DerivativeWithEquality(equalities semantic.Equalities, expected interface{}) *matchers.EqualitiesDerivativeMatcher {
	return &matchers.EqualitiesDerivativeMatcher{
		Equalities: equalities,
		Expected:   expected,
	}
}

// SemanticDerivative returns a matcher that determines whether the actual value derives from the expected
// value using equality.Semantic.Equalities.
func SemanticDerivative(expected interface{}) *matchers.EqualitiesDerivativeMatcher {
	return DerivativeWithEquality(semantic.Equalities(equality.Semantic.Equalities), expected)
}

// MatchErrorFunc returns a matcher that determines whether the actual value is an error and matches the supplied
// function. The name of the function will be dynamically inferred.
func MatchErrorFunc(f func(err error) bool) *matchers.ErrorFuncMatcher {
	return &matchers.ErrorFuncMatcher{
		Func: f,
	}
}

// MatchNamedErrorFunc returns a matcher that determines whether the actual value is an error and matches the supplied
// function. The given name will be used unless it's empty in which case it will be dynamically inferred.
func MatchNamedErrorFunc(name string, f func(err error) bool) matchers.ErrorFuncMatcher {
	return matchers.ErrorFuncMatcher{
		Name: name,
		Func: f,
	}
}
