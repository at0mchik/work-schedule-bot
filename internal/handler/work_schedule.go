// internal/bot/handler/work_schedule_handler.go
package handler

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

// addWorkSchedule добавляет новый график работы
func (h *Handler) addWorkSchedule(message *tgbotapi.Message, args string) {
	chatID := message.Chat.ID

	// Проверяем права доступа (только админы)
	isAdmin, err := h.userService.IsAdmin(chatID)
	if err != nil {
		logrus.WithError(err).Error("Error checking admin status")
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка проверки прав доступа: "+err.Error())
		h.client.Bot.Send(msg)
		return
	}

	if !isAdmin {
		logrus.WithField("chat_id", chatID).Warn("Unauthorized access to addschedule command")
		msg := tgbotapi.NewMessage(chatID, "❌ Доступ запрещен. Эта команда только для администраторов.")
		h.client.Bot.Send(msg)
		return
	}

	if args == "" {
		// Показываем инструкцию по формату
		msg := tgbotapi.NewMessage(chatID,
			`📝 Добавление графика работы

Формат команды:
/addschedule Год Месяц Дни МинутыВДень

Пример:
/addschedule 2024 12 22 480
→ Декабрь 2024, 22 рабочих дня по 8 часов (480 минут)

/addschedule 2024 1 20 450
→ Январь 2024, 20 рабочих дней по 7.5 часов (450 минут = 7ч 30м)

Или просто отправьте данные в формате:
"2024 12 22 480"`)
		h.client.Bot.Send(msg)
		return
	}

	// Парсим данные
	year, month, workDays, workMinutesPerDay, err := h.workScheduleService.ParseScheduleData(args)
	if err != nil {
		logrus.WithError(err).Warn("Failed to parse schedule data")
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка парсинга данных: "+err.Error())
		h.client.Bot.Send(msg)
		return
	}

	// Создаем график
	schedule, err := h.workScheduleService.CreateSchedule(year, month, workDays, workMinutesPerDay)
	if err != nil {
		logrus.WithError(err).Error("Failed to create work schedule")
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка создания графика: "+err.Error())
		h.client.Bot.Send(msg)
		return
	}

	// Форматируем результат
	formatted := h.workScheduleService.FormatSchedule(schedule)
	msg := tgbotapi.NewMessage(chatID, formatted)
	h.client.Bot.Send(msg)
}

// updateWorkSchedule обновляет существующий график
func (h *Handler) updateWorkSchedule(message *tgbotapi.Message, args string) {
	chatID := message.Chat.ID

	// Проверяем права доступа
	isAdmin, err := h.userService.IsAdmin(chatID)
	if err != nil {
		logrus.WithError(err).Error("Error checking admin status")
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка проверки прав доступа: "+err.Error())
		h.client.Bot.Send(msg)
		return
	}

	if !isAdmin {
		logrus.WithField("chat_id", chatID).Warn("Unauthorized access to updateschedule command")
		msg := tgbotapi.NewMessage(chatID, "❌ Доступ запрещен. Эта команда только для администраторов.")
		h.client.Bot.Send(msg)
		return
	}

	if args == "" {
		msg := tgbotapi.NewMessage(chatID,
			`✏️ Обновление графика работы

Формат команды:
/updateschedule ID Дни МинутыВДень

Пример:
/updateschedule 1 23 490
→ Обновит график с ID=1 на 23 рабочих дня по 490 минут (8ч 10м)

Сначала используйте /getschedules чтобы увидеть ID графиков`)
		h.client.Bot.Send(msg)
		return
	}

	// Парсим данные
	parts := strings.Fields(args)
	if len(parts) != 3 {
		msg := tgbotapi.NewMessage(chatID, "❌ Неверный формат. Используйте: /updateschedule ID дни минуты_в_день")
		h.client.Bot.Send(msg)
		return
	}

	// Парсим ID
	id, err := strconv.ParseUint(parts[0], 10, 32)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ Неверный формат ID. ID должен быть числом.")
		h.client.Bot.Send(msg)
		return
	}

	// Парсим рабочие дни
	workDays, err := strconv.Atoi(parts[1])
	if err != nil || workDays < 0 || workDays > 31 {
		msg := tgbotapi.NewMessage(chatID, "❌ Неверное количество дней. Должно быть между 0 и 31.")
		h.client.Bot.Send(msg)
		return
	}

	// Парсим минуты в день
	workMinutesPerDay, err := strconv.Atoi(parts[2])
	if err != nil || workMinutesPerDay <= 0 || workMinutesPerDay > 1440 {
		msg := tgbotapi.NewMessage(chatID, "❌ Неверное количество минут в день. Должно быть между 1 и 1440.")
		h.client.Bot.Send(msg)
		return
	}

	// Обновляем график
	schedule, err := h.workScheduleService.UpdateSchedule(uint(id), workDays, workMinutesPerDay)
	if err != nil {
		logrus.WithError(err).Error("Failed to update work schedule")
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка обновления графика: "+err.Error())
		h.client.Bot.Send(msg)
		return
	}

	formatted := h.workScheduleService.FormatSchedule(schedule)
	msg := tgbotapi.NewMessage(chatID, formatted)
	h.client.Bot.Send(msg)
}

