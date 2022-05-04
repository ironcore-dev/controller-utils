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

// Package conditionutils simplifies condition handling with any structurally compatible condition
// (comparable to a sort of duck-typing) via go reflection.
package conditionutils

import (
	"fmt"
	"reflect"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/conversion"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/utils/clock"
)

const (
	// DefaultTypeField is the default name for a condition's type field.
	DefaultTypeField = "Type"
	// DefaultStatusField is the default name for a condition's status field.
	DefaultStatusField = "Status"
	// DefaultLastUpdateTimeField field is the default name for a condition's last update time field.
	DefaultLastUpdateTimeField = "LastUpdateTime"
	// DefaultLastTransitionTimeField field is the default name for a condition's last transition time field.
	DefaultLastTransitionTimeField = "LastTransitionTime"
	// DefaultReasonField field is the default name for a condition's reason field.
	DefaultReasonField = "Reason"
	// DefaultMessageField field is the default name for a condition's message field.
	DefaultMessageField = "Message"
	// DefaultObservedGenerationField field is the default name for a condition's observed generation field.
	DefaultObservedGenerationField = "ObservedGeneration"
)

func enforceStruct(cond interface{}) (reflect.Value, error) {
	v := reflect.ValueOf(cond)
	if v.Kind() != reflect.Struct {
		return reflect.Value{}, fmt.Errorf("type %T is not a struct", cond)
	}
	return v, nil
}

func enforcePtrToStruct(cond interface{}) (reflect.Value, error) {
	v := reflect.ValueOf(cond)
	if v.Kind() != reflect.Ptr {
		return reflect.Value{}, fmt.Errorf("type %T is not a pointer to a struct", cond)
	}
	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return reflect.Value{}, fmt.Errorf("type %T is not a pointer to a struct", cond)
	}
	return v, nil
}

func enforceStructSlice(condSlice interface{}) (sliceV reflect.Value, structType reflect.Type, err error) {
	v := reflect.ValueOf(condSlice)
	if v.Kind() != reflect.Slice {
		return reflect.Value{}, nil, fmt.Errorf("type %T is not a slice of structs", condSlice)
	}

	structType = v.Type().Elem()
	if structType.Kind() != reflect.Struct {
		return reflect.Value{}, nil, fmt.Errorf("type %T is not a slice of structs", condSlice)
	}

	return v, structType, nil
}

func enforcePtrToStructSlice(condSlicePtr interface{}) (sliceV reflect.Value, structType reflect.Type, err error) {
	v := reflect.ValueOf(condSlicePtr)
	if v.Kind() != reflect.Ptr {
		return reflect.Value{}, nil, fmt.Errorf("type %T is not a pointer to a slice of structs", condSlicePtr)
	}

	v = v.Elem()

	if v.Kind() != reflect.Slice {
		return reflect.Value{}, nil, fmt.Errorf("type %T is not a pointer to a slice of structs", condSlicePtr)
	}

	structType = v.Type().Elem()
	if structType.Kind() != reflect.Struct {
		return reflect.Value{}, nil, fmt.Errorf("type %T is not a pointer to a slice of structs", condSlicePtr)
	}

	return v, structType, nil
}

func getAndConvertField(v reflect.Value, name string, into interface{}) error {
	f := v.FieldByName(name)
	if !v.IsValid() {
		return fmt.Errorf("type %T has no field %q", v.Interface(), name)
	}

	intoV, err := conversion.EnforcePtr(into)
	if err != nil {
		return err
	}

	fType := f.Type()
	if fType.Kind() == reflect.Ptr {
		fType = fType.Elem()
	}

	if !fType.ConvertibleTo(intoV.Type()) {
		return fmt.Errorf("type %T field %q type %s cannot be converted into %T", v.Interface(), fType, name, into)
	}
	intoV.Set(reflect.Indirect(f).Convert(intoV.Type()))
	return nil
}

// direct is the inverse to reflect.Indirect.
//
// It takes in a value and returns nil if it is a zero value.
// Otherwise, it returns a pointer to that value.
func direct(v reflect.Value) reflect.Value {
	if v.IsZero() {
		return reflect.New(reflect.PtrTo(v.Type())).Elem()
	}

	res := reflect.New(v.Type())
	res.Elem().Set(v)
	return res
}

// setFieldConverted sets the specified field to the given value, potentially converting it before.
func setFieldConverted(v reflect.Value, name string, newValue interface{}) error {
	f := v.FieldByName(name)
	if f == (reflect.Value{}) {
		return fmt.Errorf("type %T has no field %q", v.Interface(), name)
	}

	fType := f.Type()
	var isPtr bool
	if fType.Kind() == reflect.Ptr {
		isPtr = true
		fType = fType.Elem()
	}

	newV := reflect.ValueOf(newValue)
	if !newV.CanConvert(fType) {
		return fmt.Errorf("value %T cannot be converted into type %s of field %q of type %T", newValue, fType, name, v.Interface())
	}

	newV = newV.Convert(fType)
	if isPtr {
		newV = direct(newV)
	}

	f.Set(newV)
	return nil
}

