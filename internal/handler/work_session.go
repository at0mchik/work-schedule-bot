package handler

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
	"work-schedule-bot/internal/models"
)

// clockIn отмечает начало рабочего дня
func (h *Handler) clockIn(message *tgbotapi.Message) {
	chatID := message.Chat.ID
	now := time.Now()

	// Получаем пользователя
	user, err := h.userService.GetUser(chatID)
	if err != nil || user == nil {
		logrus.WithField("chat_id", chatID).Warn("User not found for clock in")
		msg := tgbotapi.NewMessage(chatID, "❌ Профиль не найден.\nИспользуйте /createprofile чтобы создать профиль.")
		h.client.Bot.Send(msg)
		return
	}

	// Проверяем, может ли пользователь начать работу
	canClockIn, reason, err := h.workSessionService.CanClockIn(user.ID)
	if err != nil {
		logrus.WithError(err).Error("Failed to check clock in eligibility")
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка проверки: "+err.Error())
		h.client.Bot.Send(msg)
		return
	}

	if !canClockIn {
		msg := tgbotapi.NewMessage(chatID, "❌ Не могу начать работу: "+reason)
		h.client.Bot.Send(msg)
		return
	}

	// Получаем необходимое время работы на сегодня

	requiredMinutes, err := h.userMonthlyStatService.GetRequiredMinutesByUserID(user.ID, time.Now().Year(), int(time.Now().Month()))
	if err != nil {
		logrus.WithError(err).Warn("Failed to get required minutes, using default")
		requiredMinutes = 200 // 8 часов 40 минут по умолчанию
	}

	// Начинаем работу
	_, err = h.workSessionService.ClockIn(user.ID, now, requiredMinutes)
	if err != nil {
		logrus.WithError(err).Error("Failed to clock in")
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка начала работы: "+err.Error())
		h.client.Bot.Send(msg)
		return
	}

	// Форматируем время
	inTime := now.Format("15:04")
	requiredHours := requiredMinutes / 60
	requiredMins := requiredMinutes % 60

	var requiredTime string
	if requiredMins == 0 {
		requiredTime = fmt.Sprintf("%d часов", requiredHours)
	} else {
		requiredTime = fmt.Sprintf("%d часов %d минут", requiredHours, requiredMins)
	}

	response := fmt.Sprintf(
		`✅ Рабочий день начат!

⏰ Время начала: %s
📅 Дата: %s
⏳ Норма на сегодня: %s

💡 Не забудьте отметить конец рабочего дня командой /out`,
		inTime,
		now.Format("02.01.2006"),
		requiredTime,
	)

	msg := tgbotapi.NewMessage(chatID, response)
	msg.ParseMode = "Markdown" // Включаем Markdown форматирование

	// Создаем клавиатуру с одной кнопкой
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"⏰ Завершить рабочий день", // Текст на кнопке
				"command_clock_out", // Callback data
			),
		),
	)

	// Настройки клавиатуры
	msg.ReplyMarkup = inlineKeyboard

	h.client.Bot.Send(msg)
}

