package telegram

import (
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/J4NN0/mycel/internal/agent"
	"github.com/J4NN0/mycel/internal/logger"
)

type Bot struct {
	api *tgbotapi.BotAPI
	log logger.Logger
}

func NewBot(token string, log logger.Logger) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("failed to create telegram bot: %w", err)
	}

	botCommands := make([]tgbotapi.BotCommand, len(agent.Commands))
	for i, c := range agent.Commands {
		botCommands[i] = tgbotapi.BotCommand{Command: c.Name, Description: c.Description}
	}
	_, err = api.Request(tgbotapi.NewSetMyCommands(botCommands...))
	if err != nil {
		return nil, fmt.Errorf("failed to register bot commands: %w", err)
	}

	return &Bot{api: api, log: log}, nil
}

func (b *Bot) Run(ctx context.Context, handler agent.MessageHandler) error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	b.log.Printf("Running telegram bot @%s", b.api.Self.UserName)

	for {
		select {
		case <-ctx.Done():
			b.api.StopReceivingUpdates()
			return nil
		case update, ok := <-updates:
			if !ok {
				return nil
			}
			if update.Message == nil || update.Message.Text == "" {
				continue
			}
			go b.handleMessage(ctx, update.Message, handler)
		}
	}
}

func (b *Bot) handleMessage(ctx context.Context, msg *tgbotapi.Message, handler agent.MessageHandler) {
	sessionID := fmt.Sprintf("%d", msg.Chat.ID)
	b.log.Printf("[%s] New message from %s: %s", sessionID, msg.From.UserName, msg.Text)

	response, err := handler(ctx, sessionID, msg.Text)
	if err != nil {
		b.log.Warningf("handler failed: %v", err)
		return
	}

	b.log.Printf("[%s] Replying to %s: %s", sessionID, msg.From.UserName, response)
	reply := tgbotapi.NewMessage(msg.Chat.ID, response)
	if !msg.IsCommand() {
		reply.ReplyToMessageID = msg.MessageID
	}

	_, err = b.api.Send(reply)
	if err != nil {
		b.log.Warningf("failed to send telegram message: %v", err)
	}
}
