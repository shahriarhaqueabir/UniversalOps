package common

import "runtime"

// RecoverPanic recovers from a panic in a goroutine, logs the panic value
// and stack trace, and prevents the application from crashing.
// Usage: defer common.RecoverPanic()
func RecoverPanic() {
	if r := recover(); r != nil {
		buf := make([]byte, 4096)
		n := runtime.Stack(buf, false)
		LogError("PANIC recovered: %v\nStack:\n%s", r, buf[:n])
	}
}
