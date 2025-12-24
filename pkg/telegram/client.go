package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Client struct {
	Bot           *tgbotapi.BotAPI
	UpdateConfig  tgbotapi.UpdateConfig
}

func NewClient(token string) (*Client, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	bot.Debug = true // отладкa

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	return &Client{
		Bot:          bot,
		UpdateConfig: updateConfig,
	}, nil
}