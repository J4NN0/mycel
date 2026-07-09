package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/J4NN0/mycel/internal/agent"
	"github.com/J4NN0/mycel/internal/logger"
	"github.com/chzyer/readline"
)

const sessionID = "terminal"

type Cli struct {
	log logger.Logger
}

func New(log logger.Logger) *Cli {
	log.Printf("Terminal chat ready. Type a message and press Enter. Type / to see available commands.")
	return &Cli{log: log}
}

func (c *Cli) Run(ctx context.Context, handler agent.MessageHandler) error {
	rl, err := newReadline(ctx)
	if err != nil {
		return err
	}
	defer rl.Close()

	for {
		line, err := rl.Readline()
		if err != nil {
			if isExitErr(err, ctx) {
				return nil
			}
			return fmt.Errorf("readline: %w", err)
		}

		c.handleLine(ctx, handler, line)
	}
}

func newReadline(ctx context.Context) (*readline.Instance, error) {
	rl, err := readline.NewEx(&readline.Config{
		Prompt:       "> ",
		AutoComplete: commandCompleter(),
		Listener:     commandHintListener{},
	})
	if err != nil {
		return nil, fmt.Errorf("init readline: %w", err)
	}

	go func() {
		<-ctx.Done()
		rl.Close()
	}()

	return rl, nil
}

func isExitErr(err error, ctx context.Context) bool {
	return errors.Is(err, readline.ErrInterrupt) || errors.Is(err, io.EOF) || ctx.Err() != nil
}

func (c *Cli) handleLine(ctx context.Context, handler agent.MessageHandler, line string) {
	if strings.TrimSpace(line) == "" {
		return
	}

	response, err := handler(ctx, sessionID, line)
	if err != nil {
		c.log.Warningf("handler failed: %v", err)
		return
	}

	fmt.Printf("\nMycel: %s\n\n", response)
}
