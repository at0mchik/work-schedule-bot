package handler

import (
	"fmt"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
	"work-schedule-bot/internal/models"
)

// addVacation добавляет отпуск
func (h *Handler) addVacation(message *tgbotapi.Message, args string) {
	chatID := message.Chat.ID

	// Получаем пользователя
	user, err := h.userService.GetUser(chatID)
	if err != nil || user == nil {
		logrus.WithField("chat_id", chatID).Warn("User not found for vacation")
		msg := tgbotapi.NewMessage(chatID, "❌ Профиль не найден.\nИспользуйте /createprofile чтобы создать профиль.")
		h.client.Bot.Send(msg)
		return
	}

	if args == "" {
		msg := tgbotapi.NewMessage(chatID,
			`🏖️ *Добавление отпуска*

Формат команды:
/vacation дата_начала дата_окончания

Примеры:
/vacation 01.07.2026 14.07.2026
→ Отпуск с 1 по 14 июля 2026

/vacation 15.08.2026 15.08.2026
→ Отпуск на один день 15 августа 2026

💡 *Важно:*
• Отпуск можно добавить только на будущие даты
• В выходные дни отпуск не добавляется
• Нельзя пересекаться с другими отпусками/больничными`)
		msg.ParseMode = "Markdown"
		h.client.Bot.Send(msg)
		return
	}

	// Парсим даты
	parts := strings.Fields(args)
	if len(parts) != 2 {
		msg := tgbotapi.NewMessage(chatID, "❌ Неверный формат. Используйте: /vacation дата_начала дата_окончания")
		h.client.Bot.Send(msg)
		return
	}

	startDate, err := parseDate(parts[0])
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка парсинга даты начала: "+err.Error())
		h.client.Bot.Send(msg)
		return
	}

	endDate, err := parseDate(parts[1])
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка парсинга даты окончания: "+err.Error())
		h.client.Bot.Send(msg)
		return
	}

	// Добавляем отпуск
	_, err = h.absenceService.AddVacation(uint(user.ID), startDate, endDate)
	if err != nil {
		logrus.WithError(err).Error("Failed to add vacation")
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка добавления отпуска: "+err.Error())
		h.client.Bot.Send(msg)
		return
	}

	// Подсчитываем количество дней
	days := int(endDate.Sub(startDate).Hours()/24) + 1
	
	response := fmt.Sprintf(
		`✅ Отпуск добавлен!

🏖️ Период: %s - %s
📅 Количество дней: %d
⏰ Всего часов: %d:%02d

📋 Дни отпуска будут автоматически засчитаны как рабочие (по 8ч 40м).`,
		startDate.Format("02.01.2006"),
		endDate.Format("02.01.2006"),
		days,
		8, 40, // 8 часов 40 минут
	)

	msg := tgbotapi.NewMessage(chatID, response)
	h.client.Bot.Send(msg)
}

