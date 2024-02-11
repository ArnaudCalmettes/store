package test

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// DefaultTimeout is the default timeout given to test contexts.
var DefaultTimeout = 20 * time.Second

// NewTestContext returns a context with a sensible default timeout for tests.
func NewTestContext() (context.Context, context.CancelFunc) {
	return TestContext(DefaultTimeout)
}

// TestContext is a short wrapper around context.WithTimeout.
func TestContext(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}

// Expect allows to chain helpers in imperative-style testing
func Expect(t *testing.T, errs ...error) {
	t.Helper()
	for _, err := range errs {
		if err != nil {
			t.Error(err)
		}
	}
}

// Require allows to chain helpers in imperative-style testing
func Require(t *testing.T, errs ...error) {
	t.Helper()
	for _, err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
}

// NoError checks that the returned error is nil.
func NoError(err error) error {
	if err != nil {
		return fmt.Errorf("unexpected error: %w", err)
	}
	return nil
}

// NoErrorf checks that the returned error is nil.
func NoErrorf(err error, msg string, args ...any) error {
	if err = NoError(err); err != nil {
		return fmt.Errorf("%s: %w", fmt.Sprintf(msg, args...), err)
	}
	return nil
}

// IsError checks that the returned error (got) is the one expected (want).
// If want is nil, it behaves the same as NoError.
func IsError(want, got error) error {
	if want == nil {
		return NoError(got)
	}
	if !errors.Is(got, want) {
		return fmt.Errorf("expected error %q, got: %v", want, got)
	}
	return nil
}

// IsErrorf checks that the returned error is the expected one.
func IsErrorf(want, got error, msg string, args ...any) error {
	if err := IsError(want, got); err != nil {
		return fmt.Errorf("%s: %w", fmt.Sprintf(msg, args...), err)
	}
	return nil
}

// ShouldPanic returns an error if given function does not panic when called.
func ShouldPanic(f func()) error {
	var didPanic bool
	var message any
	func() {
		defer func() {
			if message = recover(); message != nil {
				didPanic = true
			}
		}()
		f()
	}()
	if !didPanic {
		return fmt.Errorf("function was expected to panic but didn't")
	}
	return nil
}

func DoesNotPanic(f func()) (err error) {
	defer func() {
		if message := recover(); message != nil {
			err = fmt.Errorf("unexpected panic: %s", message)
		}
	}()
	f()
	return nil
}

var IgnoreUnexported = cmpopts.IgnoreUnexported
var IgnoreFields = cmpopts.IgnoreFields

// Equal returns an error if values are different.
func Equal[T any](want, got T, opts ...cmp.Option) error {
	if !cmp.Equal(want, got, opts...) {
		return fmt.Errorf(cmp.Diff(want, got, opts...))
	}
	return nil
}

// NotEqual returns an error if values are the same.
func NotEqual[T any](want, got T, opts ...cmp.Option) error {
	if cmp.Equal(want, got, opts...) {
		return fmt.Errorf("values are equal")
	}
	return nil
}

// Equalf returns an error if values are not "DeepEqual".
func Equalf[T any](want, got T, msg string, args ...any) error {
	if err := Equal(want, got); err != nil {
		return fmt.Errorf("%s:\n%w", fmt.Sprintf(msg, args...), err)
	}
	return nil
}

// NotEqualf returns an error if values are the same.
func NotEqualf[T any](want, got T, msg string, args ...any) error {
	if err := Equal(want, got); err == nil {
		return fmt.Errorf("%s:\n%w", fmt.Sprintf(msg, args...), err)
	}
	return nil
}

// IsNotZero returns an error if the value is not a zero value.
func IsNotZero[T any](have T, opts ...cmp.Option) error {
	var zero T
	if cmp.Equal(zero, have, opts...) {
		return fmt.Errorf("expected non-zero value, got %v", have)
	}
	return nil
}

// IsNotZerof returns an error with a custom message if the value is not a zero value.
func IsNotZerof[T any](have T, msg string, args ...any) error {
	if err := IsNotZero(have); err != nil {
		return fmt.Errorf("%s: %w", fmt.Sprintf(msg, args...), err)
	}
	return nil
}

