// package client provides a client for tests.
//
// This Client can operate on any k8s resource struct type, core or custom,
// via the runtime.Object interface.
// As well as the usual Create, Get, Update, Delete it supports  Watch
// and provides some convenience methods for tests.
//
// Most tests should use the singleton client returned by Get().
// Avoiding creation of multiple clients saves significant time in a test suite.
package client
