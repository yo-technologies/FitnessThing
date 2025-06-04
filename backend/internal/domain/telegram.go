package domain

import "time"

const (
	ChatTypeSender     ChatType = "sender"
	ChatTypePrivate    ChatType = "private"
	ChatTypeGroup      ChatType = "group"
	ChatTypeSupergroup ChatType = "supergroup"
	ChatTypeChannel    ChatType = "channel"
)

type ChatType string

type TelegramChat struct {
	ID       int64
	Type     ChatType
	Title    string
	PhotoURL string
	Username string
}

type TelegramUser struct {
	ID        int64
	Username  string
	FirstName string
	LastName  string

	IsBot     bool
	IsPremium bool

	LanguageCode string
	PhotoURL     string

	AddedToAttachmentMenu bool
	AllowsWriteToPm       bool
}

type TelegramTokenData struct {
	AuthDate     time.Time
	CanSendAfter time.Time
	Chat         TelegramChat
	QueryID      string
	StartParam   string
	User         TelegramUser
}