type CheckFunc[T any] func(T, error) error

func All[T any](checks ...CheckFunc[T]) CheckFunc[T] {
	return func(got T, err error) error {
		for _, check := range checks {
			if err := check(got, err); err != nil {
				return err
			}
		}
		return nil
	}
}

// AsCheckFunc transforms a basic want/got checker into a CheckFunc generator.
func AsCheckFunc[W any, G any](check func(W, G) error) func(W) CheckFunc[G] {
	return func(want W) CheckFunc[G] {
		return func(got G, _ error) error { return check(want, got) }
	}
}

func HasError[T any](want error) CheckFunc[T] {
	return func(_ T, err error) error { return IsError(want, err) }
}

// HasNoError is a CheckFunc that asserts the absence of an error.
func HasNoError[T any](_ T, err error) error {
	return NoError(err)
}

// IsNilPointer returns an error if given pointer is not nil.
func IsNilPointer[T any](ptr *T) error {
	if ptr != nil {
		return fmt.Errorf("expected nil %T, got %+v", ptr, ptr)
	}
	return nil
}

// IsNilPointerf returns an error if given pointer is not nil.
func IsNilPointerf[T any](ptr *T, msg string, args ...any) error {
	if ptr != nil {
		return fmt.Errorf("%s: expected nil %T, got %+v",
			fmt.Sprintf(msg, args...), ptr, ptr,
		)
	}
	return nil
}

// IsNotNilPointer returns an error if given pointer is nil.
func IsNotNilPointer[T any](ptr *T) error {
	if ptr == nil {
		return fmt.Errorf("unexpected nil %T", ptr)
	}
	return nil
}

func IsSamePointer[T any](want, got *T) error {
	if want != got {
		return fmt.Errorf("expected (%T) %p, got %p", want, want, got)
	}
	return nil
}

func IsSamePointerf[T any](want, got *T, msg string, args ...any) error {
	if err := IsSamePointer(want, got); err != nil {
		return fmt.Errorf("%s:\n%w", fmt.Sprintf(msg, args...), err)
	}
	return nil
}

func IsNotEmptySlice[T any](slice []T) error {
	if len(slice) == 0 {
		return fmt.Errorf("expected a slice with at least one element")
	}
	return nil
}

func IsEmptySlice[T any](slice []T) error {
	return SliceHasLength(0, slice)
}

func SliceHasLength[T any](want int, slice []T) error {
	if len(slice) != want {
		return fmt.Errorf("expected slice with length %d, got %v", want, slice)
	}
	return nil
}

func SliceContains[T comparable](want T, slice []T) error {
	for _, v := range slice {
		if v == want {
			return nil
		}
	}

	return fmt.Errorf("slice does not contain %v, got %v", want, slice)
}

func SliceDoesNotContain[T comparable](want T, slice []T) error {
	for _, v := range slice {
		if v == want {
			return fmt.Errorf("slice contains %v", want)
		}
	}

	return nil
}

// IsNotNilPointer returns an error if given pointer is nil.
func IsNotNilPointerf[T any](ptr *T, msg string, args ...any) error {
	if ptr == nil {
		return fmt.Errorf("%s: unexpected nil %T", fmt.Sprintf(msg, args...), ptr)
	}
	return nil
}

func IsZero[T any](obj T) error {
	var zero T
	if !reflect.DeepEqual(zero, obj) {
		return fmt.Errorf("expected zero %T, got %v", zero, obj)
	}
	return nil
}

var ErrChannelClosed = errors.New("channel was closed")
var ErrReadTimeout = errors.New("context timed out while reading channel")

func ReadChannel[T any](ctx context.Context, ch <-chan T) (T, error) {
	var m T
	var ok bool
	select {
	case <-ctx.Done():
		return m, ErrReadTimeout
	case m, ok = <-ch:
		if !ok {
			return m, ErrChannelClosed
		}
		return m, nil
	}
}

func PointerTo[T any](obj T) *T {
	return &obj
}