func valueHasField(v reflect.Value, name string) bool {
	return v.FieldByName(name) != (reflect.Value{})
}

// Accessor allows getting and setting fields from conditions as well as to check on their presence.
// In addition, it allows complex manipulations on individual conditions and condition slices.
type Accessor struct {
	typeField               string
	statusField             string
	lastUpdateTimeField     string
	lastTransitionTimeField string
	reasonField             string
	messageField            string
	observedGenerationField string

	disableTimestampUpdates bool
	transition              Transition
	clock                   clock.Clock
}

// Transition can determine whether a condition transitioned (i.e. LastTransitionTime needs to be updated) or not.
type Transition interface {
	// Checkpoint creates a TransactionCheckpoint using the current values of cond extracted with Accessor.
	Checkpoint(acc *Accessor, cond interface{}) (TransitionCheckpoint, error)
}

// TransitionCheckpoint can determine whether a condition transitioned using pre-gathered values of a condition.
type TransitionCheckpoint interface {
	// Transitioned reports whether the condition transitioned.
	Transitioned(acc *Accessor, cond interface{}) (bool, error)
}

// FieldsTransition computes whether a condition transitioned using the `Include`-Fields.
type FieldsTransition struct {
	// IncludeStatus includes Accessor.Status for the transition calculation. This is the most frequent choice.
	IncludeStatus bool
	// IncludeReason includes Accessor.Reason for the transition calculation. While more seldom, there are use cases
	// for including the reason in the transition calculation.
	IncludeReason bool
	// IncludeMessage includes Accessor.Message for the transition calculation. Used rarely, usually causes
	// a lot of transitions.
	IncludeMessage bool
}

func (f *FieldsTransition) computeValues(acc *Accessor, cond interface{}) (*fieldsTransitionValues, error) {
	var fields fieldsTransitionValues

	if f.IncludeStatus {
		status, err := acc.Status(cond)
		if err != nil {
			return nil, err
		}

		fields.Status = status
	}

	if f.IncludeReason {
		reason, err := acc.Reason(cond)
		if err != nil {
			return nil, err
		}

		fields.Reason = reason
	}

	if f.IncludeMessage {
		message, err := acc.Message(cond)
		if err != nil {
			return nil, err
		}

		fields.Message = message
	}

	return &fields, nil
}

// Checkpoint implements Transition.
func (f *FieldsTransition) Checkpoint(acc *Accessor, cond interface{}) (TransitionCheckpoint, error) {
	values, err := f.computeValues(acc, cond)
	if err != nil {
		return nil, err
	}

	return &fieldsTransitionCheckpoint{
		transition: *f,
		values:     *values,
	}, nil
}

type fieldsTransitionValues struct {
	Status  corev1.ConditionStatus
	Reason  string
	Message string
}

type fieldsTransitionCheckpoint struct {
	transition FieldsTransition
	values     fieldsTransitionValues
}

// Transitioned implements TransitionCheckpoint.
func (f *fieldsTransitionCheckpoint) Transitioned(acc *Accessor, cond interface{}) (bool, error) {
	newValues, err := f.transition.computeValues(acc, cond)
	if err != nil {
		return false, err
	}

	return newValues.Status != f.values.Status ||
		newValues.Reason != f.values.Reason ||
		newValues.Message != f.values.Message, nil
}

// Type extracts the type of the given condition.
//
// It errors if the given value is not a struct or does not have a field
// that can be converted to the output format.
func (a *Accessor) Type(cond interface{}) (string, error) {
	v, err := enforceStruct(cond)
	if err != nil {
		return "", err
	}

	var typeValue string
	if err := getAndConvertField(v, a.typeField, &typeValue); err != nil {
		return "", err
	}
	return typeValue, nil
}

// MustType extracts the type of the given condition.
//
// It panics if the given value is not a struct or does not have a field
// that can be converted to the output format.
func (a *Accessor) MustType(cond interface{}) string {
	typ, err := a.Type(cond)
	utilruntime.Must(err)
	return typ
}

// SetType sets the type of the given condition to the given value.
//
// It errors if the given value is not a pointer to a struct or does not have a field
// that can be converted to the given format.
func (a *Accessor) SetType(condPtr interface{}, typ string) error {
	v, err := enforcePtrToStruct(condPtr)
	if err != nil {
		return err
	}

	return setFieldConverted(v, a.typeField, typ)
}

