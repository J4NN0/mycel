package cli

import (
	"context"
	"errors"
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"

	"github.com/J4NN0/mycel/internal/agent"
	"github.com/J4NN0/mycel/internal/logger"
)

type Cli struct {
	log  logger.Logger
	name string
}

func New(log logger.Logger, name string) *Cli {
	return &Cli{log: log, name: name}
}

func (c *Cli) Run(ctx context.Context, handler agent.MessageHandler) error {
	p := tea.NewProgram(
		newModel(ctx, c.name, handler),
		tea.WithContext(ctx),
	)

	c.log.SetOutput(logWriter{send: p.Send})
	defer c.log.SetOutput(os.Stdout)

	_, err := p.Run()
	if err != nil && !errors.Is(err, tea.ErrProgramKilled) && ctx.Err() == nil {
		return fmt.Errorf("tui: %w", err)
	}
	return nil
}