// deleteWorkSchedule удаляет график
func (h *Handler) deleteWorkSchedule(message *tgbotapi.Message, args string) {
	chatID := message.Chat.ID

	// Проверяем права доступа
	isAdmin, err := h.userService.IsAdmin(chatID)
	if err != nil {
		logrus.WithError(err).Error("Error checking admin status")
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка проверки прав доступа: "+err.Error())
		h.client.Bot.Send(msg)
		return
	}

	if !isAdmin {
		logrus.WithField("chat_id", chatID).Warn("Unauthorized access to deleteschedule command")
		msg := tgbotapi.NewMessage(chatID, "❌ Доступ запрещен. Эта команда только для администраторов.")
		h.client.Bot.Send(msg)
		return
	}

	if args == "" {
		msg := tgbotapi.NewMessage(chatID,
			`🗑️ Удаление графика работы

Формат команды:
/deleteschedule ID

Пример:
/deleteschedule 1
→ Удалит график с ID=1

Сначала используйте /getschedules чтобы увидеть ID графиков`)
		h.client.Bot.Send(msg)
		return
	}

	// Парсим ID
	id, err := strconv.ParseUint(args, 10, 32)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ Неверный формат ID. ID должен быть числом.")
		h.client.Bot.Send(msg)
		return
	}

	// Создаем inline клавиатуру для подтверждения
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Да, удалить", fmt.Sprintf("confirm_delete_schedule_%d", id)),
			tgbotapi.NewInlineKeyboardButtonData("❌ Нет, отменить", "cancel_delete_schedule"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("⚠️ Вы уверены, что хотите удалить график с ID %d?\nЭто действие нельзя отменить.", id))
	msg.ReplyMarkup = keyboard
	h.client.Bot.Send(msg)
}

// getWorkSchedules показывает все графики
func (h *Handler) getWorkSchedules(message *tgbotapi.Message) {
	chatID := message.Chat.ID

	// Проверяем права доступа
	isAdmin, err := h.userService.IsAdmin(chatID)
	if err != nil {
		logrus.WithError(err).Error("Error checking admin status")
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка проверки прав доступа: "+err.Error())
		h.client.Bot.Send(msg)
		return
	}

	if !isAdmin {
		logrus.WithField("chat_id", chatID).Warn("Unauthorized access to getschedules command")
		msg := tgbotapi.NewMessage(chatID, "❌ Доступ запрещен. Эта команда только для администраторов.")
		h.client.Bot.Send(msg)
		return
	}

	// Получаем все графики
	schedules, err := h.workScheduleService.GetAllSchedules()
	if err != nil {
		logrus.WithError(err).Error("Failed to get work schedules")
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка получения графиков: "+err.Error())
		h.client.Bot.Send(msg)
		return
	}

	formatted := h.workScheduleService.FormatScheduleList(schedules)
	msg := tgbotapi.NewMessage(chatID, formatted)
	h.client.Bot.Send(msg)
}