// MustSetType sets the type of the given condition to the given value.
//
// It panics if the given value is not a pointer to a struct or does not have a field
// that can be converted to the given format.
func (a *Accessor) MustSetType(condPtr interface{}, typ string) {
	utilruntime.Must(a.SetType(condPtr, typ))
}

// Status extracts the status of the given condition.
//
// It errors if the given value is not a struct or does not have a field
// that can be converted to the output format.
func (a *Accessor) Status(cond interface{}) (corev1.ConditionStatus, error) {
	v, err := enforceStruct(cond)
	if err != nil {
		return "", err
	}

	var status corev1.ConditionStatus
	if err := getAndConvertField(v, a.statusField, &status); err != nil {
		return "", err
	}
	return status, nil
}

// MustStatus extracts the status of the given condition.
//
// It panics if the given value is not a struct or does not have a field
// that can be converted to the output format.
func (a *Accessor) MustStatus(cond interface{}) corev1.ConditionStatus {
	status, err := a.Status(cond)
	utilruntime.Must(err)
	return status
}

// SetStatus sets the status of the given condition.
//
// It errors if the given value is not a pointer to a struct or does not have a field
// that can be converted to the given format.
func (a *Accessor) SetStatus(condPtr interface{}, status corev1.ConditionStatus) error {
	v, err := enforcePtrToStruct(condPtr)
	if err != nil {
		return err
	}

	return setFieldConverted(v, a.statusField, status)
}

// MustSetStatus sets the status of the given condition.
//
// It panics if the given value is not a pointer to a struct or does not have a field
// that can be converted to the given format.
func (a *Accessor) MustSetStatus(condPtr interface{}, status corev1.ConditionStatus) {
	utilruntime.Must(a.SetStatus(condPtr, status))
}

// HasLastUpdateTime checks if the given condition has a 'LastUpdateTime' field.
//
// It errors if the given value is not a struct.
func (a *Accessor) HasLastUpdateTime(cond interface{}) (bool, error) {
	v, err := enforceStruct(cond)
	if err != nil {
		return false, err
	}

	return valueHasField(v, a.lastUpdateTimeField), nil
}

// MustHasLastUpdateTime checks if the given condition has a 'LastUpdateTime' field.
//
// It panics if the given value is not a struct.
func (a *Accessor) MustHasLastUpdateTime(cond interface{}) bool {
	ok, err := a.HasLastUpdateTime(cond)
	utilruntime.Must(err)
	return ok
}

// LastUpdateTime extracts the last update time of the given condition.
//
// It errors if the given value is not a struct or does not have a field
// that can be converted to the output format.
func (a *Accessor) LastUpdateTime(cond interface{}) (metav1.Time, error) {
	v, err := enforceStruct(cond)
	if err != nil {
		return metav1.Time{}, err
	}

	var lastUpdateTime metav1.Time
	if err := getAndConvertField(v, a.lastUpdateTimeField, &lastUpdateTime); err != nil {
		return metav1.Time{}, err
	}
	return lastUpdateTime, nil
}

// MustLastUpdateTime extracts the last update time of the given condition.
//
// It panics if the given value is not a struct or does not have a field
// that can be converted to the output format.
func (a *Accessor) MustLastUpdateTime(cond interface{}) metav1.Time {
	t, err := a.LastUpdateTime(cond)
	utilruntime.Must(err)
	return t
}

// SetLastUpdateTime sets the last update time of the given condition.
//
// It errors if the given value is not a pointer to a struct or does not have a field
// that can be converted to the given format.
func (a *Accessor) SetLastUpdateTime(condPtr interface{}, lastUpdateTime metav1.Time) error {
	v, err := enforcePtrToStruct(condPtr)
	if err != nil {
		return err
	}

	return setFieldConverted(v, a.lastUpdateTimeField, lastUpdateTime)
}

// MustSetLastUpdateTime sets the last update time of the given condition.
//
// It errors if the given value is not a pointer to a struct or does not have a field
// that can be converted to the given format.
func (a *Accessor) MustSetLastUpdateTime(condPtr interface{}, lastUpdateTime metav1.Time) {
	utilruntime.Must(a.SetLastUpdateTime(condPtr, lastUpdateTime))
}

// SetLastUpdateTimeIfExists sets the last update time of the given condition if the field exists.
//
// It errors if the given value is not a pointer to a struct or the field value cannot be converted
// to the given format.
func (a *Accessor) SetLastUpdateTimeIfExists(condPtr interface{}, lastUpdateTime metav1.Time) error {
	condV, err := conversion.EnforcePtr(condPtr)
	if err != nil {
		return err
	}

	cond := condV.Interface()
	if ok, err := a.HasLastUpdateTime(cond); err != nil || !ok {
		return err
	}

	return a.SetLastUpdateTime(condPtr, lastUpdateTime)
}

