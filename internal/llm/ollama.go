package llm

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

func ensureOllama(model string) error {
	if err := ensureServing(); err != nil {
		return fmt.Errorf("failed to start ollama: %w", err)
	}
	if err := ensureModelPulled(model); err != nil {
		return fmt.Errorf("failed to pull model %q: %w", model, err)
	}
	return nil
}

func ensureServing() error {
	check := exec.Command("ollama", "list")
	if err := check.Run(); err == nil {
		return nil
	}

	cmd := exec.Command("ollama", "serve")
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("could not start ollama serve: %w", err)
	}

	time.Sleep(2 * time.Second)

	return nil
}

func ensureModelPulled(model string) error {
	cmd := exec.Command("ollama", "pull", model)
	cmd.Stdout = os.Stdout // stream pull progress to terminal
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
