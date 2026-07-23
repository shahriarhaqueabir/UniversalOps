package common

import (
	"context"
	"time"

	"github.com/yusufpapurcu/wmi"
)

// WMIQueryWithTimeout performs a WMI query with a context timeout.
// This prevents the application from hanging indefinitely if the WMI service
// or a specific provider (like LibreHardwareMonitor) stalls.
func WMIQueryWithTimeout(query string, dst interface{}, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	errChan := make(chan error, 1)
	go func() {
		defer RecoverPanic()
		errChan <- wmi.Query(query, dst)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errChan:
		return err
	}
}

// WMIQueryNamespaceWithTimeout performs a WMI query on a specific namespace with a timeout.
func WMIQueryNamespaceWithTimeout(query string, dst interface{}, namespace string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	errChan := make(chan error, 1)
	go func() {
		defer RecoverPanic()
		errChan <- wmi.QueryNamespace(query, dst, namespace)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errChan:
		return err
	}
}