// MustSetLastUpdateTimeIfExists sets the last update time of the given condition if the field exists.
//
// It panics if the given value is not a pointer to a struct or the field value cannot be converted
// to the given format.
func (a *Accessor) MustSetLastUpdateTimeIfExists(condPtr interface{}, lastUpdateTime metav1.Time) {
	utilruntime.Must(a.SetLastUpdateTimeIfExists(condPtr, lastUpdateTime))
}

// HasLastTransitionTime checks if the given condition has a 'LastTransitionTime' field.
//
// It errors if the given value is not a struct.
func (a *Accessor) HasLastTransitionTime(cond interface{}) (bool, error) {
	v, err := enforceStruct(cond)
	if err != nil {
		return false, err
	}

	return valueHasField(v, a.lastTransitionTimeField), nil
}

// MustHasLastTransitionTime checks if the given condition has a 'LastTransitionTime' field.
//
// It panics if the given value is not a struct.
func (a *Accessor) MustHasLastTransitionTime(cond interface{}) bool {
	ok, err := a.HasLastTransitionTime(cond)
	utilruntime.Must(err)
	return ok
}

// LastTransitionTime extracts the last transition time of the given condition.
//
// It errors if the given value is not a struct or does not have a field
// that can be converted to the output format.
func (a *Accessor) LastTransitionTime(cond interface{}) (metav1.Time, error) {
	v, err := enforceStruct(cond)
	if err != nil {
		return metav1.Time{}, err
	}

	var lastTransitionTime metav1.Time
	if err := getAndConvertField(v, a.lastTransitionTimeField, &lastTransitionTime); err != nil {
		return metav1.Time{}, err
	}
	return lastTransitionTime, nil
}

// MustLastTransitionTime extracts the last transition time of the given condition.
//
// It panics if the given value is not a struct or does not have a field
// that can be converted to the output format.
func (a *Accessor) MustLastTransitionTime(cond interface{}) metav1.Time {
	t, err := a.LastTransitionTime(cond)
	utilruntime.Must(err)
	return t
}

// SetLastTransitionTime sets the last transition time of the given condition if the field exists.
//
// It errors if the given value is not a pointer to a struct or does not have a field
// that can be converted to the output format.
func (a *Accessor) SetLastTransitionTime(condPtr interface{}, lastTransitionTime metav1.Time) error {
	v, err := enforcePtrToStruct(condPtr)
	if err != nil {
		return err
	}

	return setFieldConverted(v, a.lastTransitionTimeField, lastTransitionTime)
}

// MustSetLastTransitionTime sets the last transition time of the given condition.
//
// It panics if the given value is not a pointer to a struct or does not have a field
// that can be converted to the output format.
func (a *Accessor) MustSetLastTransitionTime(condPtr interface{}, lastTransitionTime metav1.Time) {
	utilruntime.Must(a.SetLastTransitionTime(condPtr, lastTransitionTime))
}

// SetLastTransitionTimeIfExists sets the last transition time of the given condition.
//
// It errors if the given value is not a pointer to a struct or the field value cannot be converted
// to the given format.
func (a *Accessor) SetLastTransitionTimeIfExists(condPtr interface{}, lastTransitionTime metav1.Time) error {
	condV, err := conversion.EnforcePtr(condPtr)
	if err != nil {
		return err
	}

	cond := condV.Interface()
	if ok, err := a.HasLastTransitionTime(cond); err != nil || !ok {
		return err
	}

	return a.SetLastTransitionTime(condPtr, lastTransitionTime)
}

// MustSetLastTransitionTimeIfExists sets the last transition time of the given condition.
//
// It panics if the given value is not a pointer to a struct or the field value cannot be converted
// to the given format.
func (a *Accessor) MustSetLastTransitionTimeIfExists(condPtr interface{}, lastTransitionTime metav1.Time) {
	utilruntime.Must(a.SetLastTransitionTimeIfExists(condPtr, lastTransitionTime))
}

// Reason extracts the reason of the given condition.
//
// It errors if the given value is not a struct or does not have a field
// that can be converted to the output format.
func (a *Accessor) Reason(cond interface{}) (string, error) {
	v, err := enforceStruct(cond)
	if err != nil {
		return "", err
	}

	var reason string
	if err := getAndConvertField(v, a.reasonField, &reason); err != nil {
		return "", err
	}
	return reason, nil
}

// MustReason extracts the reason of the given condition.
//
// It panics if the given value is not a struct or does not have a field
// that can be converted to the output format.
func (a *Accessor) MustReason(cond interface{}) string {
	s, err := a.Reason(cond)
	utilruntime.Must(err)
	return s
}