// getWorkSchedule показывает конкретный график
func (h *Handler) getWorkSchedule(message *tgbotapi.Message, args string) {
	chatID := message.Chat.ID

	// Проверяем права доступа
	isAdmin, err := h.userService.IsAdmin(chatID)
	if err != nil {
		logrus.WithError(err).Error("Error checking admin status")
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка проверки прав доступа: "+err.Error())
		h.client.Bot.Send(msg)
		return
	}

	if !isAdmin {
		logrus.WithField("chat_id", chatID).Warn("Unauthorized access to getschedule command")
		msg := tgbotapi.NewMessage(chatID, "❌ Доступ запрещен. Эта команда только для администраторов.")
		h.client.Bot.Send(msg)
		return
	}

	if args == "" {
		msg := tgbotapi.NewMessage(chatID,
			`📋 Просмотр графика работы

Формат команды:
/getschedule ID
→ Покажет график с указанным ID

/getschedule 2024 12
→ Покажет график на декабрь 2024 года

Используйте /getschedules чтобы увидеть все доступные графики`)
		h.client.Bot.Send(msg)
		return
	}

	// Пробуем распарсить как ID
	if id, err := strconv.ParseUint(args, 10, 32); err == nil {
		// Это ID
		schedule, err := h.workScheduleService.GetScheduleByID(uint(id))
		if err != nil {
			logrus.WithError(err).Error("Failed to get work schedule by ID")
			msg := tgbotapi.NewMessage(chatID, "❌ Ошибка получения графика: "+err.Error())
			h.client.Bot.Send(msg)
			return
		}

		if schedule == nil {
			msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("❌ График с ID %d не найден.", id))
			h.client.Bot.Send(msg)
			return
		}

		formatted := h.workScheduleService.FormatSchedule(schedule)
		msg := tgbotapi.NewMessage(chatID, formatted)
		h.client.Bot.Send(msg)
		return
	}

	// Пробуем распарсить как год и месяц
	parts := strings.Fields(args)
	if len(parts) == 2 {
		year, err1 := strconv.Atoi(parts[0])
		month, err2 := strconv.Atoi(parts[1])

		if err1 == nil && err2 == nil && month >= 1 && month <= 12 {
			schedule, err := h.workScheduleService.GetScheduleByYearMonth(year, month)
			if err != nil {
				logrus.WithError(err).Error("Failed to get work schedule by year/month")
				msg := tgbotapi.NewMessage(chatID, "❌ Ошибка получения графика: "+err.Error())
				h.client.Bot.Send(msg)
				return
			}

			if schedule == nil {
				monthName := time.Month(month).String()
				msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("❌ График на %s %d не найден.", monthName, year))
				h.client.Bot.Send(msg)
				return
			}

			formatted := h.workScheduleService.FormatSchedule(schedule)
			msg := tgbotapi.NewMessage(chatID, formatted)
			h.client.Bot.Send(msg)
			return
		}
	}

	msg := tgbotapi.NewMessage(chatID, "❌ Неверный формат. Используйте: /getschedule ID или /getschedule год месяц")
	h.client.Bot.Send(msg)
}

// getCurrentSchedule показывает график на текущий месяц
func (h *Handler) getCurrentSchedule(message *tgbotapi.Message) {
	chatID := message.Chat.ID

	// Проверяем права доступа
	isAdmin, err := h.userService.IsAdmin(chatID)
	if err != nil {
		logrus.WithError(err).Error("Error checking admin status")
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка проверки прав доступа: "+err.Error())
		h.client.Bot.Send(msg)
		return
	}

	if !isAdmin {
		logrus.WithField("chat_id", chatID).Warn("Unauthorized access to currentschedule command")
		msg := tgbotapi.NewMessage(chatID, "❌ Доступ запрещен. Эта команда только для администраторов.")
		h.client.Bot.Send(msg)
		return
	}

	// Получаем график на текущий месяц
	schedule, err := h.workScheduleService.GetCurrentSchedule()
	if err != nil {
		logrus.WithError(err).Error("Failed to get current schedule")
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка получения текущего графика: "+err.Error())
		h.client.Bot.Send(msg)
		return
	}

	if schedule == nil {
		now := time.Now()
		monthName := now.Month().String()
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("❌ График на текущий месяц (%s %d) не установлен.", monthName, now.Year()))
		h.client.Bot.Send(msg)
		return
	}

	formatted := h.workScheduleService.FormatSchedule(schedule)
	msg := tgbotapi.NewMessage(chatID, formatted)
	h.client.Bot.Send(msg)
}