// clockOut отмечает конец рабочего дня
func (h *Handler) clockOut(message *tgbotapi.Message) {
	chatID := message.Chat.ID
	now := time.Now()

	// Получаем пользователя
	user, err := h.userService.GetUser(chatID)
	if err != nil || user == nil {
		logrus.WithField("chat_id", chatID).Warn("User not found for clock out")
		msg := tgbotapi.NewMessage(chatID, "❌ Профиль не найден.\nИспользуйте /createprofile чтобы создать профиль.")
		h.client.Bot.Send(msg)
		return
	}

	// Проверяем, может ли пользователь закончить работу
	canClockOut, reason, err := h.workSessionService.CanClockOut(user.ID)
	if err != nil {
		logrus.WithError(err).Error("Failed to check clock out eligibility")
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка проверки: "+err.Error())
		h.client.Bot.Send(msg)
		return
	}

	if !canClockOut {
		msg := tgbotapi.NewMessage(chatID, "❌ Не могу закончить работу: "+reason)
		h.client.Bot.Send(msg)
		return
	}

	// Получаем текущую сессию перед завершением
	activeSession, err := h.workSessionService.GetActiveSession(user.ID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get active session")
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка получения сессии: "+err.Error())
		h.client.Bot.Send(msg)
		return
	}

	if activeSession == nil {
		msg := tgbotapi.NewMessage(chatID, "❌ Нет активной рабочей сессии")
		h.client.Bot.Send(msg)
		return
	}

	// Завершаем работу
	session, err := h.workSessionService.ClockOut(user.ID, now)
	if err != nil {
		logrus.WithError(err).Error("Failed to clock out")
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка завершения работы: "+err.Error())
		h.client.Bot.Send(msg)
		return
	}

	// Форматируем результат
	inTime := activeSession.ClockInTime.Format("15:04")
	outTime := now.Format("15:04")

	workedHours := session.WorkedMinutes / 60
	workedMins := session.WorkedMinutes % 60
	requiredHours := session.RequiredMinutes / 60
	requiredMins := session.RequiredMinutes % 60

	var workedTime, requiredTime string
	if workedMins == 0 {
		workedTime = fmt.Sprintf("%dч", workedHours)
	} else {
		workedTime = fmt.Sprintf("%dч %dм", workedHours, workedMins)
	}

	if requiredMins == 0 {
		requiredTime = fmt.Sprintf("%dч", requiredHours)
	} else {
		requiredTime = fmt.Sprintf("%dч %dм", requiredHours, requiredMins)
	}

	diffStatus := ""
	if session.DiffMinutes > 0 {
		diffHours := session.DiffMinutes / 60
		diffMins := session.DiffMinutes % 60
		if diffMins == 0 {
			diffStatus = fmt.Sprintf("\n\n➕ Переработка: %dч", diffHours)
		} else {
			diffStatus = fmt.Sprintf("\n\n➕ Переработка: %dч %dм", diffHours, diffMins)
		}
	} else if session.DiffMinutes < 0 {
		diffHours := -session.DiffMinutes / 60
		diffMins := -session.DiffMinutes % 60
		if diffMins == 0 {
			diffStatus = fmt.Sprintf("\n\n➖ Недобор: %dч", diffHours)
		} else {
			diffStatus = fmt.Sprintf("\n\n➖ Недобор: %dч %dм", diffHours, diffMins)
		}
	}

	response := fmt.Sprintf(
		`✅ Рабочий день завершен!

⏰ Время работы: %s - %s
⏳ Отработано: %s
📊 Норма: %s%s

📈 Статистика обновлена автоматически.`,
		inTime, outTime,
		workedTime,
		requiredTime,
		diffStatus,
	)

	msg := tgbotapi.NewMessage(chatID, response)
	msg.ParseMode = "Markdown"

	// Добавляем inline-кнопку для начала новой рабочей сессии
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"🔄 Начать новый рабочий день", // Текст на кнопке
				"command_clock_in", // Callback data для обработки
			),
		),
	)

	msg.ReplyMarkup = inlineKeyboard
	// Отправляем сообщение
	h.client.Bot.Send(msg)
}

