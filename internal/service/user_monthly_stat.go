package service

import (
	"fmt"
	"strings"
	"time"
	"work-schedule-bot/internal/models"
	"work-schedule-bot/internal/repository"

	"github.com/sirupsen/logrus"
)

type UserMonthlyStatService struct {
	statRepo repository.UserMonthlyStatRepository
	userRepo repository.UserRepository
	logger   *logrus.Logger
}

func NewUserMonthlyStatService(
	statRepo repository.UserMonthlyStatRepository,
	userRepo repository.UserRepository,
) *UserMonthlyStatService {
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	return &UserMonthlyStatService{
		statRepo: statRepo,
		userRepo: userRepo,
		logger:   logger,
	}
}

// CreateStatsForNewUser создает статистику для нового пользователя на основе существующих графиков
func (s *UserMonthlyStatService) CreateStatsForNewUser(userID uint, schedules []*models.WorkSchedule) error {
	s.logger.WithField("user_id", userID).Info("Creating monthly stats for new user")

	for _, schedule := range schedules {
		stat := &models.UserMonthlyStat{
			UserID:         userID,
			Year:           schedule.Year,
			Month:          schedule.Month,
			PlannedDays:    schedule.WorkDays,
			PlannedMinutes: schedule.TotalMinutes,
		}
		stat.CalculateStats()

		if err := s.statRepo.Create(stat); err != nil {
			s.logger.WithError(err).Error("Failed to create monthly stat for new user")
			return err
		}
	}

	s.logger.WithFields(logrus.Fields{
		"user_id":   userID,
		"schedules": len(schedules),
	}).Info("Monthly stats created for new user")

	return nil
}

// UpdateStatsForWorkSchedule обновляет статистику всех пользователей при изменении графика
func (s *UserMonthlyStatService) UpdateStatsForWorkSchedule(schedule *models.WorkSchedule) error {
	s.logger.WithFields(logrus.Fields{
		"year":  schedule.Year,
		"month": schedule.Month,
	}).Info("Updating monthly stats for work schedule update")

	// Обновляем статистику для всех пользователей за этот месяц
	err := s.statRepo.UpdateForAllUsers(
		schedule.Year,
		schedule.Month,
		schedule.WorkDays,
		schedule.TotalMinutes,
	)

	if err != nil {
		s.logger.WithError(err).Error("Failed to update monthly stats for work schedule")
		return err
	}

	s.logger.WithFields(logrus.Fields{
		"year":  schedule.Year,
		"month": schedule.Month,
	}).Info("Monthly stats updated for work schedule")

	return nil
}

// GetUserStats возвращает статистику пользователя
func (s *UserMonthlyStatService) GetUserStats(userID uint) ([]*models.UserMonthlyStat, error) {
	s.logger.WithField("user_id", userID).Debug("Getting user monthly stats")
	return s.statRepo.GetByUserID(userID)
}

// GetUserStatByMonth возвращает статистику пользователя за конкретный месяц
func (s *UserMonthlyStatService) GetUserStatByMonth(userID uint, year, month int) (*models.UserMonthlyStat, error) {
	s.logger.WithFields(logrus.Fields{
		"user_id": userID,
		"year":    year,
		"month":   month,
	}).Debug("Getting user monthly stat by month")

	return s.statRepo.GetByUserAndMonth(userID, year, month)
}

// UpdateWorkedTime обновляет отработанное время пользователя
func (s *UserMonthlyStatService) UpdateWorkedTime(userID uint, year, month, workedDays, workedMinutes int) error {
	s.logger.WithFields(logrus.Fields{
		"user_id":        userID,
		"year":           year,
		"month":          month,
		"worked_days":    workedDays,
		"worked_minutes": workedMinutes,
	}).Info("Updating worked time in monthly stats")

	return s.statRepo.UpdateWorkedStats(userID, year, month, workedDays, workedMinutes)
}

