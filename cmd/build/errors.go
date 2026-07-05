package main

import (
	"errors"
	"fmt"
	"os/exec"
)

// errMissingDep indicates a required build dependency is not installed.
// It is returned by buildLib and buildWasm when exec.LookPath fails or
// when a required environment variable (e.g. EMSDK) is not set.
type errMissingDep struct {
	tool string
}

func (e errMissingDep) Error() string {
	return fmt.Sprintf("missing build dependency: %s", e.tool)
}

// isDepMissingError returns true if the error indicates a missing build
// dependency (e.g. cmake or emcmake command not found, EMSDK not set).
// Uses errors.Is and errors.As for reliable cross-platform detection
// instead of fragile string matching.
func isDepMissingError(err error) bool {
	if err == nil {
		return false
	}
	var md errMissingDep
	return errors.As(err, &md) || errors.Is(err, exec.ErrNotFound)
}
