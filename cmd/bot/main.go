package main

import (
	"os"
	"os/signal"
	"syscall"
	"work-schedule-bot/internal/config"
	"work-schedule-bot/internal/handler"
	"work-schedule-bot/internal/repository"
	"work-schedule-bot/internal/service"
	"work-schedule-bot/pkg/telegram"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	logrus.Info("Initializing config...")
	cfg := config.GetBotConfig()
	logrus.Info("Config initialized...")

	// Инициализируем SQLite базу данных
	db, err := gorm.Open(sqlite.Open(cfg.DatabaseURL), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true, // SQLite ограничения
	})
	if err != nil {
		logrus.Fatal("Failed to connect to database:", err)
	}

	// Включаем поддержку внешних ключей для SQLite
	sqlDB, err := db.DB()
	if err != nil {
		logrus.Fatal("Failed to get database instance:", err)
	}

	// Включаем поддержку внешних ключей (требуется для SQLite)
	_, err = sqlDB.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		logrus.Infof("Warning: Failed to enable foreign keys: %v", err)
	}

	userRepo, err := repository.NewGormUserRepository(db)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to create user repository")
	}

	// Создаем репозиторий графиков работы
	workScheduleRepo, err := repository.NewGormWorkScheduleRepository(db)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to create work schedule repository")
	}

	// Создаем репозиторий статистики пользователей
	userMonthlyStatRepo, err := repository.NewGormUserMonthlyStatRepository(db)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to create user monthly stat repository")
	}

	workSessionRepo, err := repository.NewGormWorkSessionRepository(db)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to create work session repository")
	}

	// Создаем сервис статистики пользователей
	userMonthlyStatService := service.NewUserMonthlyStatService(userMonthlyStatRepo, userRepo)

	// Создаем сервис пользователей с зависимостями
	userService := service.NewUserService(userRepo, workScheduleRepo, userMonthlyStatService)

	// Создаем сервис графиков работы с зависимостью от сервиса статистики
	workScheduleService := service.NewWorkScheduleService(workScheduleRepo, userMonthlyStatService)

	// Создаем сервис рабочих сессий
	workSessionService := service.NewWorkSessionService(
		workSessionRepo,
		userMonthlyStatRepo,
		workScheduleRepo,
	)

	// Инициализируем администратора из конфига
	if err := userService.InitializeAdmin(cfg.BaseAdminChatID); err != nil {
		logrus.Infof("Warning: Failed to initialize admin: %v", err)
	} else if cfg.BaseAdminChatID != 0 {
		logrus.Infof("Admin initialized with chat ID: %d", cfg.BaseAdminChatID)
	}

	// Создаем клиент Telegram
	client, err := telegram.NewClient(cfg.TelegramToken)
	if err != nil {
		logrus.Fatal("Failed to create Telegram client:", err)
	}

	logrus.Infof("Authorized on account %s", client.Bot.Self.UserName)

	// Создаем обработчик с конфигом
	botHandler := handler.NewHandler(
		client,
		userService,
		workScheduleService,
		userMonthlyStatService,
		workSessionService, // НОВОЕ
		cfg,
	)

	// Настраиваем канал обновлений
	updates := client.Bot.GetUpdatesChan(client.UpdateConfig)

	// Обработка сигналов для graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// Запускаем обработку сообщений
	go botHandler.HandleUpdates(updates)

	logrus.Info("Bot started. Press Ctrl+C to stop.")
	<-stop

	// Закрываем соединение с БД
	if err := sqlDB.Close(); err != nil {
		logrus.Infof("Error closing database: %v", err)
	}

	logrus.Info("Bot stopped gracefully")
}
