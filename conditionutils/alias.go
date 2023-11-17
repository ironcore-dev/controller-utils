// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package conditionutils

var (
	// DefaultTransition is the default Transition.
	DefaultTransition Transition = &FieldsTransition{IncludeStatus: true}

	// DefaultAccessor is an Accessor initialized with the default fields.
	// See NewAccessor for more.
	DefaultAccessor = NewAccessor(AccessorOptions{})

	// Update updates the condition with the given options.
	// See Accessor.Update for more.
	Update = DefaultAccessor.Update

	// MustUpdate updates the condition with the given options.
	// See Accessor.MustUpdate for more.
	MustUpdate = DefaultAccessor.MustUpdate

	// UpdateSlice updates the slice with the given options.
	// See Accessor.UpdateSlice for more.
	UpdateSlice = DefaultAccessor.UpdateSlice

	// MustUpdateSlice updates the slice with the given options.
	// See Accessor.MustUpdateSlice for more.
	MustUpdateSlice = DefaultAccessor.MustUpdateSlice

	// FindSliceIndex finds the index of the target condition in the given slice.
	// See Accessor.FindSliceIndex for more.
	FindSliceIndex = DefaultAccessor.FindSliceIndex

	// MustFindSliceIndex finds the index of the target condition in the given slice.
	// See Accessor.MustFindSliceIndex for more.
	MustFindSliceIndex = DefaultAccessor.MustFindSliceIndex

	// FindSlice finds the target condition in the given slice.
	// See Accessor.FindSlice for more.
	FindSlice = DefaultAccessor.FindSlice

	// MustFindSlice finds the target condition in the given slice.
	// See Accessor.MustFindSlice for more.
	MustFindSlice = DefaultAccessor.MustFindSlice

	// FindSliceStatus finds the condition status in the given slice.
	// See Accessor.FindSliceStatus for more.
	FindSliceStatus = DefaultAccessor.FindSliceStatus

	// MustFindSliceStatus finds the condition status in the given slice.
	// See Accessor.MustFindSliceStatus for more.
	MustFindSliceStatus = DefaultAccessor.MustFindSliceStatus
)
