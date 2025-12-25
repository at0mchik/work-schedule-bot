// internal/bot/handler/user_monthly_stat_handler.go
package handler

import (
    "fmt"
    "strconv"
    "strings"
    "time"

    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
    "github.com/sirupsen/logrus"
)

// getMyMonthlyStats показывает статистику пользователя за все месяцы
func (h *Handler) getMyMonthlyStats(message *tgbotapi.Message) {
    chatID := message.Chat.ID

    // Получаем пользователя
    user, err := h.userService.GetUser(chatID)
    if err != nil || user == nil {
        logrus.WithField("chat_id", chatID).Warn("User not found for stats")
        msg := tgbotapi.NewMessage(chatID, "❌ Профиль не найден.\nИспользуйте /createprofile чтобы создать профиль.")
        h.client.Bot.Send(msg)
        return
    }

    // Получаем статистику пользователя
    stats, err := h.userMonthlyStatService.GetUserStats(user.ID)
    if err != nil {
        logrus.WithError(err).Error("Failed to get user monthly stats")
        msg := tgbotapi.NewMessage(chatID, "❌ Ошибка получения статистики: "+err.Error())
        h.client.Bot.Send(msg)
        return
    }

    // Форматируем результат
    formatted := h.userMonthlyStatService.FormatStatsList(stats)
    msg := tgbotapi.NewMessage(chatID, formatted)
    h.client.Bot.Send(msg)
}

// getMonthlyStat показывает статистику за конкретный месяц
func (h *Handler) getMonthlyStat(message *tgbotapi.Message, args string) {
    chatID := message.Chat.ID

    // Получаем пользователя
    user, err := h.userService.GetUser(chatID)
    if err != nil || user == nil {
        logrus.WithField("chat_id", chatID).Warn("User not found for monthly stat")
        msg := tgbotapi.NewMessage(chatID, "❌ Профиль не найден.\nИспользуйте /createprofile чтобы создать профиль.")
        h.client.Bot.Send(msg)
        return
    }

    var year, month int
    now := time.Now()

    if args == "" {
        // Если месяц не указан, используем текущий
        year = now.Year()
        month = int(now.Month())
    } else {
        // Парсим месяц и год
        parts := strings.Fields(args)
        if len(parts) == 1 {
            // Только месяц
            month, err = strconv.Atoi(parts[0])
            if err != nil || month < 1 || month > 12 {
                msg := tgbotapi.NewMessage(chatID, "❌ Неверный месяц. Используйте число от 1 до 12.")
                h.client.Bot.Send(msg)
                return
            }
            year = now.Year()
        } else if len(parts) == 2 {
            // Год и месяц
            year, err = strconv.Atoi(parts[0])
            if err != nil || year < 2000 || year > 2100 {
                msg := tgbotapi.NewMessage(chatID, "❌ Неверный год. Используйте год между 2000 и 2100.")
                h.client.Bot.Send(msg)
                return
            }
            
            month, err = strconv.Atoi(parts[1])
            if err != nil || month < 1 || month > 12 {
                msg := tgbotapi.NewMessage(chatID, "❌ Неверный месяц. Используйте число от 1 до 12.")
                h.client.Bot.Send(msg)
                return
            }
        } else {
            msg := tgbotapi.NewMessage(chatID, "❌ Неверный формат. Используйте: /stat [год месяц] или /stat [месяц]")
            h.client.Bot.Send(msg)
            return
        }
    }

    // Получаем статистику
    stat, err := h.userMonthlyStatService.GetUserStatByMonth(user.ID, year, month)
    if err != nil {
        logrus.WithError(err).Error("Failed to get user monthly stat")
        msg := tgbotapi.NewMessage(chatID, "❌ Ошибка получения статистики: "+err.Error())
        h.client.Bot.Send(msg)
        return
    }

    if stat == nil {
        monthName := time.Month(month).String()
        msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("📭 Статистика за %s %d отсутствует.", monthName, year))
        h.client.Bot.Send(msg)
        return
    }

    // Форматируем результат
    formatted := h.userMonthlyStatService.FormatStat(stat)
    msg := tgbotapi.NewMessage(chatID, formatted)
    h.client.Bot.Send(msg)
}

// getCurrentMonthStat показывает статистику за текущий месяц
func (h *Handler) getCurrentMonthStat(message *tgbotapi.Message) {
    chatID := message.Chat.ID

    // Получаем пользователя
    user, err := h.userService.GetUser(chatID)
    if err != nil || user == nil {
        logrus.WithField("chat_id", chatID).Warn("User not found for current stat")
        msg := tgbotapi.NewMessage(chatID, "❌ Профиль не найден.\nИспользуйте /createprofile чтобы создать профиль.")
        h.client.Bot.Send(msg)
        return
    }

    // Получаем статистику за текущий месяц
    stat, err := h.userMonthlyStatService.GetCurrentMonthStat(user.ID)
    if err != nil {
        logrus.WithError(err).Error("Failed to get current month stat")
        msg := tgbotapi.NewMessage(chatID, "❌ Ошибка получения статистики: "+err.Error())
        h.client.Bot.Send(msg)
        return
    }

    if stat == nil {
        now := time.Now()
        monthName := now.Month().String()
        msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("📭 Статистика за текущий месяц (%s %d) отсутствует.", monthName, now.Year()))
        h.client.Bot.Send(msg)
        return
    }

    // Форматируем результат
    formatted := h.userMonthlyStatService.FormatStat(stat)
    msg := tgbotapi.NewMessage(chatID, formatted)
    h.client.Bot.Send(msg)
}