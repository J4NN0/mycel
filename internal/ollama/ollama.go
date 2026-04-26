package ollama

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/J4NN0/mycel/internal/logger"
)

type Checker interface {
	Ensure(model string) error
}

type ollama struct {
	log logger.Logger
}

func NewChecker(log logger.Logger) Checker {
	return &ollama{log: log}
}

func (o *ollama) Ensure(model string) error {
	if err := o.ensureServing(); err != nil {
		return fmt.Errorf("failed to start ollama: %w", err)
	}
	if err := o.ensureModelPulled(model); err != nil {
		return fmt.Errorf("failed to pull model %q: %w", model, err)
	}
	return nil
}

func (o *ollama) ensureServing() error {
	check := exec.Command("ollama", "list")
	if err := check.Run(); err == nil {
		o.log.Printf("Ollama already running")
		return nil
	}

	cmd := exec.Command("ollama", "serve")
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("could not start ollama serve: %w", err)
	}

	o.log.Printf("Ollama started")
	time.Sleep(2 * time.Second)

	return nil
}

func (o *ollama) ensureModelPulled(model string) error {
	o.log.Printf("Pulling model %q ...\n", model)

	cmd := exec.Command("ollama", "pull", model)
	cmd.Stdout = os.Stdout // stream pull progress to terminal
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
