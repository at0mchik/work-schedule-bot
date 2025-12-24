package handler

import (
	"fmt"
	"strconv"
	"strings"
	"work-schedule-bot/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// showAllUsers показывает всех пользователей
func (h *Handler) showAllUsers(message *tgbotapi.Message) {
	chatID := message.Chat.ID

	// Проверяем права доступа
	isAdmin, err := h.userService.IsAdmin(chatID)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка проверки прав доступа: "+err.Error())
		h.client.Bot.Send(msg)
		return
	}

	if !isAdmin {
		msg := tgbotapi.NewMessage(chatID, "❌ Доступ запрещен. Эта команда только для администраторов.")
		h.client.Bot.Send(msg)
		return
	}

	allUsers, err := h.userService.FormatAllUsers()
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка получения списка пользователей: "+err.Error())
		h.client.Bot.Send(msg)
		return
	}

	msg := tgbotapi.NewMessage(chatID, allUsers)
	h.client.Bot.Send(msg)
}

// showStats показывает статистику (только для админов)
func (h *Handler) showStats(message *tgbotapi.Message) {
	chatID := message.Chat.ID

	// Проверяем права доступа
	isAdmin, err := h.userService.IsAdmin(chatID)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка проверки прав доступа: "+err.Error())
		h.client.Bot.Send(msg)
		return
	}

	if !isAdmin {
		msg := tgbotapi.NewMessage(chatID, "❌ Доступ запрещен. Эта команда только для администраторов.")
		h.client.Bot.Send(msg)
		return
	}

	total, admins, err := h.userService.GetStats()
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка получения статистики: "+err.Error())
		h.client.Bot.Send(msg)
		return
	}

	clients := total - admins

	text := fmt.Sprintf(`📊 Статистика бота:

👥 Всего пользователей: %d
👑 Администраторов: %d
👤 Клиентов: %d
💾 Хранилище: В памяти (in-memory)
🔄 Данные сохраняются до перезапуска бота

⚠️ Внимание: При перезапуске бота все данные будут потеряны!`,
		total, admins, clients)

	msg := tgbotapi.NewMessage(chatID, text)
	h.client.Bot.Send(msg)
}

// showAdmins показывает всех администраторов (только для админов)
func (h *Handler) showAdmins(message *tgbotapi.Message) {
	chatID := message.Chat.ID

	// Проверяем права доступа
	isAdmin, err := h.userService.IsAdmin(chatID)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка проверки прав доступа: "+err.Error())
		h.client.Bot.Send(msg)
		return
	}

	if !isAdmin {
		msg := tgbotapi.NewMessage(chatID, "❌ Доступ запрещен. Эта команда только для администраторов.")
		h.client.Bot.Send(msg)
		return
	}

	admins, err := h.userService.GetAdmins()
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка получения списка администраторов: "+err.Error())
		h.client.Bot.Send(msg)
		return
	}

	if len(admins) == 0 {
		msg := tgbotapi.NewMessage(chatID, "👑 Список администраторов пуст.")
		h.client.Bot.Send(msg)
		return
	}

	var lines []string
	lines = append(lines, "👑 Администраторы:")
	lines = append(lines, "")

	for i, admin := range admins {
		adminInfo := fmt.Sprintf("%d. ", i+1)
		if admin.FirstName != "" {
			adminInfo += admin.FirstName + " "
		}
		if admin.LastName != "" {
			adminInfo += admin.LastName + " "
		}
		if admin.Username != "" {
			adminInfo += fmt.Sprintf("(@%s) ", admin.Username)
		}
		adminInfo += fmt.Sprintf("- ID: %d", admin.ChatID)
		lines = append(lines, adminInfo)
	}

	msg := tgbotapi.NewMessage(chatID, strings.Join(lines, "\n"))
	h.client.Bot.Send(msg)
}

