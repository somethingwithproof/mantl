// SPDX-License-Identifier: Apache-2.0

package main

import (
	"errors"
	"github.com/go-logr/logr"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestStartupSuccessDoesNotExit(t *testing.T) {
	failStartup(nil, "ready")
}

func TestStartupFailureProcess(t *testing.T) {
	if os.Getenv("MANTL_TEST_STARTUP_FAILURE") != "1" {
		return
	}
	setupLog = logr.FromSlogHandler(slog.NewJSONHandler(os.Stderr, nil))
	failStartup(errors.New("fixture dependency unavailable"), "required dependency failed", "controller", "fixture")
	t.Fatal("startup continued after a required failure")
}

func TestRequiredStartupFailureExitsWithDiagnostic(t *testing.T) {
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(binary, "-test.run=^TestStartupFailureProcess$")
	cmd.Env = append(os.Environ(), "MANTL_TEST_STARTUP_FAILURE=1")
	output, err := cmd.CombinedOutput()
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 1 {
		t.Fatalf("startup exit: %v %s", err, output)
	}
	for _, want := range []string{"required dependency failed", "fixture dependency unavailable", `"controller":"fixture"`} {
		if !strings.Contains(string(output), want) {
			t.Fatalf("missing diagnostic %q: %s", want, output)
		}
	}
}
