package app

import (
	"net"
	"os"
	"syscall"
	"testing"

	"gopkg.in/yaml.v3"
)

// freePort asks the OS for a currently-free TCP port so admin-server tests
// don't race on a fixed port.
func freePort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("freePort: %v", err)
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}

// sendInterrupt raises SIGINT on the current process so Run's signal handler
// fires. The handler is registered by Run before we send, so this delivers a
// real signal into the test binary's own process.
func sendInterrupt() { _ = syscall.Kill(syscall.Getpid(), syscall.SIGINT) }

// osWriteFile wraps os.WriteFile so lifecycle tests stay free of direct os
// imports for the file path.
func osWriteFile(path string, bs []byte) error { return os.WriteFile(path, bs, 0o644) }

// osUnsetenv unsets an environment variable; wrapped so tests that check the
// "unset placeholder" path can clear a var even when t.Setenv set it.
func osUnsetenv(key string) { _ = os.Unsetenv(key) }

// yamlUnmarshal is the unmarshal fn passed to ReadConfig in tests.
func yamlUnmarshal(bs []byte, v any) error { return yaml.Unmarshal(bs, v) }