// promoteToAdmin назначает пользователя администратором
func (h *Handler) promoteToAdmin(message *tgbotapi.Message, args string) {
	chatID := message.Chat.ID

	// Проверяем права доступа
	isAdmin, err := h.userService.IsAdmin(chatID)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка проверки прав доступа: "+err.Error())
		h.client.Bot.Send(msg)
		return
	}

	if !isAdmin {
		msg := tgbotapi.NewMessage(chatID, "❌ Доступ запрещен. Эта команда только для администраторов.")
		h.client.Bot.Send(msg)
		return
	}

	if args == "" {
		msg := tgbotapi.NewMessage(chatID, "❌ Укажите ID пользователя.\nПример: /promote 123456789")
		h.client.Bot.Send(msg)
		return
	}

	targetChatID, err := strconv.ParseInt(args, 10, 64)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ Неверный формат ID.\nID должен быть числом.")
		h.client.Bot.Send(msg)
		return
	}

	err = h.userService.UpdateRole(chatID, targetChatID, "admin")
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка назначения администратора: "+err.Error())
		h.client.Bot.Send(msg)
		return
	}

	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("✅ Пользователь с ID %d теперь администратор!", targetChatID))
	h.client.Bot.Send(msg)
}

// demoteToClient снимает пользователя с должности администратора
func (h *Handler) demoteToClient(message *tgbotapi.Message, args string) {
	chatID := message.Chat.ID

	// Проверяем права доступа
	isAdmin, err := h.userService.IsAdmin(chatID)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка проверки прав доступа: "+err.Error())
		h.client.Bot.Send(msg)
		return
	}

	if !isAdmin {
		msg := tgbotapi.NewMessage(chatID, "❌ Доступ запрещен. Эта команда только для администраторов.")
		h.client.Bot.Send(msg)
		return
	}

	if args == "" {
		msg := tgbotapi.NewMessage(chatID, "❌ Укажите ID пользователя.\nПример: /demote 123456789")
		h.client.Bot.Send(msg)
		return
	}

	targetChatID, err := strconv.ParseInt(args, 10, 64)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ Неверный формат ID.\nID должен быть числом.")
		h.client.Bot.Send(msg)
		return
	}

	// Не позволяем снять главного администратора из конфига
	if targetChatID == h.config.BaseAdminChatID && h.config.BaseAdminChatID != 0 {
		msg := tgbotapi.NewMessage(chatID, "❌ Нельзя снять главного администратора, заданного в конфигурации!")
		h.client.Bot.Send(msg)
		return
	}

	err = h.userService.UpdateRole(chatID, targetChatID, "client")
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка снятия администратора: "+err.Error())
		h.client.Bot.Send(msg)
		return
	}

	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("✅ Пользователь с ID %d теперь клиент!", targetChatID))
	h.client.Bot.Send(msg)
}

// setUserRole изменяет роль пользователя
func (h *Handler) setUserRole(message *tgbotapi.Message, args string) {
	chatID := message.Chat.ID

	// Проверяем права доступа
	isAdmin, err := h.userService.IsAdmin(chatID)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка проверки прав доступа: "+err.Error())
		h.client.Bot.Send(msg)
		return
	}

	if !isAdmin {
		msg := tgbotapi.NewMessage(chatID, "❌ Доступ запрещен. Эта команда только для администраторов.")
		h.client.Bot.Send(msg)
		return
	}

	parts := strings.Fields(args)
	if len(parts) != 2 {
		msg := tgbotapi.NewMessage(chatID, "❌ Неверный формат.\nПример: /setrole 123456789 admin\nДоступные роли: admin, client")
		h.client.Bot.Send(msg)
		return
	}

	targetChatID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ Неверный формат ID.\nID должен быть числом.")
		h.client.Bot.Send(msg)
		return
	}

	roleStr := strings.ToLower(parts[1])
	var role models.Role

	switch roleStr {
	case "admin":
		role = "admin"
	case "client":
		// Не позволяем изменить роль главного администратора на client
		if targetChatID == h.config.BaseAdminChatID && h.config.BaseAdminChatID != 0 {
			msg := tgbotapi.NewMessage(chatID, "❌ Нельзя изменить роль главного администратора, заданного в конфигурации!")
			h.client.Bot.Send(msg)
			return
		}
		role = "client"
	default:
		msg := tgbotapi.NewMessage(chatID, "❌ Неизвестная роль.\nДоступные роли: admin, client")
		h.client.Bot.Send(msg)
		return
	}

	err = h.userService.UpdateRole(chatID, targetChatID, role)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка изменения роли: "+err.Error())
		h.client.Bot.Send(msg)
		return
	}

	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("✅ Роль пользователя с ID %d изменена на '%s'!", targetChatID, role))
	h.client.Bot.Send(msg)
}
