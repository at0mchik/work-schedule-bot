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

func (h *Handler) handleCommand(message *tgbotapi.Message) {
	command := message.Command()
	args := message.CommandArguments()

	switch command {
	case "start":
		h.sendStartMessage(message)
	case "help":
		h.sendHelpMessage(message)
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

func (h *Handler) sendStartMessage(message *tgbotapi.Message) {
	chatID := message.Chat.ID

	// Проверяем, является ли пользователь админом
	isAdmin, err := h.userService.IsAdmin(chatID)
	if err != nil {
		logrus.Infof("Error checking admin status: %v", err)
	}

	text := `👋 Привет! Я бот для управления профилями на Go!

📋 Основные команды:`

	// Общие команды для всех
	text += `
/createprofile - Создать профиль
/myprofile - Показать мой профиль
/updateprofile - Обновить профиль
/deleteprofile - Удалить профиль
/help - Показать все команды`

	// Команды только для админов
	if isAdmin {
		text += `

👑 Команды администратора:
/allusers - Показать всех пользователей
/stats - Статистика бота
/admins - Показать администраторов
/promote [ID] - Назначить администратора
/demote [ID] - Снять администратора
/setrole [ID] [role] - Изменить роль (admin/client)`
	}

	text += `

💡 Примечание: Просто отправьте любое сообщение, и я его повторю!`

	msg := tgbotapi.NewMessage(chatID, text)
	h.client.Bot.Send(msg)
}

func (h *Handler) sendHelpMessage(message *tgbotapi.Message) {
	chatID := message.Chat.ID

	// Проверяем, является ли пользователь админом
	isAdmin, err := h.userService.IsAdmin(chatID)
	if err != nil {
		logrus.Infof("Error checking admin status: %v", err)
	}

	text := `📋 Доступные команды:

👤 Профиль:
/createprofile - Создать профиль (ФИО)
/myprofile - Показать мой профиль
/updateprofile - Обновить профиль
/deleteprofile - Удалить профиль`

	// Команды только для админов
	if isAdmin {
		text += `

👑 Администрирование:
/allusers - Показать всех пользователей
/stats - Статистика бота
/admins - Показать администраторов
/promote [ID] - Назначить администратора
/demote [ID] - Снять администратора
/setrole [ID] [role] - Изменить роль`

		// Показываем ID админа из конфига
		if h.config.BaseAdminChatID != 0 {
			text += fmt.Sprintf("\n\n🔧 ID главного администратора: %d", h.config.BaseAdminChatID)
		}
	}

	text += `

🛠 Утилиты:
/start - Начать работу с ботом
/help - Показать это сообщение
/echo [текст] - Отправить эхо с текстом

💡 Примечание: Просто отправьте любое сообщение, и я его повторю!`

	msg := tgbotapi.NewMessage(chatID, text)
	h.client.Bot.Send(msg)
}