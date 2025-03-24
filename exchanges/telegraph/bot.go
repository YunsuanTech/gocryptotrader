package telegram // 修改为你想要的包名

import (
	"context"
	"fmt"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// Bot 结构体包含机器人实例
type Bot struct {
	bot *bot.Bot
}

// NewBot 创建一个新的机器人实例
func NewBot(token string) (*Bot, error) {
	opts := []bot.Option{
		bot.WithDefaultHandler(defaultHandler),
	}

	b, err := bot.New(token, opts...)
	if err != nil {
		return nil, err
	}

	return &Bot{bot: b}, nil
}

// Start 启动机器人
func (b *Bot) Start(ctx context.Context) {
	b.bot.Start(ctx)
}

// SendMessage 发送消息
func (b *Bot) SendMessage(ctx context.Context, chatID int64, text string) error {
	_, err := b.bot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   text,
	})
	return err
}

// 默认的消息处理函数
func defaultHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message != nil {
		_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   update.Message.Text,
		})
		fmt.Printf("Your Chat ID is: %d\n", update.Message.Chat.ID)
	}
}
