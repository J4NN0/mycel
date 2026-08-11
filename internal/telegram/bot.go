package telegram

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/J4NN0/mycel/internal/agent"
	"github.com/J4NN0/mycel/internal/llm"
	"github.com/J4NN0/mycel/internal/logger"
)

const (
	resumeCallbackPrefix = "resume:"

	// maxPhotoBytes caps a download. Telegram's own getFile limit for bots is 20MB and its compressed
	// photos land far below that, so this is a backstop against reading an unbounded body into memory.
	maxPhotoBytes = 20 << 20
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

	log.Printf("Telegram bot @%s ready", api.Self.UserName)

	return &Bot{api: api, log: log}, nil
}

func (b *Bot) Run(ctx context.Context, handler agent.MessageHandler) error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			b.api.StopReceivingUpdates()
			return nil
		case update, ok := <-updates:
			if !ok {
				return nil
			}
			if update.CallbackQuery != nil {
				go b.handleCallback(ctx, update.CallbackQuery, handler)
				continue
			}
			if update.Message == nil {
				continue
			}
			if update.Message.Text == "" && len(update.Message.Photo) == 0 {
				continue
			}
			go b.handleMessage(ctx, update.Message, handler)
		}
	}
}

func (b *Bot) handleMessage(ctx context.Context, msg *tgbotapi.Message, handler agent.MessageHandler) {
	sessionID := fmt.Sprintf("%d", msg.Chat.ID)

	input := agent.Input{Text: msg.Text}
	if len(msg.Photo) > 0 {
		input.Text = msg.Caption
		image, err := b.downloadPhoto(ctx, msg.Photo)
		if err != nil {
			b.log.Warningf("[%s] Failed to download photo: %v", sessionID, err)
			b.reply(msg, "I couldn't download that image, sorry.")
			return
		}
		input.Images = []llm.Image{image}
	}

	b.log.Debugf("[%s] Message from %s: %s (%d image(s))", sessionID, msg.From.UserName, input.Text, len(input.Images))

	// No streaming: Telegram rate-limits message edits, so the reply is sent in one go.
	result, err := handler(ctx, sessionID, input, nil)
	if err != nil {
		b.log.Warningf("handler failed: %v", err)
		return
	}

	if len(result.Conversations) > 0 {
		b.sendResumePicker(msg.Chat.ID, sessionID, result.Conversations)
		return
	}

	b.log.Debugf("[%s] Replying to %s: %s", sessionID, msg.From.UserName, result.Text)
	b.reply(msg, result.Text)
}

func (b *Bot) reply(msg *tgbotapi.Message, text string) {
	replyMsg := tgbotapi.NewMessage(msg.Chat.ID, text)
	if !msg.IsCommand() {
		replyMsg.ReplyToMessageID = msg.MessageID
	}

	if _, err := b.api.Send(replyMsg); err != nil {
		b.log.Warningf("failed to send telegram message: %v", err)
	}
}

func (b *Bot) downloadPhoto(ctx context.Context, sizes []tgbotapi.PhotoSize) (llm.Image, error) {
	largest := sizes[len(sizes)-1]

	url, err := b.api.GetFileDirectURL(largest.FileID)
	if err != nil {
		return llm.Image{}, fmt.Errorf("resolve file url: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return llm.Image{}, fmt.Errorf("build request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return llm.Image{}, fmt.Errorf("download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return llm.Image{}, fmt.Errorf("download returned %s", resp.Status)
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, maxPhotoBytes+1))
	if err != nil {
		return llm.Image{}, fmt.Errorf("read body: %w", err)
	}
	if len(data) > maxPhotoBytes {
		return llm.Image{}, fmt.Errorf("photo is larger than %dMB", maxPhotoBytes>>20)
	}

	return llm.Image{Data: data}, nil
}

func (b *Bot) sendResumePicker(chatID int64, sessionID string, conversations []agent.Conversation) {
	rows := make([][]tgbotapi.InlineKeyboardButton, len(conversations))
	for i, c := range conversations {
		rows[i] = tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(c.Preview, resumeCallbackPrefix+c.ID),
		)
	}

	msg := tgbotapi.NewMessage(chatID, "Pick a conversation to resume:")
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
	if _, err := b.api.Send(msg); err != nil {
		b.log.Warningf("[%s] Failed to send resume picker: %v", sessionID, err)
	}
}

// handleCallback dispatches an inline-keyboard tap by its callback_data prefix. Only "resume:" is
// recognized today; anything else isn't ours to handle, so it's left untouched.
func (b *Bot) handleCallback(ctx context.Context, cq *tgbotapi.CallbackQuery, handler agent.MessageHandler) {
	conversationID, ok := strings.CutPrefix(cq.Data, resumeCallbackPrefix)
	if !ok {
		return
	}

	b.handleResumeCallback(ctx, cq, conversationID, handler)
}

func (b *Bot) handleResumeCallback(ctx context.Context, cq *tgbotapi.CallbackQuery, conversationID string, handler agent.MessageHandler) {
	if cq.Message == nil {
		b.answerCallback(cq, "This selection expired, run /resume again.")
		return
	}

	sessionID := fmt.Sprintf("%d", cq.Message.Chat.ID)
	result, err := handler(ctx, sessionID, agent.Input{Text: "/resume " + conversationID}, nil)

	var text string
	if err != nil {
		b.log.Warningf("[%s] Failed to resume conversation %s: %v", sessionID, conversationID, err)
		text = "Failed to resume that conversation."
	} else {
		text = result.Text
	}

	b.answerCallback(cq, "")
	b.editMessage(cq.Message.Chat.ID, cq.Message.MessageID, text)
}

func (b *Bot) answerCallback(cq *tgbotapi.CallbackQuery, alert string) {
	if _, err := b.api.Request(tgbotapi.NewCallback(cq.ID, alert)); err != nil {
		b.log.Warningf("failed to answer callback %s: %v", cq.ID, err)
	}
}

func (b *Bot) editMessage(chatID int64, messageID int, text string) {
	edit := tgbotapi.NewEditMessageTextAndMarkup(chatID, messageID, text, tgbotapi.InlineKeyboardMarkup{})
	if _, err := b.api.Send(edit); err != nil {
		b.log.Warningf("failed to edit message %d in chat %d: %v", messageID, chatID, err)
	}
}