// SetReason sets the reason of the given condition.
//
// It errors if the given value is not a struct or does not have a field
// that can be converted to the given format.
func (a *Accessor) SetReason(condPtr interface{}, reason string) error {
	v, err := enforcePtrToStruct(condPtr)
	if err != nil {
		return err
	}

	return setFieldConverted(v, a.reasonField, reason)
}

// MustSetReason sets the reason of the given condition.
//
// It panics if the given value is not a struct or does not have a field
// that can be converted to the given format.
func (a *Accessor) MustSetReason(condPtr interface{}, reason string) {
	utilruntime.Must(a.SetReason(condPtr, reason))
}

// Message gets the message of the given condition.
//
// It errors if the given value is not a struct or does not have a field
// that can be converted to the input format.
func (a *Accessor) Message(cond interface{}) (string, error) {
	v, err := enforceStruct(cond)
	if err != nil {
		return "", err
	}

	var message string
	if err := getAndConvertField(v, a.messageField, &message); err != nil {
		return "", err
	}
	return message, nil
}

// MustMessage gets the message of the given condition.
//
// It panics if the given value is not a struct or does not have a field
// that can be converted to the input format.
func (a *Accessor) MustMessage(cond interface{}) string {
	s, err := a.Message(cond)
	utilruntime.Must(err)
	return s
}

// SetMessage sets the message of the given condition.
//
// It errors if the given value is not a struct or does not have a field
// that can be converted to the given format.
func (a *Accessor) SetMessage(condPtr interface{}, message string) error {
	v, err := enforcePtrToStruct(condPtr)
	if err != nil {
		return err
	}

	return setFieldConverted(v, a.messageField, message)
}

// MustSetMessage sets the message of the given condition.
//
// It panics if the given value is not a struct or does not have a field
// that can be converted to the given format.
func (a *Accessor) MustSetMessage(condPtr interface{}, message string) {
	utilruntime.Must(a.SetMessage(condPtr, message))
}

// HasObservedGeneration checks if the given condition has a observed generation field.
//
// It errors if the given value is not a struct.
func (a *Accessor) HasObservedGeneration(cond interface{}) (bool, error) {
	v, err := enforceStruct(cond)
	if err != nil {
		return false, err
	}

	return valueHasField(v, a.observedGenerationField), nil
}

// MustHasObservedGeneration checks if the given condition has a observed generation field.
//
// It panics if the given value is not a struct.
func (a *Accessor) MustHasObservedGeneration(cond interface{}) bool {
	ok, err := a.HasObservedGeneration(cond)
	utilruntime.Must(err)
	return ok
}

// ObservedGeneration gets the observed generation of the given condition.
//
// It errors if the given value is not a struct or does not have a field
// that can be converted to the input format.
func (a *Accessor) ObservedGeneration(cond interface{}) (int64, error) {
	v, err := enforceStruct(cond)
	if err != nil {
		return 0, err
	}

	var gen int64
	if err := getAndConvertField(v, a.observedGenerationField, &gen); err != nil {
		return 0, err
	}

	return gen, nil
}

// MustObservedGeneration gets the observed generation of the given condition.
//
// It panics if the given value is not a struct or does not have a field
// that can be converted to the input format.
func (a *Accessor) MustObservedGeneration(cond interface{}) int64 {
	gen, err := a.ObservedGeneration(cond)
	utilruntime.Must(err)
	return gen
}

// SetObservedGeneration sets the observed generation of the given condition.
//
// It errors if the given value is not a pointer to a struct or does not have a field
// that can be converted to the given format.
func (a *Accessor) SetObservedGeneration(condPtr interface{}, gen int64) error {
	v, err := enforcePtrToStruct(condPtr)
	if err != nil {
		return err
	}

	return setFieldConverted(v, a.observedGenerationField, gen)
}

// MustSetObservedGeneration sets the observed generation of the given condition.
//
// It panics if the given value is not a pointer to a struct or does not have a field
// that can be converted to the given format.
func (a *Accessor) MustSetObservedGeneration(condPtr interface{}, gen int64) {
	utilruntime.Must(a.SetObservedGeneration(condPtr, gen))
}

// MustSetMessage sets the message of the given condition.
//
// It panics if the given value is not a struct or does not have a field
// that can be converted to the given format.
func (a *Accessor) findTypeIndex(condSliceV reflect.Value, typ string) (int, error) {
	for i, n := 0, condSliceV.Len(); i < n; i++ {
		it := condSliceV.Index(i)
		itType, err := a.Type(it.Interface())
		if err != nil {
			return -1, fmt.Errorf("[index %d]: %w", i, err)
		}

		if itType == typ {
			return i, nil
		}
	}
	return -1, nil
}

