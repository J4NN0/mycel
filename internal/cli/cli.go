package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/J4NN0/mycel/internal/agent"
	"github.com/J4NN0/mycel/internal/logger"
)

const sessionID = "terminal"

type Cli struct {
	log logger.Logger
}

func New(log logger.Logger) *Cli {
	log.Printf("Terminal chat ready. Type a message and press Enter.")
	return &Cli{log: log}
}

func (c *Cli) Run(ctx context.Context, handler agent.MessageHandler) error {
	lines := make(chan string)
	go func() {
		defer close(lines)
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			select {
			case lines <- scanner.Text():
			case <-ctx.Done():
				return
			}
		}
	}()

	for {
		fmt.Print("> ")
		select {
		case <-ctx.Done():
			return nil
		case line, ok := <-lines:
			if !ok {
				return nil
			}
			if strings.TrimSpace(line) == "" {
				continue
			}
			response, err := handler(ctx, sessionID, line)
			if err != nil {
				c.log.Warningf("handler failed: %v", err)
				continue
			}
			fmt.Printf("\nMycel: %s\n\n", response)
		}
	}
}
