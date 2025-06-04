package token_parser

import (
	"fitness-trainer/internal/domain"
	"fmt"
	"time"

	initdata "github.com/telegram-mini-apps/init-data-golang"
)

type TelegramTokenParser interface {
	Parse(token string) (domain.TelegramTokenData, error)
}

type TelegramTokenParserImpl struct {
	botToken string
	expireIn time.Duration
}

func NewTelegramTokenParser(botToken string, expireIn time.Duration) TelegramTokenParser {
	return &TelegramTokenParserImpl{
		botToken: botToken,
		expireIn: expireIn,
	}
}

func (t *TelegramTokenParserImpl) Parse(token string) (domain.TelegramTokenData, error) {
	err := initdata.Validate(token, t.botToken, t.expireIn)
	if err != nil {
		return domain.TelegramTokenData{}, fmt.Errorf("invalid token: %w", err)
	}

	data, err := initdata.Parse(token)
	if err != nil {
		return domain.TelegramTokenData{}, fmt.Errorf("failed to parse token: %w", err)
	}

	return convertToDomain(data), nil
}

func convertToDomain(data initdata.InitData) domain.TelegramTokenData {
	return domain.TelegramTokenData{
		AuthDate:     data.AuthDate(),
		CanSendAfter: data.CanSendAfter(),
		Chat: domain.TelegramChat{
			ID:       data.Chat.ID,
			Type:     domain.ChatType(data.Chat.Type),
			Title:    data.Chat.Title,
			PhotoURL: data.Chat.PhotoURL,
			Username: data.Chat.Username,
		},
		QueryID:    data.QueryID,
		StartParam: data.StartParam,
		User: domain.TelegramUser{
			ID:                    data.User.ID,
			Username:              data.User.Username,
			FirstName:             data.User.FirstName,
			LastName:              data.User.LastName,
			IsBot:                 data.User.IsBot,
			IsPremium:             data.User.IsPremium,
			LanguageCode:          data.User.LanguageCode,
			PhotoURL:              data.User.PhotoURL,
			AddedToAttachmentMenu: data.User.AddedToAttachmentMenu,
			AllowsWriteToPm:       data.User.AllowsWriteToPm,
		},
	}
}