// getTodayWorkSession показывает сегодняшнюю сессию
func (h *Handler) getTodayWorkSession(message *tgbotapi.Message) {
	chatID := message.Chat.ID

	// Получаем пользователя
	user, err := h.userService.GetUser(chatID)
	if err != nil || user == nil {
		logrus.WithField("chat_id", chatID).Warn("User not found for today session")
		msg := tgbotapi.NewMessage(chatID, "❌ Профиль не найден.\nИспользуйте /createprofile чтобы создать профиль.")
		h.client.Bot.Send(msg)
		return
	}

	// Получаем сегодняшнюю сессию
	sessions, err := h.workSessionService.GetAllTodaySessions(user.ID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get today's work session")
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка получения сессии: "+err.Error())
		h.client.Bot.Send(msg)
		return
	}

	if sessions == nil {
		msg := tgbotapi.NewMessage(chatID, "📭 Сегодня вы еще не начинали работу.\nИспользуйте /in чтобы начать рабочий день.")
		h.client.Bot.Send(msg)
		return
	}

	// Форматируем сессию
	var formated_all strings.Builder
	for _, session := range *sessions {
		formatted := h.workSessionService.FormatSession(&session)
		formated_all.WriteString("\n" + formatted)
	}
	msg := tgbotapi.NewMessage(chatID, formated_all.String())
	h.client.Bot.Send(msg)
}

// getWorkHistory показывает историю рабочих дней
func (h *Handler) getWorkHistory(message *tgbotapi.Message, args string) {
	chatID := message.Chat.ID

	// Получаем пользователя
	user, err := h.userService.GetUser(chatID)
	if err != nil || user == nil {
		logrus.WithField("chat_id", chatID).Warn("User not found for work history")
		msg := tgbotapi.NewMessage(chatID, "❌ Профиль не найден.\nИспользуйте /createprofile чтобы создать профиль.")
		h.client.Bot.Send(msg)
		return
	}

	limit := 10 // По умолчанию 10 последних записей
	if args != "" {
		parsedLimit, err := strconv.Atoi(args)
		if err == nil && parsedLimit > 0 && parsedLimit <= 50 {
			limit = parsedLimit
		}
	}

	// Получаем историю
	sessions, err := h.workSessionService.GetSessionHistory(user.ID, limit)
	if err != nil {
		logrus.WithError(err).Error("Failed to get work history")
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка получения истории: "+err.Error())
		h.client.Bot.Send(msg)
		return
	}

	// Форматируем результат
	formatted := h.workSessionService.FormatSessionList(sessions)
	msg := tgbotapi.NewMessage(chatID, formatted)
	h.client.Bot.Send(msg)
}

// getMonthWorkSessions показывает рабочие дни за месяц
func (h *Handler) getMonthWorkSessions(message *tgbotapi.Message, args string) {
	chatID := message.Chat.ID

	// Получаем пользователя
	user, err := h.userService.GetUser(chatID)
	if err != nil || user == nil {
		logrus.WithField("chat_id", chatID).Warn("User not found for month sessions")
		msg := tgbotapi.NewMessage(chatID, "❌ Профиль не найден.\nИспользуйте /createprofile чтобы создать профиль.")
		h.client.Bot.Send(msg)
		return
	}

	now := time.Now()
	year := now.Year()
	month := int(now.Month())

	if args != "" {
		parts := strings.Fields(args)
		if len(parts) == 1 {
			// Только месяц
			parsedMonth, err := strconv.Atoi(parts[0])
			if err != nil || parsedMonth < 1 || parsedMonth > 12 {
				msg := tgbotapi.NewMessage(chatID, "❌ Неверный месяц. Используйте число от 1 до 12.")
				h.client.Bot.Send(msg)
				return
			}
			month = parsedMonth
		} else if len(parts) == 2 {
			// Год и месяц
			parsedYear, err := strconv.Atoi(parts[0])
			if err != nil || parsedYear < 2000 || parsedYear > 2100 {
				msg := tgbotapi.NewMessage(chatID, "❌ Неверный год. Используйте год между 2000 и 2100.")
				h.client.Bot.Send(msg)
				return
			}
			year = parsedYear

			parsedMonth, err := strconv.Atoi(parts[1])
			if err != nil || parsedMonth < 1 || parsedMonth > 12 {
				msg := tgbotapi.NewMessage(chatID, "❌ Неверный месяц. Используйте число от 1 до 12.")
				h.client.Bot.Send(msg)
				return
			}
			month = parsedMonth
		}
	}

	// Получаем сессии за месяц
	sessions, err := h.workSessionService.GetMonthSessions(user.ID, year, month)
	if err != nil {
		logrus.WithError(err).Error("Failed to get month sessions")
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка получения сессий: "+err.Error())
		h.client.Bot.Send(msg)
		return
	}

	monthName := time.Month(month).String()
	if len(sessions) == 0 {
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("📭 В %s %d у вас не было рабочих дней.", monthName, year))
		h.client.Bot.Send(msg)
		return
	}

	// Подсчитываем статистику
	var totalMinutes, completedDays int
	for _, session := range sessions {
		if session.Status == models.StatusCompleted {
			completedDays++
			totalMinutes += session.WorkedMinutes
		}
	}

	totalHours := totalMinutes / 60
	totalMins := totalMinutes % 60

	var totalTime string
	if totalMins == 0 {
		totalTime = fmt.Sprintf("%dч", totalHours)
	} else {
		totalTime = fmt.Sprintf("%dч %dм", totalHours, totalMins)
	}

	response := fmt.Sprintf(
		`📊 Рабочие дни за %s %d

%s

📈 Итоги за месяц:
   📋 Отработано дней: %d
   ⏰ Всего времени: %s`,
		monthName, year,
		h.workSessionService.FormatSessionList(sessions),
		completedDays, totalTime,
	)

	msg := tgbotapi.NewMessage(chatID, response)
	h.client.Bot.Send(msg)
}