// handleScheduleCallback обрабатывает callback для графиков (добавить в существующий handleCallbackQuery)
func (h *Handler) handleScheduleCallback(callback *tgbotapi.CallbackQuery) {
	chatID := callback.Message.Chat.ID
	data := callback.Data

	// Удаляем клавиатуру
	editMsg := tgbotapi.NewEditMessageReplyMarkup(chatID, callback.Message.MessageID, tgbotapi.NewInlineKeyboardMarkup())
	h.client.Bot.Send(editMsg)

	// Обработка подтверждения удаления графика
	if strings.HasPrefix(data, "confirm_delete_schedule_") {
		// Извлекаем ID графика
		idStr := strings.TrimPrefix(data, "confirm_delete_schedule_")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			msg := tgbotapi.NewMessage(chatID, "❌ Ошибка: неверный ID графика")
			h.client.Bot.Send(msg)
			return
		}

		// Удаляем график
		err = h.workScheduleService.DeleteSchedule(uint(id))
		if err != nil {
			logrus.WithError(err).Error("Failed to delete work schedule via callback")
			msg := tgbotapi.NewMessage(chatID, "❌ Ошибка удаления графика: "+err.Error())
			h.client.Bot.Send(msg)
		} else {
			msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("✅ График с ID %d успешно удален!", id))
			h.client.Bot.Send(msg)
		}
	} else if data == "cancel_delete_schedule" {
		msg := tgbotapi.NewMessage(chatID, "❌ Удаление графика отменено.")
		h.client.Bot.Send(msg)
	}

	// Отвечаем на callback
	callbackConfig := tgbotapi.NewCallback(callback.ID, "")
	h.client.Bot.Send(callbackConfig)
}

func (h *Handler) generateSchedules(message *tgbotapi.Message, args string) {
	chatID := message.Chat.ID

	// Проверяем права доступа (только админы)
	isAdmin, err := h.userService.IsAdmin(chatID)
	if err != nil {
		logrus.WithError(err).Error("Error checking admin status")
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка проверки прав доступа: "+err.Error())
		h.client.Bot.Send(msg)
		return
	}

	if !isAdmin {
		logrus.WithField("chat_id", chatID).Warn("Unauthorized access to generateschedules command")
		msg := tgbotapi.NewMessage(chatID, "❌ Доступ запрещен. Эта команда только для администраторов.")
		h.client.Bot.Send(msg)
		return
	}

	var year int
	var workMinutesPerDay int = 480 // 8 часов по умолчанию

	if args == "" {
		// Используем текущий год и дефолтное время
		year = time.Now().Year()
	} else {
		// Парсим аргументы
		parts := strings.Fields(args)
		if len(parts) == 1 {
			// Только год
			parsedYear, err := strconv.Atoi(parts[0])
			if err != nil {
				msg := tgbotapi.NewMessage(chatID, "❌ Неверный формат года. Используйте: /generateschedules [год] [минуты_в_день]")
				h.client.Bot.Send(msg)
				return
			}
			year = parsedYear
		} else if len(parts) == 2 {
			// Год и минуты в день
			parsedYear, err1 := strconv.Atoi(parts[0])
			parsedMinutes, err2 := strconv.Atoi(parts[1])
			if err1 != nil || err2 != nil {
				msg := tgbotapi.NewMessage(chatID, "❌ Неверный формат. Используйте: /generateschedules [год] [минуты_в_день]")
				h.client.Bot.Send(msg)
				return
			}
			year = parsedYear
			workMinutesPerDay = parsedMinutes
		} else {
			msg := tgbotapi.NewMessage(chatID, "❌ Неверный формат. Используйте: /generateschedules [год] [минуты_в_день]")
			h.client.Bot.Send(msg)
			return
		}
	}

	// Проверяем корректность года
	if year < 2000 || year > 2100 {
		msg := tgbotapi.NewMessage(chatID, "❌ Неверный год. Год должен быть между 2000 и 2100.")
		h.client.Bot.Send(msg)
		return
	}

	// Проверяем корректность минут в день
	if workMinutesPerDay <= 0 || workMinutesPerDay > 1440 {
		msg := tgbotapi.NewMessage(chatID, "❌ Неверное количество минут в день. Должно быть между 1 и 1440.")
		h.client.Bot.Send(msg)
		return
	}

	// Генерируем графики
	schedules, err := h.workScheduleService.GenerateSchedulesFromNonWorkingDays(year, workMinutesPerDay)
	if err != nil {
		logrus.WithError(err).Error("Failed to generate schedules")
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка генерации графиков: "+err.Error())
		h.client.Bot.Send(msg)
		return
	}

	// Форматируем результат
	response := fmt.Sprintf("✅ Сгенерировано %d графиков на %d год\n\n", len(schedules), year)
	response += "📋 Список созданных/обновленных графиков:\n\n"

	for i, schedule := range schedules {
		hours := schedule.WorkMinutesPerDay / 60
		minutes := schedule.WorkMinutesPerDay % 60
		monthName := time.Month(schedule.Month).String()
		
		response += fmt.Sprintf("%d. %s %d: %d рабочих дней × %d:%02d часов\n",
			i+1, monthName, schedule.Year, schedule.WorkDays, hours, minutes)
	}

	msg := tgbotapi.NewMessage(chatID, response)
	h.client.Bot.Send(msg)
}

