package handler

import (
	"strings"
	"work-schedule-bot/internal/config"
	"work-schedule-bot/internal/service"
	"work-schedule-bot/pkg/telegram"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

type Handler struct {
    client                  *telegram.Client
    userService             *service.UserService
    workScheduleService     *service.WorkScheduleService
    userMonthlyStatService  *service.UserMonthlyStatService
    workSessionService      *service.WorkSessionService // НОВОЕ
    userStates              map[int64]string
    config                  *config.BotConfig
}

func NewHandler(
    client *telegram.Client,
    userService *service.UserService,
    workScheduleService *service.WorkScheduleService,
    userMonthlyStatService *service.UserMonthlyStatService,
    workSessionService *service.WorkSessionService, // НОВОЕ
    cfg *config.BotConfig,
) *Handler {
    return &Handler{
        client:                  client,
        userService:             userService,
        workScheduleService:     workScheduleService,
        userMonthlyStatService:  userMonthlyStatService,
        workSessionService:      workSessionService, // НОВОЕ
        userStates:              make(map[int64]string),
        config:                  cfg,
    }
}

func (h *Handler) HandleUpdates(updates tgbotapi.UpdatesChannel) {
	for update := range updates {
		// Обработка callback query (для inline кнопок)
		if update.CallbackQuery != nil {
			h.handleCallbackQuery(update.CallbackQuery)
			continue
		}

		if update.Message == nil {
			continue
		}

		h.handleMessage(update.Message)
	}
}

// handleCallbackQuery обрабатывает inline кнопки
func (h *Handler) handleCallbackQuery(callback *tgbotapi.CallbackQuery) {
    chatID := callback.Message.Chat.ID
    data := callback.Data

    // Удаляем клавиатуру
    editMsg := tgbotapi.NewEditMessageReplyMarkup(chatID, callback.Message.MessageID, tgbotapi.NewInlineKeyboardMarkup())
    h.client.Bot.Send(editMsg)

    // Обработка callback для графиков
    if strings.HasPrefix(data, "confirm_delete_schedule_") || data == "cancel_delete_schedule" {
        h.handleScheduleCallback(callback)
        return
    }

    // Существующая обработка для профилей
    switch data {
    case "confirm_delete":
        err := h.userService.DeleteUser(chatID)
        if err != nil {
            msg := tgbotapi.NewMessage(chatID, "❌ Ошибка удаления профиля: "+err.Error())
            h.client.Bot.Send(msg)
        } else {
            msg := tgbotapi.NewMessage(chatID, "✅ Ваш профиль успешно удален!")
            h.client.Bot.Send(msg)
        }

    case "cancel_delete":
        msg := tgbotapi.NewMessage(chatID, "❌ Удаление профиля отменено.")
        h.client.Bot.Send(msg)
    }

    // Отвечаем на callback (убираем "часики" у кнопки)
    callbackConfig := tgbotapi.NewCallback(callback.ID, "")
    h.client.Bot.Send(callbackConfig)
}

func (h *Handler) handleMessage(message *tgbotapi.Message) {
	logrus.Infof("[%s] %s", message.From.UserName, message.Text)

	chatID := message.Chat.ID

	// Проверяем, находится ли пользователь в процессе создания/обновления профиля
	if state, exists := h.userStates[chatID]; exists {
		h.handleProfileState(message, state)
		return
	}

	// Обработка команд
	if message.IsCommand() {
		h.handleCommand(message)
		return
	}

	// Эхо-ответ на обычные сообщения
	h.sendEchoMessage(message)
}