// FindSliceIndex finds the index of the condition with the given type.
//
// If the target type is not found, -1 is returned.
// FindSliceIndex errors if condSlice is not a slice of structs.
func (a *Accessor) FindSliceIndex(condSlice interface{}, typ string) (int, error) {
	v, _, err := enforceStructSlice(condSlice)
	if err != nil {
		return 0, err
	}

	return a.findTypeIndex(v, typ)
}

// MustFindSliceIndex finds the index of the condition with the given type.
//
// If the target type is not found, -1 is returned.
// MustFindSliceIndex panics if condSlice is not a slice of structs.
func (a *Accessor) MustFindSliceIndex(condSlice interface{}, typ string) int {
	idx, err := a.FindSliceIndex(condSlice, typ)
	utilruntime.Must(err)
	return idx
}

// FindSlice finds the condition with the given type from the given slice and updates the target value with it.
//
// If the target type is not found, false is returned and the target value is not updated.
// FindSlice errors if condSlice is not a slice, intoPtr is not a pointer to a struct and if intoPtr's target
// value is not settable with an element of condSlice.
func (a *Accessor) FindSlice(condSlice interface{}, typ string, intoPtr interface{}) (ok bool, err error) {
	v, elemType, err := enforceStructSlice(condSlice)
	if err != nil {
		return false, err
	}

	intoV, err := enforcePtrToStruct(intoPtr)
	if err != nil {
		return false, err
	}

	if intoV.Type() != elemType {
		return false, fmt.Errorf("into type %T cannot accept slice type %T", intoPtr, condSlice)
	}

	idx, err := a.findTypeIndex(v, typ)
	if err != nil {
		return false, err
	}

	if idx == -1 {
		return false, nil
	}

	intoV.Set(v.Index(idx))
	return true, nil
}

// MustFindSlice finds the condition with the given type from the given slice and updates the target value with it.
//
// If the target type is not found, false is returned and the target value is not updated.
// FindSlice panics if condSlice is not a slice, intoPtr is not a pointer to a struct and if intoPtr's target
// value is not settable with an element of condSlice.
func (a *Accessor) MustFindSlice(condSlice interface{}, typ string, intoPtr interface{}) bool {
	ok, err := a.FindSlice(condSlice, typ, intoPtr)
	utilruntime.Must(err)
	return ok
}

// FindSliceStatus finds the status of the condition with the given type.
// If the condition cannot be found, corev1.ConditionUnknown is returned.
//
// FindSliceStatus errors if the given condSlice is not a slice of structs or if any
// of the conditions does not support access.
func (a *Accessor) FindSliceStatus(condSlice interface{}, typ string) (corev1.ConditionStatus, error) {
	v, _, err := enforceStructSlice(condSlice)
	if err != nil {
		return "", err
	}

	idx, err := a.findTypeIndex(v, typ)
	if err != nil {
		return "", err
	}

	if idx == -1 {
		return corev1.ConditionUnknown, nil
	}

	condV := v.Index(idx)
	return a.Status(condV.Interface())
}

// MustFindSliceStatus finds the status of the condition with the given type.
// If the condition cannot be found, corev1.ConditionUnknown is returned.
//
// MustFindSliceStatus errors if the given condSlice is not a slice of structs or if any
// of the conditions does not support access.
func (a *Accessor) MustFindSliceStatus(condSlice interface{}, typ string) corev1.ConditionStatus {
	status, err := a.FindSliceStatus(condSlice, typ)
	utilruntime.Must(err)
	return status
}

// UpdateOption is an option given to Accessor.UpdateSlice that updates an individual condition.
type UpdateOption interface {
	// ApplyUpdate applies the update, given a condition pointer.
	ApplyUpdate(a *Accessor, condPtr interface{}) error
}

// Update updates the condition with the given options, setting transition- and update time accordingly.
//
// Update errors if the given condPtr is not a pointer to a struct supporting the required condition fields.
func (a *Accessor) Update(condPtr interface{}, opts ...UpdateOption) error {
	if !a.disableTimestampUpdates {
		opts = []UpdateOption{
			UpdateTimestamps{
				Transition: a.transition,
				Clock:      a.clock,
				Updates:    opts,
			},
		}
	}

	for _, opt := range opts {
		if err := opt.ApplyUpdate(a, condPtr); err != nil {
			return err
		}
	}

	return nil
}

// MustUpdate updates the condition with the given options, setting transition- and update time accordingly.
//
// MustUpdate panics if the given condPtr is not a pointer to a struct supporting the required condition fields.
func (a *Accessor) MustUpdate(condPtr interface{}, opts ...UpdateOption) {
	utilruntime.Must(a.Update(condPtr, opts...))
}