// FormatStat форматирует статистику для отображения
func (s *UserMonthlyStatService) FormatStat(stat *models.UserMonthlyStat) string {
	if stat == nil {
		return "❌ Статистика не найдена"
	}

	// Конвертируем минуты в часы:минуты
	plannedHours := stat.PlannedMinutes / 60
	plannedMinutes := stat.PlannedMinutes % 60
	workedHours := stat.WorkedMinutes / 60
	workedMinutes := stat.WorkedMinutes % 60
	overtimeHours := stat.OvertimeMinutes / 60
	overtimeMinutes := stat.OvertimeMinutes % 60
	deficitHours := stat.DeficitMinutes / 60
	deficitMinutes := stat.DeficitMinutes % 60

	monthName := time.Month(stat.Month).String()

	var plannedTime, workedTime, overtimeTime, deficitTime string

	if plannedMinutes == 0 {
		plannedTime = fmt.Sprintf("%dч", plannedHours)
	} else {
		plannedTime = fmt.Sprintf("%dч %dм", plannedHours, plannedMinutes)
	}

	if workedMinutes == 0 {
		workedTime = fmt.Sprintf("%dч", workedHours)
	} else {
		workedTime = fmt.Sprintf("%dч %dм", workedHours, workedMinutes)
	}

	if stat.OvertimeMinutes > 0 {
		if overtimeMinutes == 0 {
			overtimeTime = fmt.Sprintf("%dч", overtimeHours)
		} else {
			overtimeTime = fmt.Sprintf("%dч %dм", overtimeHours, overtimeMinutes)
		}
	}

	if stat.DeficitMinutes > 0 {
		if deficitMinutes == 0 {
			deficitTime = fmt.Sprintf("%dч", deficitHours)
		} else {
			deficitTime = fmt.Sprintf("%dч %dм", deficitHours, deficitMinutes)
		}
	}

	result := fmt.Sprintf(
		`📊 Статистика за %s %d

📅 Плановые показатели:
   📋 Рабочих дней: %d
   ⏰ Плановое время: %s

✅ Фактические показатели:
   📋 Отработано дней: %d
   ⏰ Отработано времени: %s`,
		monthName, stat.Year,
		stat.PlannedDays, plannedTime,
		stat.WorkedDays, workedTime,
	)

	if stat.OvertimeMinutes > 0 {
		result += fmt.Sprintf("\n\n➕ Переработка за месяц: %s", overtimeTime)
	}

	if stat.DeficitMinutes > 0 {
		result += fmt.Sprintf("\n\n➖ Недобор за месяц: %s", deficitTime)
	}

	// Расчет оставшегося времени
	remainingMinutes := stat.PlannedMinutes - stat.WorkedMinutes
	if remainingMinutes > 0 {
		remainingHours := remainingMinutes / 60
		remainingMins := remainingMinutes % 60
		var remainingTime string
		if remainingMins == 0 {
			remainingTime = fmt.Sprintf("%dч", remainingHours)
		} else {
			remainingTime = fmt.Sprintf("%dч %dм", remainingHours, remainingMins)
		}

		remainingDays := stat.PlannedDays - stat.WorkedDays
		result += fmt.Sprintf("\n\n⏳ Осталось отработать: %d дней, %s", remainingDays, remainingTime)

		// Расчет необходимого времени работы в день (ДОБАВЛЕНО)
		if remainingDays > 0 {
			minutesPerDay := remainingMinutes / remainingDays
			
			hoursPerDay := minutesPerDay / 60
			minsPerDay := minutesPerDay % 60
			
			var dailyTime string
			if minsPerDay == 0 {
				dailyTime = fmt.Sprintf("%dч", hoursPerDay)
			} else {
				dailyTime = fmt.Sprintf("%dч %dм", hoursPerDay, minsPerDay)
			}

			result += fmt.Sprintf("\n📈 Необходимо работать в день: %s", dailyTime)
			
			extraStr := "\n\n"
			extraMin := 0
			
			if minutesPerDay < 520 {
				extraStr += "➕ Переработка: "
				extraMin = remainingDays * 520 - remainingMinutes
			} else if minutesPerDay > 520 {
				extraStr += "➖ Недобор: "
				extraMin = remainingMinutes - remainingDays * 520 
			} else {
				extraStr += "✅ План выполняется идеально"
				extraMin = 0
			}
			
			extraHours := extraMin / 60
			extraMinutes := extraMin % 60
			
			var extraTime string
			if extraMinutes == 0 {
				extraTime = fmt.Sprintf("%dч", extraHours)
			} else {
				extraTime = fmt.Sprintf("%dч %dм", extraHours, extraMinutes)
			}
			
			result += fmt.Sprintf("%s%s", extraStr, extraTime)
		}
	}

	// Если уже есть переработка
	if stat.OvertimeMinutes > 0 && remainingMinutes <= 0 {
		overHours := stat.OvertimeMinutes / 60
		overMins := stat.OvertimeMinutes % 60
		var overTime string
		if overMins == 0 {
			overTime = fmt.Sprintf("%dч", overHours)
		} else {
			overTime = fmt.Sprintf("%dч %dм", overHours, overMins)
		}
		result += fmt.Sprintf("\n\n✅ Вы уже выполнили норму месяца! Переработка: %s", overTime)
	}

	result += fmt.Sprintf("\n\n📅 Последнее обновление: %s",
		stat.UpdatedAt.Format("02.01.2006 15:04"))

	return result
}

