package handler

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

func (h *Handler) sendEchoMessage(message *tgbotapi.Message) {
	responseText := message.Text

	if responseText == "" {
		responseText = "Я получил ваше сообщение, но не могу его повторить 😊"
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, "🔁 Эхо: "+responseText)
	h.client.Bot.Send(msg)
}

// internal/bot/handler/commands.go
func (h *Handler) handleCommand(message *tgbotapi.Message) {
	command := message.Command()
	args := message.CommandArguments()

	switch command {
	case "start":
		h.sendStartMessage(message)
	case "help":
		h.sendHelpMessage(message)
	case "helpadmin":
		h.sendAdminHelpMessage(message)
	case "createprofile":
		h.startProfileCreation(message)
	case "myprofile":
		h.showProfile(message)
	case "updateprofile":
		h.startProfileUpdate(message)
	case "deleteprofile":
		h.deleteProfile(message)
	case "allusers":
		h.showAllUsers(message)
	case "stats":
		h.showStats(message)
	case "setrole":
		h.setUserRole(message, args)
	case "promote":
		h.promoteToAdmin(message, args)
	case "demote":
		h.demoteToClient(message, args)
	case "admins":
		h.showAdmins(message)
	case "echo":
		h.sendEchoWithArgs(message, args)

	// Команды для работы с графиками (админы)
	case "addschedule":
		h.addWorkSchedule(message, args)
	case "updateschedule":
		h.updateWorkSchedule(message, args)
	case "deleteschedule":
		h.deleteWorkSchedule(message, args)
	case "getschedules":
		h.getWorkSchedules(message)
	case "getschedule":
		h.getWorkSchedule(message, args)
	case "currentschedule":
		h.getCurrentSchedule(message)
	case "generateschedules":
		h.generateSchedules(message, args)
	case "updateallschedules":
		h.updateAllSchedules(message)
	case "checkday":
		h.checkWorkingDay(message, args)		
		
	// Команды для статистики (все пользователи)
	case "mystats":
		h.getMyMonthlyStats(message)
	case "stat":
		h.getMonthlyStat(message, args)
	case "currentstat":
		h.getCurrentMonthStat(message)

	// Команды для работы (все пользователи)
	case "in", "startwork":
		h.clockIn(message)
	case "out", "endwork", "finish":
		h.clockOut(message)
	case "helptime":
		h.showTimeFormatsHelp(message)
	case "today":
		h.getTodayWorkSession(message)
	case "history":
		h.getWorkHistory(message, args)
	case "monthwork":
		h.getMonthWorkSessions(message, args)
	case "status":
		h.getWorkStatus(message)

	// Команды для отпусков/больничных/отгулов (все пользователи) // ДОБАВЛЕНО
	case "vacation":
		h.addVacation(message, args)
	case "sick", "sickleave":
		h.addSickLeave(message, args)
	case "dayoff":
		h.addDayOff(message, args)
	case "myabsences":
		h.showMyAbsences(message, args)

	default:
		h.sendUnknownCommand(message)
	}
}

func (h *Handler) sendUnknownCommand(message *tgbotapi.Message) {
	msg := tgbotapi.NewMessage(message.Chat.ID, "❌ Неизвестная команда. Используйте /help для списка команд.")
	h.client.Bot.Send(msg)
}

func (h *Handler) sendEchoWithArgs(message *tgbotapi.Message, args string) {
	if strings.TrimSpace(args) == "" {
		args = "Вы не указали текст для эхо!"
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, "📢: "+args)
	h.client.Bot.Send(msg)
}

// В sendStartMessage добавляем команды для админов:
func (h *Handler) sendStartMessage(message *tgbotapi.Message) {
	chatID := message.Chat.ID

	text := `📋 Доступные команды:

👤 Профиль:
/createprofile - Создать профиль (ФИО)
/myprofile - Показать мой профиль
/updateprofile - Обновить профиль
/deleteprofile - Удалить профиль

⏰ Учет рабочего времени:
/in - Начать рабочий день (можно указать дату и время, подробности /helptime)
/out - Завершить рабочий день (можно указать дату и время, подробности /helptime)
/helptime - Показать справку по указанию времени для начача\конца рабочего дня
/today - Информация о сегодняшнем рабочем дне
/status - Текущий статус работы
/history [N] - История рабочих дней (последние N, по умолчанию 10)
/monthwork [месяц] - Рабочие дни за конкретный месяц
/monthwork [год месяц] - Рабочие дни за месяц и год
/checkday [дата] - Проверить, является ли день рабочим
    Пример: /checkday 01.05.2026 или /checkday 01.05

🏖️ Отпуска/Больничные/Отгулы:
/vacation дата_начала дата_окончания - Добавить отпуск
    Пример: /vacation 01.07.2026 14.07.2026
/sick дата_начала дата_окончания - Добавить больничный
    Пример: /sick 01.07.2026 07.07.2026
/dayoff дата - Добавить отгул
    Пример: /dayoff 15.08.2026
/myabsences - Мои периоды отсутствия

📊 Статистика работы:
/mystats - Вся моя статистика
/stat [месяц] - Статистика за конкретный месяц
/stat [год месяц] - Статистика за месяц и год
/currentstat - Статистика за текущий месяц

🛠 Утилиты:
/start - Начать работу с ботом
/help - Показать это сообщение
/echo [текст] - Отправить эхо с текстом

💡 Как пользоваться:
1. Создайте профиль командой /createprofile
2. Начинайте рабочий день командой /in
3. Завершайте рабочий день командой /out
4. Добавляйте отпуска/больничные командой /vacation или /sick
5. Следите за статистикой командой /mystats
6. Администраторы настраивают графики работы

📈 Автоматические обновления:
• При завершении рабочего дня статистика обновляется автоматически
• При добавлении отпуска/больничного дни засчитываются как отработанные`


	msg := tgbotapi.NewMessage(chatID, text)
	h.client.Bot.Send(msg)
}

