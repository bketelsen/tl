package bdd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"
)

// --- completion-install.feature support ----------------------------------

func initializeCompletionInstallSteps(ctx *godog.ScenarioContext, w *world) {
	ctx.Step(`^environment variable "([^"]*)" is the scenario temp directory$`, w.setEnvToTempDir)
	ctx.Step(`^the file "([^"]*)" exists in the scenario temp directory$`, w.fileExistsInTempDir)
	ctx.Step(`^the output reports that the shell is unsupported$`, w.outputReportsUnsupportedShell)
}

func (w *world) setEnvToTempDir(key string) error {
	w.envOverrides = append(w.envOverrides, key)
	return os.Setenv(key, w.tempDir)
}

func (w *world) fileExistsInTempDir(rel string) error {
	abs := filepath.Join(w.tempDir, rel)
	info, err := os.Stat(abs)
	if err != nil {
		return fmt.Errorf("expected file at %q: %w", abs, err)
	}
	if info.IsDir() {
		return fmt.Errorf("expected file at %q but found a directory", abs)
	}
	if info.Size() == 0 {
		return fmt.Errorf("file at %q is empty", abs)
	}
	return nil
}

func (w *world) outputReportsUnsupportedShell() error {
	combined := strings.ToLower(w.stdout.String() + w.stderr.String())
	if w.cmdErr != nil {
		combined += "\n" + strings.ToLower(w.cmdErr.Error())
	}
	if !strings.Contains(combined, "unsupported") && !strings.Contains(combined, "could not detect") {
		return fmt.Errorf("output does not report an unsupported shell; got:\n%s", combined)
	}
	return nil
}
