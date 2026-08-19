package config

type Tool struct {
	Resend  Resend
	Searxng Searxng
}

type Resend struct {
	APIKey string `envconfig:"RESEND_API_KEY"`
	From   string `envconfig:"RESEND_FROM"`
}

type Searxng struct {
	URL string `envconfig:"SEARXNG_URL" default:"http://localhost:8888"`
}