// FormatStatsList форматирует список статистики
func (s *UserMonthlyStatService) FormatStatsList(stats []*models.UserMonthlyStat) string {
	if len(stats) == 0 {
		return "📭 Статистика пока отсутствует"
	}

	var result strings.Builder
	result.WriteString("📋 Ваша статистика по месяцам:\n\n")

	for i, stat := range stats {

		nowYear := time.Now().Year()
		nowMonth := int(time.Now().Month())
		if nowYear < stat.Year {
			continue
		} else if nowMonth < stat.Month && nowYear == stat.Year {
			continue
		}

		monthName := time.Month(stat.Month).String()

		// Краткий формат
		workedHours := stat.WorkedMinutes / 60
		workedMinutes := stat.WorkedMinutes % 60
		plannedHours := stat.PlannedMinutes / 60
		plannedMinutes := stat.PlannedMinutes % 60

		var workedTime, plannedTime string
		if workedMinutes == 0 {
			workedTime = fmt.Sprintf("%dч", workedHours)
		} else {
			workedTime = fmt.Sprintf("%dч %dм", workedHours, workedMinutes)
		}

		if plannedMinutes == 0 {
			plannedTime = fmt.Sprintf("%dч", plannedHours)
		} else {
			plannedTime = fmt.Sprintf("%dч %dм", plannedHours, plannedMinutes)
		}

		status := "✅"
		if stat.DeficitMinutes > 0 {
			status = "⚠️"
		} else if stat.OvertimeMinutes > 0 {
			status = "➕"
		}

		fmt.Fprintf(&result, "%d. %s %s %d - %d/%d дней, %s/%s %s\n",
			i+1,
			status,
			monthName,
			stat.Year,
			stat.WorkedDays,
			stat.PlannedDays,
			workedTime,
			plannedTime,
			func() string {
				if stat.OvertimeMinutes > 0 {
					overtimeHours := stat.OvertimeMinutes / 60
					overtimeMinutes := stat.OvertimeMinutes % 60
					if overtimeMinutes == 0 {
						return fmt.Sprintf("(+%dч)", overtimeHours)
					}
					return fmt.Sprintf("(+%dч %dм)", overtimeHours, overtimeMinutes)
				} else if stat.DeficitMinutes > 0 {
					deficitHours := stat.DeficitMinutes / 60
					deficitMinutes := stat.DeficitMinutes % 60
					if deficitMinutes == 0 {
						return fmt.Sprintf("(-%dч)", deficitHours)
					}
					return fmt.Sprintf("(-%dч %dм)", deficitHours, deficitMinutes)
				}
				return ""
			}())
	}

	return result.String()
}

// CalculateCompletionPercentage вычисляет процент выполнения
func (s *UserMonthlyStatService) CalculateCompletionPercentage(stat *models.UserMonthlyStat) float64 {
	if stat == nil || stat.PlannedMinutes == 0 {
		return 0
	}

	percentage := (float64(stat.WorkedMinutes) / float64(stat.PlannedMinutes)) * 100
	if percentage > 100 {
		return 100
	}
	return percentage
}

// GetCurrentMonthStat возвращает статистику за текущий месяц
func (s *UserMonthlyStatService) GetCurrentMonthStat(userID uint) (*models.UserMonthlyStat, error) {
	now := time.Now()
	return s.GetUserStatByMonth(userID, now.Year(), int(now.Month()))
}

// CreateStatsForNewSchedule создает статистику для всех пользователей при создании нового графика
func (s *UserMonthlyStatService) CreateStatsForNewSchedule(schedule *models.WorkSchedule) error {
	s.logger.WithFields(logrus.Fields{
		"year":  schedule.Year,
		"month": schedule.Month,
	}).Info("Creating monthly stats for all users for new schedule")

	return s.statRepo.CreateForAllUsers(
		schedule.Year,
		schedule.Month,
		schedule.WorkDays,
		schedule.TotalMinutes,
	)
}

// UpdateStatsForSchedule обновляет статистику всех пользователей при изменении графика
func (s *UserMonthlyStatService) UpdateStatsForSchedule(schedule *models.WorkSchedule) error {
	s.logger.WithFields(logrus.Fields{
		"year":  schedule.Year,
		"month": schedule.Month,
	}).Info("Updating monthly stats for all users for schedule update")

	return s.statRepo.UpdateForAllUsers(
		schedule.Year,
		schedule.Month,
		schedule.WorkDays,
		schedule.TotalMinutes,
	)
}

func (s *UserMonthlyStatService) GetRequiredMinutesByUserID(userID uint, year int, month int) (int, error) {
	stats, err := s.statRepo.GetByUserAndMonth(userID, year, month)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get monthly stats")
		return 0, err
	}
	// if stats == nil{
	// 	s.logger.Info("No monthly stat")
	// 	return 0, nil
	// }

	return stats.DeficitMinutes / (stats.PlannedDays - stats.WorkedDays), nil
}
