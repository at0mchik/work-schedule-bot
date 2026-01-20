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
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		logrus.Fatal("Failed to connect to database:", err)
	}

	// Включаем поддержку внешних ключей для SQLite
	sqlDB, err := db.DB()
	if err != nil {
		logrus.Fatal("Failed to get database instance:", err)
	}

	_, err = sqlDB.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		logrus.Infof("Warning: Failed to enable foreign keys: %v", err)
	}

	// Создаем репозитории
	userRepo, err := repository.NewUserRepository(db)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to create user repository")
	}

	workScheduleRepo, err := repository.NewGormWorkScheduleRepository(db)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to create work schedule repository")
	}

	userMonthlyStatRepo, err := repository.NewGormUserMonthlyStatRepository(db)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to create user monthly stat repository")
	}

	workSessionRepo, err := repository.NewGormWorkSessionRepository(db)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to create work session repository")
	}

	nonWorkingDayRepo, err := repository.NewGormNonWorkingDayRepository(db)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to create non-working day repository")
	}

	absencePeriodRepo, err := repository.NewGormAbsencePeriodRepository(db)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to create absence period repository")
	}

	// Создаем сервисы
	nonWorkingDayService := service.NewNonWorkingDayService(nonWorkingDayRepo)

	// Загружаем выходные дни из JSON
	logrus.Info("Loading non-working days from JSON...")
	count, err := nonWorkingDayService.LoadFromJSON("jsons/weekends_2026.json")
	if err != nil {
		logrus.WithError(err).Fatal("Failed to load non-working days")
	}
	logrus.Infof("Loaded %d non-working days for 2026", count)

	userMonthlyStatService := service.NewUserMonthlyStatService(userMonthlyStatRepo, userRepo)

	// Создаем WorkScheduleService с зависимостью от NonWorkingDayService
	workScheduleService := service.NewWorkScheduleService(
		workScheduleRepo,
		userMonthlyStatService,
		nonWorkingDayService, // ДОБАВЛЕНО
	)

	absenceService := service.NewAbsenceService( // ДОБАВЛЕНО
		absencePeriodRepo,
		workSessionRepo,
		userRepo,
		workScheduleRepo,
		nonWorkingDayService,
	)

	// Автоматически создаем/обновляем графики на основе выходных дней
	logrus.Info("Generating work schedules from non-working days...")
	generatedSchedules, err := workScheduleService.GenerateSchedulesFromNonWorkingDays(2026, 8*60+40) // 8 часов = 480 минут
	if err != nil {
		logrus.WithError(err).Error("Failed to generate work schedules")
	} else {
		logrus.Infof("Generated %d work schedules for 2026", len(generatedSchedules))

		// Проверяем созданные графики
		totalDays := 0
		for _, schedule := range generatedSchedules {
			totalDays += schedule.WorkDays
			logrus.Infof("Month %02d: %d working days, %d minutes per day",
				schedule.Month, schedule.WorkDays, schedule.WorkMinutesPerDay)
		}
		logrus.Infof("Total working days in 2026: %d", totalDays)
	}

	// Создаем остальные сервисы
	userService := service.NewUserService(userRepo, workScheduleRepo, userMonthlyStatService)
	workSessionService := service.NewWorkSessionService(
		workSessionRepo,
		userMonthlyStatRepo,
		workScheduleRepo,
		absencePeriodRepo,
	)

	// Инициализируем администратора
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

	// Создаем обработчик
	botHandler := handler.NewHandler(
		client,
		userService,
		workScheduleService,
		userMonthlyStatService,
		workSessionService,
		nonWorkingDayService,
		absenceService,
		cfg,
	)

	// Настраиваем канал обновлений
	updates := client.Bot.GetUpdatesChan(client.UpdateConfig)

	// Обработка сигналов
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