// getWorkStatus показывает текущий статус работы
func (h *Handler) getWorkStatus(message *tgbotapi.Message) {
	chatID := message.Chat.ID

	// Получаем пользователя
	user, err := h.userService.GetUser(chatID)
	if err != nil || user == nil {
		logrus.WithField("chat_id", chatID).Warn("User not found for work status")
		msg := tgbotapi.NewMessage(chatID, "❌ Профиль не найден.\nИспользуйте /createprofile чтобы создать профиль.")
		h.client.Bot.Send(msg)
		return
	}

	// Проверяем активную сессию
	activeSession, err := h.workSessionService.GetActiveSession(user.ID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get active session")
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка получения статуса: "+err.Error())
		h.client.Bot.Send(msg)
		return
	}

	if activeSession != nil {
		// Пользователь на работе
		inTime := activeSession.ClockInTime.Format("15:04")
		duration := time.Since(activeSession.ClockInTime)
		hours := int(duration.Hours())
		minutes := int(duration.Minutes()) % 60

		var durationStr string
		if minutes == 0 {
			durationStr = fmt.Sprintf("%dч", hours)
		} else {
			durationStr = fmt.Sprintf("%dч %dм", hours, minutes)
		}

		response := fmt.Sprintf(
			`🟢 Вы на работе!

⏰ Начали работу: %s
⏳ Прошло времени: %s
📅 Дата: %s

💡 Используйте /out чтобы завершить рабочий день.`,
			inTime,
			durationStr,
			activeSession.Date.Format("02.01.2006"),
		)

		msg := tgbotapi.NewMessage(chatID, response)
		h.client.Bot.Send(msg)
		return
	}

	// Проверяем сегодняшнюю сессию
	todaySession, err := h.workSessionService.GetTodaySession(user.ID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get today's session")
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка получения статуса: "+err.Error())
		h.client.Bot.Send(msg)
		return
	}

	if todaySession != nil && todaySession.Status == models.StatusCompleted {
		// Рабочий день завершен
		formatted := h.workSessionService.FormatSession(todaySession)
		msg := tgbotapi.NewMessage(chatID, formatted)
		h.client.Bot.Send(msg)
		return
	}

	// Нет активной сессии и сегодняшней завершенной
	msg := tgbotapi.NewMessage(chatID,
		`📭 Сегодня вы еще не работали.

💡 Доступные команды:
/in - Начать рабочий день
/today - Информация о сегодняшнем дне
/history - История рабочих дней
/status - Текущий статус`)
	h.client.Bot.Send(msg)
}