// UpdateSlice finds and updates the condition with the given target type.
//
// UpdateSlice errors if condSlicePtr is not a pointer to a slice of structs that can be accessed with
// this Accessor.
// If no condition with the given type can be found, a new one is appended with the given type and updates
// applied.
// The last update time and last transition time of the new / existing condition is correctly updated:
// For new conditions, it's always set to the current time while for existing conditions, it's checked
// whether the status changed and then updated.
func (a *Accessor) UpdateSlice(condSlicePtr interface{}, typ string, opts ...UpdateOption) error {
	sliceV, elemType, err := enforcePtrToStructSlice(condSlicePtr)
	if err != nil {
		return err
	}

	idx, err := a.findTypeIndex(sliceV, typ)
	if err != nil {
		return err
	}

	var v reflect.Value
	if idx != -1 {
		v = sliceV.Index(idx).Addr()
	} else {
		v = reflect.New(elemType)
	}

	condPtr := v.Interface()

	// Ensure both type and initial transition time (if applicable) are set
	// for new conditions.
	if idx == -1 {
		if err := a.SetType(condPtr, typ); err != nil {
			return err
		}

		now := metav1.NewTime(a.clock.Now())
		if err := a.SetLastTransitionTimeIfExists(condPtr, now); err != nil {
			return err
		}
	}

	if err := a.Update(condPtr, opts...); err != nil {
		return err
	}

	// Make sure to append to the slice in case the condition is new, otherwise
	// it was already updated in-place.
	if idx == -1 {
		sliceV.Set(reflect.Append(sliceV, v.Elem()))
	}
	return nil
}

// MustUpdateSlice finds and updates the condition with the given target type.
//
// MustUpdateSlice panics if condSlicePtr is not a pointer to a slice of structs that can be accessed with
// this Accessor.
// If no condition with the given type can be found, a new one is appended with the given type and updates
// applied.
// The last update time and last transition time of the new / existing condition is correctly updated:
// For new conditions, it's always set to the current time while for existing conditions, it's checked
// whether the status changed and then updated.
func (a *Accessor) MustUpdateSlice(condSlicePtr interface{}, typ string, opts ...UpdateOption) {
	utilruntime.Must(a.UpdateSlice(condSlicePtr, typ, opts...))
}

// UpdateTimestamps manages the LastUpdateTime and LastTransitionTime field by creating a checkpoint with
// Transition, running all Updates and then checking if the TransitionCheckpoint reports transitioned.
// If so, the LastTransitionTimeField and the LastUpdateTimeField will be set to the current time using Clock (if
// Clock is unset, it uses clock.RealClock). Otherwise, only LastUpdateTimeField is updated..
type UpdateTimestamps struct {
	// Transition is the Transition to check whether a condition transitioned. Required.
	Transition Transition
	// Updates are all updates to run.
	Updates []UpdateOption
	// Clock is the clock to yield the current time. If unset, clock.RealClock is used.
	Clock clock.Clock
}

// UpdateTimestampsWith updates timestamps with the DefaultTransition and clock.RealClock. See UpdateTimestamps for
// more information.
func UpdateTimestampsWith(updates ...UpdateOption) UpdateOption {
	return UpdateTimestamps{
		Transition: DefaultTransition,
		Updates:    updates,
	}
}

// ApplyUpdate implements UpdateOption.
func (u UpdateTimestamps) ApplyUpdate(a *Accessor, condPtr interface{}) error {
	condV, err := enforcePtrToStruct(condPtr)
	if err != nil {
		return err
	}

	checkpoint, err := u.Transition.Checkpoint(a, condV.Interface())
	if err != nil {
		return err
	}

	for _, update := range u.Updates {
		if err := update.ApplyUpdate(a, condPtr); err != nil {
			return err
		}
	}

	c := u.Clock
	if c == nil {
		c = clock.RealClock{}
	}
	now := c.Now()

	ok, err := checkpoint.Transitioned(a, condV.Interface())
	if err != nil {
		return err
	}
	if ok {
		if err := a.SetLastTransitionTimeIfExists(condPtr, metav1.NewTime(now)); err != nil {
			return err
		}
	}

	if err := a.SetLastUpdateTimeIfExists(condPtr, metav1.NewTime(now)); err != nil {
		return err
	}

	return nil
}

// UpdateStatus implements UpdateOption to set a corev1.ConditionStatus.
type UpdateStatus corev1.ConditionStatus

// ApplyUpdate implements UpdateOption.
func (u UpdateStatus) ApplyUpdate(a *Accessor, condPtr interface{}) error {
	return a.SetStatus(condPtr, corev1.ConditionStatus(u))
}

// UpdateMessage implements UpdateOption to set the message.
type UpdateMessage string

// ApplyUpdate implements UpdateOption.
func (u UpdateMessage) ApplyUpdate(a *Accessor, condPtr interface{}) error {
	return a.SetMessage(condPtr, string(u))
}

