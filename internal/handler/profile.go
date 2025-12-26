package handler

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// startProfileCreation начинает процесс создания профиля
func (h *Handler) startProfileCreation(message *tgbotapi.Message) {
	chatID := message.Chat.ID

	// Проверяем, есть ли уже профиль
	user, err := h.userService.GetUser(chatID)
	if err == nil && user != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ У вас уже есть профиль!\nИспользуйте /myprofile чтобы посмотреть его или /updateprofile чтобы изменить.")
		h.client.Bot.Send(msg)
		return
	}

	// Начинаем процесс создания
	h.userStates[chatID] = "awaiting_first_name"

	text := `👤 Создание профиля

Шаг 1 из 3:
✏️ Пожалуйста, отправьте ваше имя:`

	msg := tgbotapi.NewMessage(chatID, text)
	h.client.Bot.Send(msg)
}

// handleProfileState обрабатывает состояния создания/обновления профиля
func (h *Handler) handleProfileState(message *tgbotapi.Message, state string) {
	chatID := message.Chat.ID
	text := message.Text

	if state == "awaiting_first_name" {
		// Сохраняем имя и запрашиваем фамилию
		h.userStates[chatID] = "awaiting_last_name:" + text

		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf(
			`Шаг 2 из 3:
	✅ Имя сохранено: %s
	✏️ Теперь отправьте вашу фамилию (если нет фамилии, отправьте "-"):`,
			text))
		h.client.Bot.Send(msg)
	} else if strings.Contains(state, "awaiting_last_name") {
		// Извлекаем имя из состояния
		firstName := strings.TrimPrefix(state, "awaiting_last_name:")
		lastName := text

		// Обрабатываем случай, когда фамилии нет
		if lastName == "-" {
			lastName = ""
		}

		// Получаем username
		username := ""
		if message.From.UserName != "" {
			username = message.From.UserName
		}

		// Создаем профиль
		user, err := h.userService.CreateUser(chatID, username, firstName, lastName)
		if err != nil {
			// Удаляем состояние в случае ошибки
			delete(h.userStates, chatID)

			msg := tgbotapi.NewMessage(chatID, "❌ Ошибка создания профиля: "+err.Error())
			h.client.Bot.Send(msg)
			return
		}

		// Удаляем состояние после успешного создания
		delete(h.userStates, chatID)

		// Форматируем и отправляем информацию о профиле
		profileInfo := h.userService.FormatUserInfo(user)

		responseText := fmt.Sprintf(`🎉 Профиль успешно создан!
	
	%s
	
	Теперь вы можете использовать команду /myprofile чтобы посмотреть свой профиль в любое время.`,
			profileInfo)

		msg := tgbotapi.NewMessage(chatID, responseText)
		h.client.Bot.Send(msg)
	} else if state == "awaiting_update" {
		// Обработка обновления профиля
		delete(h.userStates, chatID)

		parts := strings.Fields(text)
		if len(parts) < 1 {
			msg := tgbotapi.NewMessage(chatID, "❌ Неверный формат. Пожалуйста, отправьте имя и фамилию.")
			h.client.Bot.Send(msg)
			return
		}

		firstName := parts[0]
		lastName := ""
		if len(parts) > 1 {
			lastName = parts[1]
		}

		username := ""
		if message.From.UserName != "" {
			username = message.From.UserName
		}

		user, err := h.userService.UpdateUser(chatID, username, firstName, lastName)
		if err != nil {
			msg := tgbotapi.NewMessage(chatID, "❌ Ошибка обновления профиля: "+err.Error())
			h.client.Bot.Send(msg)
			return
		}

		profileInfo := h.userService.FormatUserInfo(user)
		responseText := fmt.Sprintf(`✅ Профиль успешно обновлен!
	
	%s`,
			profileInfo)

		msg := tgbotapi.NewMessage(chatID, responseText)
		h.client.Bot.Send(msg)
	}
}

// showProfile показывает профиль пользователя
func (h *Handler) showProfile(message *tgbotapi.Message) {
	chatID := message.Chat.ID

	user, err := h.userService.GetUser(chatID)
	if err != nil || user == nil {
		msg := tgbotapi.NewMessage(chatID, "❌ Профиль не найден.\nИспользуйте /createprofile чтобы создать профиль.")
		h.client.Bot.Send(msg)
		return
	}

	profileInfo := h.userService.FormatUserInfo(user)
	msg := tgbotapi.NewMessage(chatID, profileInfo)
	h.client.Bot.Send(msg)
}

// startProfileUpdate начинает процесс обновления профиля
func (h *Handler) startProfileUpdate(message *tgbotapi.Message) {
	chatID := message.Chat.ID

	user, err := h.userService.GetUser(chatID)
	if err != nil || user == nil {
		msg := tgbotapi.NewMessage(chatID, "❌ Профиль не найден.\nИспользуйте /createprofile чтобы создать профиль.")
		h.client.Bot.Send(msg)
		return
	}

	text := `✏️ Обновление профиля

Отправьте новые данные в формате:
Имя Фамилия

Например: *Иван Иванов*
Или просто: *Иван* (если нужно обновить только имя)`

	msg := tgbotapi.NewMessage(chatID, text)
	h.client.Bot.Send(msg)

	// Устанавливаем состояние для обновления
	h.userStates[chatID] = "awaiting_update"
}

// deleteProfile удаляет профиль пользователя
func (h *Handler) deleteProfile(message *tgbotapi.Message) {
	chatID := message.Chat.ID

	// Создаем inline клавиатуру для подтверждения
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Да, удалить", "confirm_delete"),
			tgbotapi.NewInlineKeyboardButtonData("❌ Нет, отменить", "cancel_delete"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, "⚠️ Вы уверены, что хотите удалить свой профиль?\nЭто действие нельзя отменить.")
	msg.ReplyMarkup = keyboard
	h.client.Bot.Send(msg)
}