// addSickLeave добавляет больничный
func (h *Handler) addSickLeave(message *tgbotapi.Message, args string) {
	chatID := message.Chat.ID

	// Получаем пользователя
	user, err := h.userService.GetUser(chatID)
	if err != nil || user == nil {
		logrus.WithField("chat_id", chatID).Warn("User not found for sick leave")
		msg := tgbotapi.NewMessage(chatID, "❌ Профиль не найден.\nИспользуйте /createprofile чтобы создать профиль.")
		h.client.Bot.Send(msg)
		return
	}

	if args == "" {
		msg := tgbotapi.NewMessage(chatID,
			`🏥 *Добавление больничного*

Формат команды:
/sick дата_начала дата_окончания

Примеры:
/sick 01.07.2026 07.07.2026
→ Больничный с 1 по 7 июля 2026

/sick 15.08.2026 15.08.2026
→ Больничный на один день 15 августа 2026

💡 *Важно:*
• Больничный можно добавить на любые даты (включая прошедшие)
• Можно добавлять на выходные дни
• Нельзя пересекаться с другими отпусками/больничными`)
		msg.ParseMode = "Markdown"
		h.client.Bot.Send(msg)
		return
	}

	// Парсим даты
	parts := strings.Fields(args)
	if len(parts) != 2 {
		msg := tgbotapi.NewMessage(chatID, "❌ Неверный формат. Используйте: /sick дата_начала дата_окончания")
		h.client.Bot.Send(msg)
		return
	}

	startDate, err := parseDate(parts[0])
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка парсинга даты начала: "+err.Error())
		h.client.Bot.Send(msg)
		return
	}

	endDate, err := parseDate(parts[1])
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка парсинга даты окончания: "+err.Error())
		h.client.Bot.Send(msg)
		return
	}

	// Добавляем больничный
	_, err = h.absenceService.AddSickLeave(uint(user.ID), startDate, endDate)
	if err != nil {
		logrus.WithError(err).Error("Failed to add sick leave")
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка добавления больничного: "+err.Error())
		h.client.Bot.Send(msg)
		return
	}

	// Подсчитываем количество дней
	days := int(endDate.Sub(startDate).Hours()/24) + 1
	
	response := fmt.Sprintf(
		`✅ Больничный добавлен!

🏥 Период: %s - %s
📅 Количество дней: %d
⏰ Всего часов: %d:%02d

📋 Дни больничного будут автоматически засчитаны как рабочие (по 8ч 40м).`,
		startDate.Format("02.01.2006"),
		endDate.Format("02.01.2006"),
		days,
		8, 40,
	)

	msg := tgbotapi.NewMessage(chatID, response)
	h.client.Bot.Send(msg)
}

// addDayOff добавляет отгул
func (h *Handler) addDayOff(message *tgbotapi.Message, args string) {
	chatID := message.Chat.ID

	// Получаем пользователя
	user, err := h.userService.GetUser(chatID)
	if err != nil || user == nil {
		logrus.WithField("chat_id", chatID).Warn("User not found for day off")
		msg := tgbotapi.NewMessage(chatID, "❌ Профиль не найден.\nИспользуйте /createprofile чтобы создать профиль.")
		h.client.Bot.Send(msg)
		return
	}

	if args == "" {
		msg := tgbotapi.NewMessage(chatID,
			`🎯 *Добавление отгула*

Формат команды:
/dayoff дата

Примеры:
/dayoff 01.07.2026
→ Отгул 1 июля 2026

/dayoff 15.08.2026
→ Отгул 15 августа 2026

💡 *Важно:*
• Отгул можно добавить на любые даты
• Нельзя добавить на выходной день
• Нельзя пересекаться с другими отпусками/больничными`)
		msg.ParseMode = "Markdown"
		h.client.Bot.Send(msg)
		return
	}

	// Парсим дату
	date, err := parseDate(args)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка парсинга даты: "+err.Error())
		h.client.Bot.Send(msg)
		return
	}

	// Добавляем отгул
	_, err = h.absenceService.AddDayOff(uint(user.ID), date)
	if err != nil {
		logrus.WithError(err).Error("Failed to add day off")
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка добавления отгула: "+err.Error())
		h.client.Bot.Send(msg)
		return
	}
	
	response := fmt.Sprintf(
		`✅ Отгул добавлен!

🎯 Дата: %s
⏰ Часы: 8:40

📋 Этот день будет засчитан как рабочий (8ч 40м).`,
		date.Format("02.01.2006"),
	)

	msg := tgbotapi.NewMessage(chatID, response)
	h.client.Bot.Send(msg)
}

