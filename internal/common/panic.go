package common

import "runtime"

// RecoverPanic recovers from a panic in a goroutine, logs the panic value
// and stack trace, and prevents the application from crashing.
// Usage: defer common.RecoverPanic()
func RecoverPanic() {
	RecoverPanicWithContext("")
}

// RecoverPanicWithContext recovers from a panic with a module identifier
// so the log output identifies which subsystem crashed.
// Usage: defer common.RecoverPanicWithContext("scheduler.cpu")
func RecoverPanicWithContext(module string) {
	if r := recover(); r != nil {
		buf := make([]byte, 4096)
		n := runtime.Stack(buf, false)
		if module != "" {
			LogError("PANIC recovered in %s: %v\nStack:\n%s", module, r, buf[:n])
		} else {
			LogError("PANIC recovered: %v\nStack:\n%s", r, buf[:n])
		}
	}
}