func (h *Handler) sendHelpMessage(message *tgbotapi.Message) {
	chatID := message.Chat.ID

	text := `📋 Доступные команды:

👤 Профиль:
/createprofile - Создать профиль (ФИО)
/myprofile - Показать мой профиль
/updateprofile - Обновить профиль
/deleteprofile - Удалить профиль

⏰ Учет рабочего времени:
/in - Начать рабочий день (можно указать дату и время, подробности /helptime)
/out - Завершить рабочий день (можно указать дату и время, подробности /helptime)
/helptime - Показать справку по указанию времени для начача\конца рабочего дня
/today - Информация о сегодняшнем рабочем дне
/status - Текущий статус работы
/history [N] - История рабочих дней (последние N, по умолчанию 10)
/monthwork [месяц] - Рабочие дни за конкретный месяц
/monthwork [год месяц] - Рабочие дни за месяц и год
/checkday [дата] - Проверить, является ли день рабочим
    Пример: /checkday 01.05.2026 или /checkday 01.05

🏖️ Отпуска/Больничные/Отгулы:
/vacation дата_начала дата_окончания - Добавить отпуск
    Пример: /vacation 01.07.2026 14.07.2026
/sick дата_начала дата_окончания - Добавить больничный
    Пример: /sick 01.07.2026 07.07.2026
/dayoff дата - Добавить отгул
    Пример: /dayoff 15.08.2026
/myabsences - Мои периоды отсутствия

📊 Статистика работы:
/mystats - Вся моя статистика
/stat [месяц] - Статистика за конкретный месяц
/stat [год месяц] - Статистика за месяц и год
/currentstat - Статистика за текущий месяц

🛠 Утилиты:
/start - Начать работу с ботом
/help - Показать это сообщение
/echo [текст] - Отправить эхо с текстом

💡 Как пользоваться:
1. Создайте профиль командой /createprofile
2. Начинайте рабочий день командой /in
3. Завершайте рабочий день командой /out
4. Добавляйте отпуска/больничные командой /vacation или /sick
5. Следите за статистикой командой /mystats
6. Администраторы настраивают графики работы

📈 Автоматические обновления:
• При завершении рабочего дня статистика обновляется автоматически
• При добавлении отпуска/больничного дни засчитываются как отработанные`

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	h.client.Bot.Send(msg)
}

func (h *Handler) sendAdminHelpMessage(message *tgbotapi.Message){
	chatID := message.Chat.ID

	// Проверяем, является ли пользователь админом
	isAdmin, err := h.userService.IsAdmin(chatID)
	if err != nil {
		logrus.WithError(err).Error("Error checking admin status")
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка проверки прав доступа: "+err.Error())
		h.client.Bot.Send(msg)
		return
	}

	if !isAdmin {
		logrus.WithField("chat_id", chatID).Warn("Unauthorized access to admin help commands command")
		msg := tgbotapi.NewMessage(chatID, "❌ Доступ запрещен. Эта команда только для администраторов.")
		h.client.Bot.Send(msg)
		return
	}

	text := `📋 Доступные команды:

👑 Администрирование:
/allusers - Показать всех пользователей
/stats - Статистика бота
/admins - Показать администраторов
/promote [ID] - Назначить администратора
/demote [ID] - Снять администратора
/setrole [ID] [role] - Изменить роль

📅 Управление графиками:
/addschedule [год месяц дни минуты] - Добавить график
Пример: /addschedule 2024 12 22 480

/updateschedule [ID дни минуты] - Обновить график
Пример: /updateschedule 1 23 490

/deleteschedule [ID] - Удалить график
/getschedules - Все графики работы
/getschedule [ID] - Показать график по ID
/getschedule [год месяц] - Показать график по дате
/currentschedule - График на текущий месяц
/generateschedules [год] [минуты] - Автоматически создать графики на год
    Пример: /generateschedules 2026 520 (8ч 40м)
/updateallschedules - Обновить все графики на основе выходных дней`

	if h.config.BaseAdminChatID != 0 {
		text += fmt.Sprintf("\n\n🔧 ID главного администратора: %d", h.config.BaseAdminChatID)
	}
	
	msg := tgbotapi.NewMessage(chatID, text)
	h.client.Bot.Send(msg)
}


func (h *Handler) showTimeFormatsHelp(message *tgbotapi.Message) {
	chatID := message.Chat.ID
	helpText := `📝 *Форматы указания даты и времени:*

*Дата (необязательно):*
• dd.mm.yyyy (25.12.2023)
• dd-mm-yyyy (25-12-2023)

*Время (необязательно):*
• hh:mm (09:30)
• hh.mm (09.30)
• hh-mm (09-30)

*Примеры использования:*
• /in 09:00 — начать работу сегодня в 9:00
• /in 25.12.2023 09:30 — начать работу 25 декабря 2023 в 9:30
• /out 18:00 — завершить работу сегодня в 18:00
• /out 25-12-2023 18-30 — завершить работу 25 декабря 2023 в 18:30
• /in — начать работу сейчас
• /out — завершить работу сейчас

⚠️ *Примечания:*
• Если указать только время, будет использована сегодняшняя дата
• *Нельзя начать работу в выходной день* согласно производственному календарю
• Используйте /checkday [дата] чтобы проверить, является ли день рабочим
• Год должен быть 2026 или позже`

	msg := tgbotapi.NewMessage(chatID, helpText)
	msg.ParseMode = "Markdown"
	h.client.Bot.Send(msg)
}