// showMyAbsences показывает мои отпуска/больничные/отгулы
func (h *Handler) showMyAbsences(message *tgbotapi.Message, args string) {
	chatID := message.Chat.ID

	// Получаем пользователя
	user, err := h.userService.GetUser(chatID)
	if err != nil || user == nil {
		logrus.WithField("chat_id", chatID).Warn("User not found for absences")
		msg := tgbotapi.NewMessage(chatID, "❌ Профиль не найден.\nИспользуйте /createprofile чтобы создать профиль.")
		h.client.Bot.Send(msg)
		return
	}

	// Получаем все периоды отсутствия
	periods, err := h.absenceService.GetUserAbsences(uint(user.ID))
	if err != nil {
		logrus.WithError(err).Error("Failed to get user absences")
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка получения данных: "+err.Error())
		h.client.Bot.Send(msg)
		return
	}

	if len(periods) == 0 {
		msg := tgbotapi.NewMessage(chatID, "📭 У вас нет запланированных отпусков, больничных или отгулов.")
		h.client.Bot.Send(msg)
		return
	}

	// Группируем по типу
	vacations := []models.AbsencePeriod{}
	sickLeaves := []models.AbsencePeriod{}
	dayOffs := []models.AbsencePeriod{}
	
	for _, period := range periods {
		switch period.Type {
		case models.AbsenceTypeVacation:
			vacations = append(vacations, period)
		case models.AbsenceTypeSickLeave:
			sickLeaves = append(sickLeaves, period)
		case models.AbsenceTypeDayOff:
			dayOffs = append(dayOffs, period)
		}
	}

	response := "📋 *Мои периоды отсутствия:*\n\n"

	// Отпуска
	if len(vacations) > 0 {
		response += "🏖️ *Отпуска:*\n"
		for _, v := range vacations {
			days := int(v.EndDate.Sub(v.StartDate).Hours()/24) + 1
			response += fmt.Sprintf("• %s - %s (%d дней)\n", 
				v.StartDate.Format("02.01.2006"), 
				v.EndDate.Format("02.01.2006"),
				days)
		}
		response += "\n"
	}

	// Больничные
	if len(sickLeaves) > 0 {
		response += "🏥 *Больничные:*\n"
		for _, s := range sickLeaves {
			days := int(s.EndDate.Sub(s.StartDate).Hours()/24) + 1
			response += fmt.Sprintf("• %s - %s (%d дней)\n", 
				s.StartDate.Format("02.01.2006"), 
				s.EndDate.Format("02.01.2006"),
				days)
		}
		response += "\n"
	}

	// Отгулы
	if len(dayOffs) > 0 {
		response += "🎯 *Отгулы:*\n"
		for _, d := range dayOffs {
			response += fmt.Sprintf("• %s\n", d.StartDate.Format("02.01.2006"))
		}
	}

	// Подсчет статистики
	totalVacationDays := 0
	for _, v := range vacations {
		totalVacationDays += int(v.EndDate.Sub(v.StartDate).Hours()/24) + 1
	}

	totalSickDays := 0
	for _, s := range sickLeaves {
		totalSickDays += int(s.EndDate.Sub(s.StartDate).Hours()/24) + 1
	}

	response += "\n📊 *Статистика:*\n"
	response += fmt.Sprintf("• Всего отпускных дней: %d\n", totalVacationDays)
	response += fmt.Sprintf("• Всего больничных дней: %d\n", totalSickDays)
	response += fmt.Sprintf("• Всего отгулов: %d\n", len(dayOffs))

	msg := tgbotapi.NewMessage(chatID, response)
	msg.ParseMode = "Markdown"
	h.client.Bot.Send(msg)
}

// parseDate парсит дату из строки
func parseDate(dateStr string) (time.Time, error) {
	// Пробуем разные форматы
	formats := []string{
		"02.01.2006",
		"02-01-2006",
		"02.01",
		"02-01",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			// Если указан только день и месяц, добавляем текущий год
			if !strings.Contains(format, "2006") {
				now := time.Now()
				t = time.Date(now.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
			}
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("неверный формат даты. Используйте ДД.ММ.ГГГГ или ДД.ММ")
}