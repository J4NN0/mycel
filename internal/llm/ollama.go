package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/J4NN0/mycel/internal/logger"
)

const (
	ollamaBaseURL = "http://localhost:11434"

	capabilityCompletion = "completion"
	capabilityTools      = "tools"
	capabilityVision     = "vision"
	capabilityThinking   = "thinking"
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

var capabilityLosses = []struct {
	capability string
	loss       string
}{
	{capabilityCompletion, "cannot generate text: every reply will fail"},
	{capabilityTools, "cannot call tools: the agent's tools will be ignored"},
	{capabilityVision, "cannot process images: messages carrying one will be rejected"},
	{capabilityThinking, "does not reason before answering"},
}

type modelCapabilities map[string]bool

func (c modelCapabilities) has(capability string) bool {
	return c[capability]
}

func fetchCapabilities(log logger.Logger, model string) (modelCapabilities, error) {
	body, err := json.Marshal(map[string]string{"model": model})
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(ollamaBaseURL+"/api/show", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("call /api/show: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("/api/show returned %s", resp.Status)
	}

	var show struct {
		Capabilities []string `json:"capabilities"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&show); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	capabilities := make(modelCapabilities, len(show.Capabilities))
	for _, c := range show.Capabilities {
		capabilities[c] = true
	}
	warnMissingCapabilities(log, model, capabilities)

	return capabilities, nil
}

func warnMissingCapabilities(log logger.Logger, model string, capabilities modelCapabilities) {
	for _, c := range capabilityLosses {
		if !capabilities.has(c.capability) {
			log.Warningf("Model %s %s", model, c.loss)
		}
	}
}
