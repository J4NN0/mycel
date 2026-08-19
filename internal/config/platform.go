package config

type Platform struct {
	Telegram Telegram
}

type Telegram struct {
	BotToken       string  `envconfig:"TELEGRAM_BOT_TOKEN"`
	AllowedUserIDs []int64 `envconfig:"TELEGRAM_ALLOWED_USER_IDS"`
}