// updateAllSchedules обновляет все существующие графики на основе выходных дней
func (h *Handler) updateAllSchedules(message *tgbotapi.Message) {
	chatID := message.Chat.ID

	// Проверяем права доступа
	isAdmin, err := h.userService.IsAdmin(chatID)
	if err != nil {
		logrus.WithError(err).Error("Error checking admin status")
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка проверки прав доступа: "+err.Error())
		h.client.Bot.Send(msg)
		return
	}

	if !isAdmin {
		logrus.WithField("chat_id", chatID).Warn("Unauthorized access to updateallschedules command")
		msg := tgbotapi.NewMessage(chatID, "❌ Доступ запрещен. Эта команда только для администраторов.")
		h.client.Bot.Send(msg)
		return
	}

	// Обновляем все графики
	updatedCount, err := h.workScheduleService.UpdateAllSchedulesFromNonWorkingDays()
	if err != nil {
		logrus.WithError(err).Error("Failed to update all schedules")
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка обновления графиков: "+err.Error())
		h.client.Bot.Send(msg)
		return
	}

	if updatedCount == 0 {
		msg := tgbotapi.NewMessage(chatID, "✅ Все графики уже актуальны. Ничего не обновлено.")
		h.client.Bot.Send(msg)
	} else {
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("✅ Обновлено %d графиков.", updatedCount))
		h.client.Bot.Send(msg)
	}
}

// checkWorkingDay проверяет, является ли день рабочим
func (h *Handler) checkWorkingDay(message *tgbotapi.Message, args string) {
	chatID := message.Chat.ID

	var date time.Time
	if args == "" {
		// Используем сегодняшнюю дату
		date = time.Now()
	} else {
		// Парсим дату из аргументов
		parsedDate, err := time.Parse("02.01.2006", args)
		if err != nil {
			// Пробуем другой формат
			parsedDate, err = time.Parse("02.01", args)
			if err != nil {
				msg := tgbotapi.NewMessage(chatID, "❌ Неверный формат даты. Используйте ДД.ММ.ГГГГ или ДД.ММ")
				h.client.Bot.Send(msg)
				return
			}
			// Устанавливаем текущий год
			parsedDate = time.Date(time.Now().Year(), parsedDate.Month(), parsedDate.Day(), 0, 0, 0, 0, time.Local)
		}
		date = parsedDate
	}

	// Проверяем, является ли день рабочим
	isWorking, err := h.workScheduleService.IsWorkingDay(date)
	if err != nil {
		logrus.WithError(err).Error("Failed to check if day is working")
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка проверки дня: "+err.Error())
		h.client.Bot.Send(msg)
		return
	}

	// Получаем количество рабочих минут для этого дня
	workMinutes, err := h.workScheduleService.GetWorkMinutesForDay(date)
	if err != nil {
		logrus.WithError(err).Warn("Failed to get work minutes for day")
	}

	response := fmt.Sprintf("📅 Дата: %s\n", date.Format("02.01.2006"))
	
	if isWorking {
		hours := workMinutes / 60
		minutes := workMinutes % 60
		response += fmt.Sprintf("✅ Рабочий день\n⏰ Время работы: %d:%02d часов", hours, minutes)
	} else {
		response += "❌ Выходной день"
	}

	msg := tgbotapi.NewMessage(chatID, response)
	h.client.Bot.Send(msg)
}