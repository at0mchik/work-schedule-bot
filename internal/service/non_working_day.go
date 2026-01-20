package service

import (
	"time"
	"work-schedule-bot/internal/models"
	"work-schedule-bot/internal/repository"
	"work-schedule-bot/pkg/weekends"

	"github.com/sirupsen/logrus"
)

type NonWorkingDayService struct {
	repo repository.NonWorkingDayRepository
}

func NewNonWorkingDayService(repo repository.NonWorkingDayRepository) *NonWorkingDayService {
	return &NonWorkingDayService{repo: repo}
}

// LoadFromJSON загружает выходные дни из JSON файла в базу данных
func (s *NonWorkingDayService) LoadFromJSON(filePath string) (int, error) {
	// Парсим JSON
	weekendDays, err := weekends.ParseWeekendsJSON(filePath)
	if err != nil {
		return 0, err
	}

	// Преобразуем в модели
	var nonWorkingDays []models.NonWorkingDay
	for _, wd := range weekendDays {
		nonWorkingDays = append(nonWorkingDays, models.NonWorkingDay{
			Date:  wd.Date,
			Year:  wd.Year,
			Month: wd.Month,
			Day:   wd.Day,
		})
	}

	// Удаляем старые записи (чтобы избежать дублирования)
	if err := s.repo.DeleteAll(); err != nil {
		logrus.Warnf("Failed to delete old non-working days: %v", err)
	}

	// Сохраняем в базу
	if err := s.repo.BulkCreate(nonWorkingDays); err != nil {
		return 0, err
	}

	return len(nonWorkingDays), nil
}

// GetNonWorkingDays возвращает все выходные дни
func (s *NonWorkingDayService) GetNonWorkingDays() ([]models.NonWorkingDay, error) {
	return s.repo.GetAll()
}

// GetNonWorkingDaysForMonth возвращает выходные дни для указанного месяца
func (s *NonWorkingDayService) GetNonWorkingDaysForMonth(year, month int) ([]models.NonWorkingDay, error) {
	return s.repo.GetByYearMonth(year, month)
}

// IsNonWorkingDay проверяет, является ли дата выходным днем
func (s *NonWorkingDayService) IsNonWorkingDay(date time.Time) (bool, error) {
	return s.repo.IsNonWorkingDay(date)
}

// CountNonWorkingDays возвращает количество выходных дней
func (s *NonWorkingDayService) CountNonWorkingDays() (int64, error) {
	days, err := s.repo.GetAll()
	if err != nil {
		return 0, err
	}
	return int64(len(days)), nil
}