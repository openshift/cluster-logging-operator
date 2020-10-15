package kverrors

import (
	"encoding/json"
	"errors"
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	_ Error                   = &KVError{}
	_ zapcore.ObjectMarshaler = &KVError{}
)

const (
	keyMessage string = "msg"
	keyCause   string = "cause"
)

// Error is a structured error
type Error interface {
	Ctx(Context) *KVError
	Error() string
	Message() string
	Unwrap() error
}

// New creates a new KVError with keys and values
func New(msg string, keysAndValues ...interface{}) *KVError {
	return &KVError{
		kv: appendMap(map[string]interface{}{
			keyMessage: msg,
		}, toMap(keysAndValues...)),
	}
}

// Wrap wraps an error as a new error with keys and values
func Wrap(err error, msg string, keysAndValues ...interface{}) *KVError {
	if err == nil {
		return nil
	}
	e := New(msg, append(keysAndValues, []interface{}{keyCause, err}...)...)
	return e
}

// KVError is an error that contains structured keys and values
type KVError struct {
	kv map[string]interface{}
}

// KVs returns the key/value pairs associated with this error
func (e *KVError) KVs() map[string]interface{} {
	return e.kv
}

// KVSlice returns the key/value pairs associated with this error
// as a slice
func (e *KVError) KVSlice() []interface{} {
	s := make([]interface{}, 0, len(e.kv)*2)
	for k, v := range e.kv {
		s = append(s, k, v)
	}
	return s
}

// Unwrap returns the error that caused this error
func (e *KVError) Unwrap() error {
	if cause, ok := e.kv[keyCause]; ok {
		e, _ := cause.(error)
		// if ok is false then e will be empty anyway so no need to check if ok
		return e
	}
	return nil
}

func (e *KVError) Error() string {
	base := e.Unwrap()
	if base != nil {
		return fmt.Sprintf("%s: %s", e.Message(), base.Error())
	}
	return e.Message()
}

func (e *KVError) Message() string {
	if msg, ok := e.kv[keyMessage]; ok {
		return fmt.Sprint(msg)
	}
	return ""
}

// Add adds key/value pairs to an error and returns the error
// WARNING: The original error is modified with this operation
func (e *KVError) Add(keyValuePairs ...interface{}) *KVError {
	for k, v := range toMap(keyValuePairs...) {
		e.kv[k] = v
	}
	return e
}

func (e *KVError) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.kv)
}

func (e *KVError) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	for k, v := range e.kv {
		zap.Any(k, v).AddTo(enc)
	}
	return nil
}

// Ctx appends Context to the error
func (e *KVError) Ctx(ctx Context) *KVError {
	e.kv = appendMap(e.kv, toMap(ctx...))
	return e
}

// Wrap sets err as the cause of this error and appends optional keysAndValues
func (e *KVError) Wrap(err error, keysAndValues ...interface{}) *KVError {
	ne := e.deepCopy().Add(keysAndValues...)
	ne.kv[keyCause] = err
	return ne
}

func (e *KVError) deepCopy() *KVError {
	return New(e.Message(), e.KVSlice()...)
}

// Unwrap provides compatibility with the standard errors package
func Unwrap(err error) error {
	return errors.Unwrap(err)
}

// Context creates key/value pairs to be used with errors later in
// the callstack. This provides the ability to create contextual
// information that will be used with any returned error
//
// Example:
//   errCtx := kverrors.Context("cluster", clusterName,
//       "namespace", namespace)
//
//   ...
//
//   if err != nil {
//       return kverrors.Wrap(err, "failed to get namespace").Ctx(errCtx)
//   }
//
//   ...
//
//   if err != nil {
//       return kverrors.Wrap(err, "failed to update cluster").Ctx(errCtx)
//   }
func NewContext(keysAndValues ...interface{}) Context {
	return keysAndValues
}

// Context is keyValuePairs wrapped to use later. Usage of this
// directly is not necessary
// See Context for more information
type Context []interface{}

// Root unwraps the error until it reaches the root error
func Root(err error) error {
	root := err
	for next := Unwrap(root); next != nil; next = Unwrap(root) {
		root = next
	}
	return root
}

func toMap(keysAndValues ...interface{}) map[string]interface{} {
	kve := map[string]interface{}{}

	for i, kv := range keysAndValues {
		if i%2 == 1 {
			continue
		}
		if len(keysAndValues) <= i+1 {
			continue
		}
		kve[fmt.Sprintf("%s", kv)] = keysAndValues[i+1]
	}
	return kve
}

func appendMap(a, b map[string]interface{}) map[string]interface{} {
	for k, v := range b {
		a[k] = v
	}
	return a
}
