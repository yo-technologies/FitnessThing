package domain

// LLMLimits описывает состояние лимитов/квот LLM для пользователя на текущий день
type LLMLimits struct {
	DailyLimit int // дневной лимит токенов
	Used       int // использовано
	Reserved   int // зарезервировано (в процессе)
	Remaining  int // остаток = max(daily - used - reserved, 0)
}

// NewLLMLimits конструирует структуру лимитов и вычисляет Remaining
func NewLLMLimits(dailyLimit, used, reserved int) LLMLimits {
	rem := max(dailyLimit-used-reserved, 0)
	return LLMLimits{
		DailyLimit: dailyLimit,
		Used:       used,
		Reserved:   reserved,
		Remaining:  rem,
	}
}