// UpdateReason implements UpdateOption to set the reason.
type UpdateReason string

// ApplyUpdate implements UpdateOption.
func (u UpdateReason) ApplyUpdate(a *Accessor, condPtr interface{}) error {
	return a.SetReason(condPtr, string(u))
}

// UpdateObservedGeneration implements UpdateOption to set the observed generation.
type UpdateObservedGeneration int64

// ApplyUpdate implements UpdateOption.
func (u UpdateObservedGeneration) ApplyUpdate(a *Accessor, condPtr interface{}) error {
	return a.SetObservedGeneration(condPtr, int64(u))
}

// UpdateObserved is a shorthand for updating the observed generation from a metav1.Object's generation.
func UpdateObserved(obj metav1.Object) UpdateObservedGeneration {
	return UpdateObservedGeneration(obj.GetGeneration())
}

// UpdateFromCondition updates a condition from a source Condition.
//
// If Accessor is set, the ApplyUpdate function will use the Accessor for reading the properties for the source
// Condition over the one supplied. Otherwise, the Accessor supplied by ApplyUpdate will be used.
// All properties of the source condition will be transferred to the target condition with the exception of
// Type, LastTransitionTime and LastUpdateTime.
type UpdateFromCondition struct {
	// Accessor is the Accessor to access the Condition.
	// If unset, the Accessor from ApplyUpdate is used.
	Accessor *Accessor
	// Condition is a condition struct. Must not be nil.
	Condition interface{}
}

func (u UpdateFromCondition) ApplyUpdate(a *Accessor, condPtr interface{}) error {
	srcAccessor := u.Accessor
	if srcAccessor == nil {
		srcAccessor = a
	}

	status, err := srcAccessor.Status(u.Condition)
	if err != nil {
		return err
	}
	if err := a.SetStatus(condPtr, status); err != nil {
		return err
	}

	reason, err := srcAccessor.Reason(u.Condition)
	if err != nil {
		return err
	}
	if err := a.SetReason(condPtr, reason); err != nil {
		return err
	}

	message, err := srcAccessor.Message(u.Condition)
	if err != nil {
		return err
	}
	if err := a.SetMessage(condPtr, message); err != nil {
		return err
	}

	ok, err := srcAccessor.HasObservedGeneration(u.Condition)
	if err != nil {
		return err
	}
	if ok {
		observedGeneration, err := srcAccessor.ObservedGeneration(u.Condition)
		if err != nil {
			return err
		}
		if err := a.SetObservedGeneration(condPtr, observedGeneration); err != nil {
			return err
		}
	}
	return nil
}

// AccessorOptions are options to create an Accessor.
//
// If left blank, defaults are being used via AccessorOptions.SetDefaults.
type AccessorOptions struct {
	TypeField               string
	StatusField             string
	LastUpdateTimeField     string
	LastTransitionTimeField string
	ReasonField             string
	MessageField            string
	ObservedGenerationField string

	DisableTimestampUpdates bool
	Transition              Transition
	Clock                   clock.Clock
}

// SetDefaults sets default values for AccessorOptions.
func (o *AccessorOptions) SetDefaults() {
	if o.TypeField == "" {
		o.TypeField = DefaultTypeField
	}
	if o.StatusField == "" {
		o.StatusField = DefaultStatusField
	}
	if o.LastUpdateTimeField == "" {
		o.LastUpdateTimeField = DefaultLastUpdateTimeField
	}
	if o.LastTransitionTimeField == "" {
		o.LastTransitionTimeField = DefaultLastTransitionTimeField
	}
	if o.ReasonField == "" {
		o.ReasonField = DefaultReasonField
	}
	if o.MessageField == "" {
		o.MessageField = DefaultMessageField
	}
	if o.ObservedGenerationField == "" {
		o.ObservedGenerationField = DefaultObservedGenerationField
	}
	if o.Transition == nil {
		o.Transition = DefaultTransition
	}
	if o.Clock == nil {
		o.Clock = clock.RealClock{}
	}
}

// NewAccessor creates a new Accessor with the given AccessorOptions.
func NewAccessor(opts AccessorOptions) *Accessor {
	opts.SetDefaults()
	return &Accessor{
		typeField:               opts.TypeField,
		statusField:             opts.StatusField,
		lastUpdateTimeField:     opts.LastUpdateTimeField,
		lastTransitionTimeField: opts.LastTransitionTimeField,
		reasonField:             opts.ReasonField,
		messageField:            opts.MessageField,
		observedGenerationField: opts.ObservedGenerationField,
		disableTimestampUpdates: opts.DisableTimestampUpdates,
		transition:              opts.Transition,
		clock:                   opts.Clock,
	}
}
