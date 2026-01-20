package repository

import (
	"time"
	"work-schedule-bot/internal/models"

	"gorm.io/gorm"
)

type NonWorkingDayRepository interface {
	Create(day *models.NonWorkingDay) error
	GetByDate(date time.Time) (*models.NonWorkingDay, error)
	GetByYearMonth(year, month int) ([]models.NonWorkingDay, error)
	GetAll() ([]models.NonWorkingDay, error)
	BulkCreate(days []models.NonWorkingDay) error
	DeleteAll() error
	IsNonWorkingDay(date time.Time) (bool, error)
}

type GormNonWorkingDayRepository struct {
	db *gorm.DB
}

func NewGormNonWorkingDayRepository(db *gorm.DB) (NonWorkingDayRepository, error) {
	// Автомиграция для таблицы non_working_days
	if err := db.AutoMigrate(&models.NonWorkingDay{}); err != nil {
		return nil, err
	}

	return &GormNonWorkingDayRepository{db: db}, nil
}

func (r *GormNonWorkingDayRepository) Create(day *models.NonWorkingDay) error {
	return r.db.Create(day).Error
}

func (r *GormNonWorkingDayRepository) BulkCreate(days []models.NonWorkingDay) error {
	if len(days) == 0 {
		return nil
	}
	return r.db.Create(&days).Error
}

func (r *GormNonWorkingDayRepository) GetByDate(date time.Time) (*models.NonWorkingDay, error) {
	var day models.NonWorkingDay
	err := r.db.Where("date = ?", date.Format("2006-01-02")).First(&day).Error
	if err != nil {
		return nil, err
	}
	return &day, nil
}

func (r *GormNonWorkingDayRepository) GetByYearMonth(year, month int) ([]models.NonWorkingDay, error) {
	var days []models.NonWorkingDay
	err := r.db.Where("year = ? AND month = ?", year, month).Find(&days).Error
	return days, err
}

func (r *GormNonWorkingDayRepository) GetAll() ([]models.NonWorkingDay, error) {
	var days []models.NonWorkingDay
	err := r.db.Find(&days).Error
	return days, err
}

func (r *GormNonWorkingDayRepository) DeleteAll() error {
	return r.db.Exec("DELETE FROM non_working_days").Error
}

func (r *GormNonWorkingDayRepository) IsNonWorkingDay(date time.Time) (bool, error) {
	var count int64
	err := r.db.Model(&models.NonWorkingDay{}).
		Where("date = ?", date.Format("2006-01-02")).
		Count(&count).Error
	return count > 0, err
}